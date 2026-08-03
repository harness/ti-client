package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func newUploadITCgClient(t *testing.T, handler http.HandlerFunc) (*HTTPClient, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return &HTTPClient{
		Client:    server.Client(),
		Endpoint:  server.URL,
		AccountID: "acc",
		Token:     "tok",
	}, server
}

func isOpenRetry(c *HTTPClient, url string, payload []byte, maxElapsed time.Duration, retryOn5xx bool) (*http.Response, error) {
	return c.retry(context.Background(), url, "POST", "", payload, nil, true, retryOn5xx, createBackoff(maxElapsed))
}

// TestUploadITCg_BodyReplayedOnRetry guards the reader-exhaustion bug fixed for
// isOpen retries: each attempt must POST the full payload, not an empty body.
func TestUploadITCg_BodyReplayedOnRetry(t *testing.T) {
	var attempts int32
	var lastBody []byte
	payload := []byte{0x1f, 0x8b, 0x08, 0x00, 0xde, 0xad, 0xbe, 0xef} // fake gzip header+bytes

	c, _ := newUploadITCgClient(t, func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		b, _ := io.ReadAll(r.Body)
		lastBody = append([]byte(nil), b...)
		if n < 2 {
			http.Error(w, "warming up", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	if err := c.UploadITCg(context.Background(), payload, "cg-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := atomic.LoadInt32(&attempts); got != 2 {
		t.Errorf("expected 2 attempts, got %d", got)
	}
	if !bytes.Equal(lastBody, payload) {
		t.Errorf("retry sent body %v, want %v (reader was likely exhausted)", lastBody, payload)
	}
}

// TestUploadITCg_AllAttemptBodiesIdentical records every attempt body so a
// partial-read / truncated replay cannot hide behind checking only the last one.
func TestUploadITCg_AllAttemptBodiesIdentical(t *testing.T) {
	var attempts int32
	var mu sync.Mutex
	bodies := make([][]byte, 0, 3)
	payload := bytes.Repeat([]byte{0xab, 0xcd}, 2048) // large enough to catch truncation

	c, server := newUploadITCgClient(t, func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		b, _ := io.ReadAll(r.Body)
		mu.Lock()
		bodies = append(bodies, append([]byte(nil), b...))
		mu.Unlock()
		if n < 3 {
			http.Error(w, "retry", http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	url := server.URL + "/it/uploadcg?accountId=acc&cgId=cg-large"
	if _, err := isOpenRetry(c, url, payload, 30*time.Second, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(bodies) != 3 {
		t.Fatalf("expected 3 recorded bodies, got %d", len(bodies))
	}
	for i, b := range bodies {
		if !bytes.Equal(b, payload) {
			t.Errorf("attempt %d body len=%d mismatch (want %d)", i+1, len(b), len(payload))
		}
	}
}

// TestUploadITCg_5xxExhaustionReturnsError verifies that when every attempt
// returns 5xx and the backoff budget is exhausted, the isOpen path returns a
// non-nil error (never nil,nil / false success).
func TestUploadITCg_5xxExhaustionReturnsError(t *testing.T) {
	var attempts int32

	c, server := newUploadITCgClient(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		http.Error(w, "down", http.StatusServiceUnavailable)
	})

	payload := []byte("gzipped-callgraph")
	path := server.URL + "/it/uploadcg?accountId=acc&cgId=cg-1"
	res, err := isOpenRetry(c, path, payload, 50*time.Millisecond, true)
	if res != nil {
		t.Errorf("expected nil response after closed exhaustion path, got %#v", res)
	}
	if err == nil {
		t.Fatal("expected non-nil error on 5xx exhaustion, got nil (false success)")
	}
	var clientErr *Error
	if !errors.As(err, &clientErr) {
		t.Fatalf("expected *client.Error, got %T: %v", err, err)
	}
	if clientErr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, clientErr.Code)
	}
	if got := atomic.LoadInt32(&attempts); got < 1 {
		t.Errorf("expected at least 1 attempt, got %d", got)
	}
}

func TestUploadITCg_NonRetriableClientErrors(t *testing.T) {
	cases := []struct {
		name   string
		status int
	}{
		{"bad_request", http.StatusBadRequest},
		{"unauthorized", http.StatusUnauthorized},
		{"forbidden", http.StatusForbidden},
		{"not_found", http.StatusNotFound},
		{"conflict", http.StatusConflict},
		{"unprocessable", http.StatusUnprocessableEntity},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var attempts int32
			c, _ := newUploadITCgClient(t, func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&attempts, 1)
				http.Error(w, tc.name, tc.status)
			})

			err := c.UploadITCg(context.Background(), []byte("payload"), "cg-1")
			if err == nil {
				t.Fatalf("expected error on %d, got nil", tc.status)
			}
			var clientErr *Error
			if !errors.As(err, &clientErr) {
				t.Fatalf("expected *client.Error, got %T: %v", err, err)
			}
			if clientErr.Code != tc.status {
				t.Errorf("expected status %d, got %d", tc.status, clientErr.Code)
			}
			if got := atomic.LoadInt32(&attempts); got != 1 {
				t.Errorf("4xx must not retry: expected 1 attempt, got %d", got)
			}
		})
	}
}

