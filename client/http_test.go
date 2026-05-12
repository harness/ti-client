package client

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	v2types "github.com/harness/ti-client/chrysalis/types"
	"github.com/harness/ti-client/types"
)

func TestGetTestTimesSignaturesIncludeEnableAverages(t *testing.T) {
	t.Run("client interface", func(t *testing.T) {
		method, ok := reflect.TypeOf((*Client)(nil)).Elem().MethodByName("GetTestTimes")
		if !ok {
			t.Fatal("GetTestTimes method not found on Client interface")
		}

		if got, want := method.Type.NumIn(), 5; got != want {
			t.Fatalf("Client.GetTestTimes should accept 5 inputs, got %d", got)
		}

		if got := method.Type.In(4).Kind(); got != reflect.Bool {
			t.Fatalf("Client.GetTestTimes last input should be bool, got %s", got)
		}
	})

	t.Run("http client implementation", func(t *testing.T) {
		method, ok := reflect.TypeOf(&HTTPClient{}).MethodByName("GetTestTimes")
		if !ok {
			t.Fatal("GetTestTimes method not found on HTTPClient")
		}

		if got, want := method.Type.NumIn(), 6; got != want {
			t.Fatalf("HTTPClient.GetTestTimes should accept 6 inputs including receiver, got %d", got)
		}

		if got := method.Type.In(5).Kind(); got != reflect.Bool {
			t.Fatalf("HTTPClient.GetTestTimes last input should be bool, got %s", got)
		}
	})
}

func TestGetTestTimesSendsEnableAveragesQueryParam(t *testing.T) {
	var (
		gotPath            string
		gotBuildStartTime  string
		gotEnableAverages  string
		gotRequestPayload  types.GetTestTimesReq
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotBuildStartTime = r.URL.Query().Get("buildStartTime")
		gotEnableAverages = r.URL.Query().Get("enableAverages")
		if err := json.NewDecoder(r.Body).Decode(&gotRequestPayload); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(types.GetTestTimesResp{
			FileTimeMap: map[string]int{"pkg/test.go": 42},
		}); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := &HTTPClient{
		Client:         server.Client(),
		Endpoint:       server.URL,
		Token:          "token",
		AccountID:      "account",
		OrgID:          "org",
		ProjectID:      "project",
		PipelineID:     "pipeline",
		BuildID:        "build",
		StageID:        "stage",
		ParentUniqueID: "parent",
	}

	resp, err := c.GetTestTimes(
		context.Background(),
		"step",
		&types.GetTestTimesReq{IncludeFilename: true},
		1710000000,
		true,
	)
	if err != nil {
		t.Fatalf("GetTestTimes returned error: %v", err)
	}

	if gotPath != "/tests/timedata" {
		t.Fatalf("GetTestTimes should request /tests/timedata, got %q", gotPath)
	}

	if gotBuildStartTime != "1710000000" {
		t.Fatalf("GetTestTimes should send buildStartTime query param, got %q", gotBuildStartTime)
	}

	if gotEnableAverages != "true" {
		t.Fatalf("GetTestTimes should send enableAverages=true, got %q", gotEnableAverages)
	}

	if !gotRequestPayload.IncludeFilename {
		t.Fatal("GetTestTimes should preserve the request payload")
	}

	if got := resp.FileTimeMap["pkg/test.go"]; got != 42 {
		t.Fatalf("GetTestTimes should decode response body, got %d", got)
	}
}

func decodeRequestBody(t *testing.T, r *http.Request) []byte {
	t.Helper()

	var reader io.Reader = r.Body
	if r.Header.Get("Content-Encoding") == "gzip" {
		gr, err := gzip.NewReader(r.Body)
		if err != nil {
			t.Fatalf("failed to create gzip reader: %v", err)
		}
		defer gr.Close()
		reader = gr
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read request body: %v", err)
	}
	return body
}

func TestRequestsSendGzipBodies(t *testing.T) {
	t.Run("Write sends gzip body", func(t *testing.T) {
		var gotEncoding string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotEncoding = r.Header.Get("Content-Encoding")
			var gotTests []*types.TestCase
			if err := json.Unmarshal(decodeRequestBody(t, r), &gotTests); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}
			if len(gotTests) != 1 || gotTests[0].Name != "testOne" {
				t.Fatalf("Write should preserve payload, got %+v", gotTests)
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		c := &HTTPClient{
			Client:       server.Client(),
			Endpoint:     server.URL,
			Token:        "token",
			AccountID:    "account",
			OrgID:        "org",
			ProjectID:    "project",
			PipelineID:   "pipeline",
			BuildID:      "build",
			StageID:      "stage",
			Repo:         "repo",
			Sha:          "sha",
		}

		err := c.Write(context.Background(), "step", "junit", []*types.TestCase{{Name: "testOne"}})
		if err != nil {
			t.Fatalf("Write returned error: %v", err)
		}

		if gotEncoding != "gzip" {
			t.Fatalf("Write should set Content-Encoding=gzip when enabled, got %q", gotEncoding)
		}
	})

	t.Run("UploadCgV2 sends gzip body", func(t *testing.T) {
		var (
			gotEncoding string
			gotPath     string
			gotRepo     string
		)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotEncoding = r.Header.Get("Content-Encoding")
			gotPath = r.URL.Path

			var req v2types.UploadCgRequest
			if err := json.Unmarshal(decodeRequestBody(t, r), &req); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}
			gotRepo = req.Identifier.Repo
			w.WriteHeader(http.StatusAccepted)
		}))
		defer server.Close()

		c := &HTTPClient{
			Client:       server.Client(),
			Endpoint:     server.URL,
			Token:        "token",
			AccountID:    "account",
			OrgID:        "org",
			ProjectID:    "project",
			PipelineID:   "pipeline",
			BuildID:      "build",
			StageID:      "stage",
			Repo:         "repo",
		}

		err := c.UploadCgV2(
			context.Background(),
			v2types.UploadCgRequest{Identifier: v2types.Identifier{Repo: "repo"}},
			"step",
			123,
			url.QueryEscape("feature"),
			url.QueryEscape("main"),
		)
		if err != nil {
			t.Fatalf("UploadCgV2 returned error: %v", err)
		}

		if gotEncoding != "gzip" {
			t.Fatalf("UploadCgV2 should set Content-Encoding=gzip when enabled, got %q", gotEncoding)
		}
		if gotPath != "/v2/uploadcg" {
			t.Fatalf("UploadCgV2 should call /v2/uploadcg, got %q", gotPath)
		}
		if gotRepo != "repo" {
			t.Fatalf("UploadCgV2 should preserve payload, got repo %q", gotRepo)
		}
	})
}

