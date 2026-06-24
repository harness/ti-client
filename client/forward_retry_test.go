package client

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// TestForwardWithRetry_NoRetryWhenMaxElapsedZero verifies that with a zero
// MaxElapsed the method behaves like the legacy Forward — single attempt,
// no retry — even if the server returns an error.
func TestForwardWithRetry_NoRetryWhenMaxElapsedZero(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		http.Error(w, "server busy", http.StatusInternalServerError)
	}))
	defer server.Close()

	c := &HTTPClient{Client: server.Client(), Endpoint: server.URL}

	res, err := c.ForwardWithRetry(context.Background(), "POST", "/v2/stage-batch", `{"stageIndex":0}`,
		ForwardRetryOptions{MaxElapsed: 0, RetryOnServerErrors: true})
	if err != nil {
		t.Fatalf("unexpected transport error: %v", err)
	}
	defer res.Body.Close()

	if got := atomic.LoadInt32(&attempts); got != 1 {
		t.Errorf("expected exactly 1 attempt with MaxElapsed=0, got %d", got)
	}
	if res.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500 status returned to caller, got %d", res.StatusCode)
	}
}

// TestForwardWithRetry_RetriesOn5xxWhenEnabled verifies that 5xx responses
// trigger a retry when RetryOnServerErrors is true, and that the retry
// succeeds once the server starts returning 200.
func TestForwardWithRetry_RetriesOn5xxWhenEnabled(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		// First two attempts return 503, third succeeds.
		if n < 3 {
			http.Error(w, "warming up", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	c := &HTTPClient{Client: server.Client(), Endpoint: server.URL}

	res, err := c.ForwardWithRetry(context.Background(), "POST", "/v2/stage-batch", `{"stageIndex":0}`,
		ForwardRetryOptions{MaxElapsed: 30 * time.Second, RetryOnServerErrors: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected eventual 200, got %d", res.StatusCode)
	}
	if got := atomic.LoadInt32(&attempts); got != 3 {
		t.Errorf("expected 3 attempts (2 retries + final success), got %d", got)
	}
}

// TestForwardWithRetry_NoRetryOn5xxWhenDisabled verifies that
// RetryOnServerErrors=false stops the retry loop on 5xx — used for
// non-idempotent POSTs where a 500 may have partially applied.
func TestForwardWithRetry_NoRetryOn5xxWhenDisabled(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer server.Close()

	c := &HTTPClient{Client: server.Client(), Endpoint: server.URL}

	res, err := c.ForwardWithRetry(context.Background(), "POST", "/api/test-mgmt/set", `{}`,
		ForwardRetryOptions{MaxElapsed: 5 * time.Second, RetryOnServerErrors: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer res.Body.Close()

	if got := atomic.LoadInt32(&attempts); got != 1 {
		t.Errorf("expected exactly 1 attempt with RetryOnServerErrors=false, got %d", got)
	}
	if res.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500 returned to caller, got %d", res.StatusCode)
	}
}

// TestForwardWithRetry_RetriesOnTransportError verifies that a transport-level
// error (server closed before responding) triggers a retry regardless of the
// RetryOnServerErrors setting. Transport errors mean the request never landed,
// so they are always safe to retry.
func TestForwardWithRetry_RetriesOnTransportError(t *testing.T) {
	var attempts int32

	// Server closes the connection mid-response on the first attempt, then
	// behaves normally. Implemented by hijacking the connection.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n == 1 {
			hj, ok := w.(http.Hijacker)
			if !ok {
				t.Fatal("server does not support hijacking")
			}
			conn, _, err := hj.Hijack()
			if err != nil {
				t.Fatalf("hijack failed: %v", err)
			}
			// Slam the connection shut without writing any response. The
			// client will see a transport-level error on read.
			conn.Close()
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	c := &HTTPClient{Client: server.Client(), Endpoint: server.URL}

	res, err := c.ForwardWithRetry(context.Background(), "POST", "/v2/stage-batch", `{"stageIndex":0}`,
		// RetryOnServerErrors=false: we want to confirm transport errors retry
		// even when 5xx retry is disabled.
		ForwardRetryOptions{MaxElapsed: 30 * time.Second, RetryOnServerErrors: false})
	if err != nil {
		t.Fatalf("expected eventual success after retry, got error: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", res.StatusCode)
	}
	if got := atomic.LoadInt32(&attempts); got < 2 {
		t.Errorf("expected at least 2 attempts (1 closed conn + retry), got %d", got)
	}
}

// TestForwardWithRetry_BodyReplayedOnRetry guards the reader-exhaustion bug
// fixed in ForwardWithRetry: each retry must send the full body, not an empty
// payload.
func TestForwardWithRetry_BodyReplayedOnRetry(t *testing.T) {
	var attempts int32
	var lastBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		b, _ := io.ReadAll(r.Body)
		lastBody = string(b)
		if n < 2 {
			http.Error(w, "warming up", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := &HTTPClient{Client: server.Client(), Endpoint: server.URL}

	const bodyText = `{"stageIndex":42}`
	res, err := c.ForwardWithRetry(context.Background(), "POST", "/v2/stage-batch", bodyText,
		ForwardRetryOptions{MaxElapsed: 30 * time.Second, RetryOnServerErrors: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", res.StatusCode)
	}
	// The body the server saw on the retried attempt must match what the caller passed.
	if lastBody != bodyText {
		t.Errorf("retry sent body %q, want %q (reader was likely exhausted)", lastBody, bodyText)
	}
}

// TestForwardWithRetry_HonoursContextCancellation verifies that a cancelled
// context aborts the retry loop promptly rather than waiting out backoff.
func TestForwardWithRetry_HonoursContextCancellation(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		http.Error(w, "down", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	c := &HTTPClient{Client: server.Client(), Endpoint: server.URL}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_, err := c.ForwardWithRetry(ctx, "POST", "/v2/stage-batch", `{}`,
		ForwardRetryOptions{MaxElapsed: 30 * time.Second, RetryOnServerErrors: true})
	if err == nil {
		t.Fatal("expected context-cancellation error, got nil")
	}
	if got := atomic.LoadInt32(&attempts); got > 5 {
		// 200ms / few-ms backoff = at most a handful of attempts. The exact
		// number depends on backoff scheduling; the key check is "didn't
		// spin forever".
		t.Errorf("expected at most ~5 attempts before cancellation, got %d", got)
	}
}