func TestUploadITCg_RetriableServerStatuses(t *testing.T) {
	statuses := []int{
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
	}

	for _, status := range statuses {
		t.Run(fmt.Sprintf("status_%d", status), func(t *testing.T) {
			var attempts int32
			c, server := newUploadITCgClient(t, func(w http.ResponseWriter, r *http.Request) {
				n := atomic.AddInt32(&attempts, 1)
				if n == 1 {
					http.Error(w, "transient", status)
					return
				}
				w.WriteHeader(http.StatusOK)
			})

			url := server.URL + "/it/uploadcg?accountId=acc&cgId=cg-1"
			if _, err := isOpenRetry(c, url, []byte("p"), 30*time.Second, true); err != nil {
				t.Fatalf("expected eventual success after %d, got %v", status, err)
			}
			if got := atomic.LoadInt32(&attempts); got != 2 {
				t.Errorf("expected 2 attempts for status %d, got %d", status, got)
			}
		})
	}
}

// TestUploadITCg_5xxNoRetryWhenDisabled mirrors ForwardWithRetry's
// RetryOnServerErrors=false contract for non-idempotent-sensitive callers.
func TestUploadITCg_5xxNoRetryWhenDisabled(t *testing.T) {
	var attempts int32
	c, server := newUploadITCgClient(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		http.Error(w, "boom", http.StatusInternalServerError)
	})

	url := server.URL + "/it/uploadcg?accountId=acc&cgId=cg-1"
	_, err := isOpenRetry(c, url, []byte("p"), 5*time.Second, false)
	if err == nil {
		t.Fatal("expected error on 500 with retry disabled")
	}
	if got := atomic.LoadInt32(&attempts); got != 1 {
		t.Errorf("expected exactly 1 attempt, got %d", got)
	}
}

func TestUploadITCg_SuccessStatuses(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
	}{
		{"ok", http.StatusOK, `{"status":"ok","id":"cg-1"}`},
		{"ok_already_exists", http.StatusOK, `{"status":"ok","already_exists":true,"id":"cg-1"}`},
		{"created", http.StatusCreated, `{"status":"created"}`},
		{"no_content", http.StatusNoContent, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, _ := newUploadITCgClient(t, func(w http.ResponseWriter, r *http.Request) {
				if tc.body == "" {
					w.WriteHeader(tc.status)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			})
			if err := c.UploadITCg(context.Background(), []byte("gzip"), "cg-1"); err != nil {
				t.Fatalf("expected success for %s (%d), got %v", tc.name, tc.status, err)
			}
		})
	}
}

func TestUploadITCg_EmptyPayload(t *testing.T) {
	var gotLen int
	c, _ := newUploadITCgClient(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotLen = len(b)
		w.WriteHeader(http.StatusOK)
	})
	if err := c.UploadITCg(context.Background(), nil, "cg-empty"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotLen != 0 {
		t.Errorf("expected empty body, got %d bytes", gotLen)
	}
}

func TestUploadITCg_RequestContractMatchesHarnessTI(t *testing.T) {
	// Mirrors harness-ti HandleUploadCg: POST /it/uploadcg?accountId=&cgId=
	// with auth header and raw body (pre-gzipped by caller).
	var (
		gotMethod    string
		gotPath      string
		gotAccountID string
		gotCgID      string
		gotToken     string
		gotAPIKey    string
		gotBody      []byte
	)

	c, _ := newUploadITCgClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAccountID = r.URL.Query().Get("accountId")
		gotCgID = r.URL.Query().Get("cgId")
		gotToken = r.Header.Get("X-Harness-Token")
		gotAPIKey = r.Header.Get("x-api-key")
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})

	payload := []byte{0x1f, 0x8b, 0x08}
	if err := c.UploadITCg(context.Background(), payload, "cg-dedup-key"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method: got %q want POST", gotMethod)
	}
	if gotPath != "/it/uploadcg" {
		t.Errorf("path: got %q want /it/uploadcg", gotPath)
	}
	if gotAccountID != "acc" {
		t.Errorf("accountId: got %q want acc", gotAccountID)
	}
	if gotCgID != "cg-dedup-key" {
		t.Errorf("cgId: got %q want cg-dedup-key", gotCgID)
	}
	if gotToken != "tok" {
		t.Errorf("X-Harness-Token: got %q want tok", gotToken)
	}
	if gotAPIKey != "" {
		t.Errorf("x-api-key should be empty for non-PAT token, got %q", gotAPIKey)
	}
	if !bytes.Equal(gotBody, payload) {
		t.Errorf("body mismatch: got %v want %v", gotBody, payload)
	}
}