func TestSelectTestsHandlesGzipResponses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request types.SelectTestsReq
		if err := json.Unmarshal(decodeRequestBody(t, r), &request); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		var compressed bytes.Buffer
		gz := gzip.NewWriter(&compressed)
		err := json.NewEncoder(gz).Encode(types.SelectTestsResp{
			Tests: []types.RunnableTest{
				{Pkg: "pkg", Method: "TestOne"},
			},
		})
		if err != nil {
			t.Fatalf("failed to encode gzip response: %v", err)
		}
		if err := gz.Close(); err != nil {
			t.Fatalf("failed to close gzip writer: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Encoding", "gzip")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(compressed.Bytes()); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	c := &HTTPClient{
		Client:         server.Client(),
		Endpoint:       server.URL,
		Token:          "token",
		AccountID:      "account",
		OrgID:          "org",
		ProjectID:      "project",
		PipelineID:     "pipeline",
		BuildID:        "build",
		StageID:        "stage",
		Repo:           "repo",
		Sha:            "sha",
		ParentUniqueID: "parent",
	}

	resp, err := c.SelectTests(
		context.Background(),
		"step",
		"feature",
		"main",
		&types.SelectTestsReq{Files: []types.File{{Name: "main.go"}}},
		false,
	)
	if err != nil {
		t.Fatalf("SelectTests returned error: %v", err)
	}

	if len(resp.Tests) != 1 || resp.Tests[0].Pkg != "pkg" || resp.Tests[0].Method != "TestOne" {
		t.Fatalf("SelectTests should transparently decode gzip response, got %+v", resp.Tests)
	}
}
