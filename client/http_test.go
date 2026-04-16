package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

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