func TestUploadITCg_PATUsesAPIKeyHeader(t *testing.T) {
	var gotAPIKey, gotToken string
	c, _ := newUploadITCgClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotAPIKey = r.Header.Get("x-api-key")
		gotToken = r.Header.Get("X-Harness-Token")
		w.WriteHeader(http.StatusOK)
	})
	c.Token = "pat.abcdef"

	if err := c.UploadITCg(context.Background(), []byte("p"), "cg-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotAPIKey != "pat.abcdef" {
		t.Errorf("x-api-key: got %q want pat.abcdef", gotAPIKey)
	}
	if gotToken != "" {
		t.Errorf("X-Harness-Token should be empty for PAT, got %q", gotToken)
	}
}

func TestUploadITCg_ValidationErrors(t *testing.T) {
	cases := []struct {
		name     string
		client   *HTTPClient
		wantSubstr string
	}{
		{
			name:       "missing_endpoint",
			client:     &HTTPClient{Token: "tok", AccountID: "acc"},
			wantSubstr: "endpoint",
		},
		{
			name:       "missing_token",
			client:     &HTTPClient{Endpoint: "http://example", AccountID: "acc"},
			wantSubstr: "token",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.client.UploadITCg(context.Background(), []byte("p"), "cg-1")
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if !strings.Contains(strings.ToLower(err.Error()), tc.wantSubstr) {
				t.Errorf("error %q should mention %q", err.Error(), tc.wantSubstr)
			}
		})
	}
}

func TestUploadITCg_TransportErrorThenSuccess(t *testing.T) {
	var attempts int32
	payload := []byte("after-transport-blip")
	var lastBody []byte

	c, server := newUploadITCgClient(t, func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		b, _ := io.ReadAll(r.Body)
		lastBody = append([]byte(nil), b...)
		if n == 1 {
			hj, ok := w.(http.Hijacker)
			if !ok {
				t.Fatal("server does not support hijacking")
			}
			conn, _, err := hj.Hijack()
			if err != nil {
				t.Fatalf("hijack failed: %v", err)
			}
			conn.Close()
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	url := server.URL + "/it/uploadcg?accountId=acc&cgId=cg-1"
	if _, err := isOpenRetry(c, url, payload, 30*time.Second, true); err != nil {
		t.Fatalf("expected success after transport retry, got %v", err)
	}
	if got := atomic.LoadInt32(&attempts); got < 2 {
		t.Errorf("expected at least 2 attempts, got %d", got)
	}
	if !bytes.Equal(lastBody, payload) {
		t.Errorf("body after transport retry mismatch")
	}
}

func TestUploadITCg_TransportExhaustionReturnsError(t *testing.T) {
	c, server := newUploadITCgClient(t, func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Fatal("server does not support hijacking")
		}
		conn, _, err := hj.Hijack()
		if err != nil {
			t.Fatalf("hijack failed: %v", err)
		}
		conn.Close()
	})

	url := server.URL + "/it/uploadcg?accountId=acc&cgId=cg-1"
	res, err := isOpenRetry(c, url, []byte("p"), 80*time.Millisecond, true)
	if res != nil {
		t.Errorf("expected nil response on transport exhaustion, got %#v", res)
	}
	if err == nil {
		t.Fatal("expected transport error on exhaustion, got nil")
	}
}

func TestUploadITCg_ContextCancellation(t *testing.T) {
	var attempts int32
	c, server := newUploadITCgClient(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		http.Error(w, "down", http.StatusServiceUnavailable)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	url := server.URL + "/it/uploadcg?accountId=acc&cgId=cg-1"
	_, err := c.retry(ctx, url, "POST", "", []byte("p"), nil, true, true, createBackoff(30*time.Second))
	if err == nil {
		t.Fatal("expected context error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) && ctx.Err() == nil {
		// Accept any error once the context is done; backoff sleep may surface
		// the last request error instead of ctx.Err() depending on timing.
		if ctx.Err() == nil {
			t.Fatalf("expected context to be done, err=%v", err)
		}
	}
	if got := atomic.LoadInt32(&attempts); got > 8 {
		t.Errorf("expected few attempts before cancel, got %d", got)
	}
}

func TestUploadITCg_IsOpenRequiresByteSlice(t *testing.T) {
	c, server := newUploadITCgClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be reached for bad input type")
	})

	url := server.URL + "/it/uploadcg?accountId=acc&cgId=cg-1"
	_, err := c.retry(context.Background(), url, "POST", "", bytes.NewReader([]byte("nope")), nil, true, true,
		createBackoff(time.Second))
	if err == nil {
		t.Fatal("expected type error for non-[]byte isOpen input")
	}
	if !strings.Contains(err.Error(), "[]byte") {
		t.Errorf("error should mention []byte requirement, got %q", err.Error())
	}
}

// TestUploadITCg_ConcurrentSharedClient exercises the shared HTTPClient under
// the race detector — concurrent UploadITCg calls must not race on client fields.
func TestUploadITCg_ConcurrentSharedClient(t *testing.T) {
	var hits int32
	c, _ := newUploadITCgClient(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		b, _ := io.ReadAll(r.Body)
		if len(b) == 0 {
			http.Error(w, "empty body", http.StatusBadRequest)
			return
		}
		// Simulate harness-ti dedup path for even cgIds.
		if strings.HasSuffix(r.URL.Query().Get("cgId"), "even") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok","already_exists":true}`))
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	const goroutines = 16
	var wg sync.WaitGroup
	errCh := make(chan error, goroutines)
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cgID := fmt.Sprintf("cg-%d", i)
			if i%2 == 0 {
				cgID += "-even"
			}
			payload := []byte(fmt.Sprintf("payload-%d", i))
			if err := c.UploadITCg(context.Background(), payload, cgID); err != nil {
				errCh <- err
			}
		}(i)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Errorf("concurrent UploadITCg failed: %v", err)
	}
	if got := atomic.LoadInt32(&hits); got != goroutines {
		t.Errorf("expected %d hits, got %d", goroutines, got)
	}
}

func TestUploadITCg_MissingAccountIDQueryStillHitsServer(t *testing.T) {
	// validateTiArgs does not require AccountID today; document that the client
	// still POSTs (harness-ti would return 400 for missing accountId).
	var gotAccountID string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAccountID = r.URL.Query().Get("accountId")
		http.Error(w, "missing required param: accountId", http.StatusBadRequest)
	}))
	t.Cleanup(server.Close)

	c := &HTTPClient{Client: server.Client(), Endpoint: server.URL, Token: "tok"} // AccountID empty
	err := c.UploadITCg(context.Background(), []byte("p"), "cg-1")
	if err == nil {
		t.Fatal("expected 400 from service when accountId empty")
	}
	if gotAccountID != "" {
		t.Errorf("expected empty accountId query value, got %q", gotAccountID)
	}
	var clientErr *Error
	if !errors.As(err, &clientErr) || clientErr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 *Error, got %v", err)
	}
}

// TestForwardWithRetry_5xxExhaustionReturnsNilBody verifies that when the
// 5xx retry budget is exhausted, ForwardWithRetry returns nil response plus
// a status error (body already closed — callers must not read it).
func TestForwardWithRetry_5xxExhaustionReturnsNilBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "down", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	c := &HTTPClient{Client: server.Client(), Endpoint: server.URL}

	res, err := c.ForwardWithRetry(context.Background(), "POST", "/v2/stage-batch", `{}`,
		ForwardRetryOptions{MaxElapsed: 50 * time.Millisecond, RetryOnServerErrors: true})
	if res != nil {
		t.Errorf("expected nil response after exhaustion (body already closed), got status %d", res.StatusCode)
	}
	if err == nil {
		t.Fatal("expected non-nil error on 5xx exhaustion, got nil")
	}
	var clientErr *Error
	if !errors.As(err, &clientErr) {
		t.Fatalf("expected *client.Error, got %T: %v", err, err)
	}
	if clientErr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, clientErr.Code)
	}
}

func TestForwardWithRetry_EmptyBody(t *testing.T) {
	var gotLen int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotLen = len(b)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := &HTTPClient{Client: server.Client(), Endpoint: server.URL}
	res, err := c.ForwardWithRetry(context.Background(), "GET", "/health", "",
		ForwardRetryOptions{MaxElapsed: time.Second, RetryOnServerErrors: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer res.Body.Close()
	if gotLen != 0 {
		t.Errorf("expected empty body, got %d bytes", gotLen)
	}
}

func TestForwardWithRetry_4xxNotRetried(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		http.Error(w, "nope", http.StatusBadRequest)
	}))
	defer server.Close()

	c := &HTTPClient{Client: server.Client(), Endpoint: server.URL}
	res, err := c.ForwardWithRetry(context.Background(), "POST", "/v2/stage-batch", `{}`,
		ForwardRetryOptions{MaxElapsed: 5 * time.Second, RetryOnServerErrors: true})
	if err != nil {
		t.Fatalf("transport error unexpected: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", res.StatusCode)
	}
	if got := atomic.LoadInt32(&attempts); got != 1 {
		t.Errorf("4xx must not retry, got %d attempts", got)
	}
}
