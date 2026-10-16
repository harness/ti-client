package types

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestTestCaseArtifactsOmittedWhenEmpty(t *testing.T) {
	tests := []struct {
		name string
		tc   TestCase
	}{
		{name: "nil artifacts", tc: TestCase{Name: "a"}},
		{name: "empty artifacts", tc: TestCase{Name: "a", Artifacts: []TestArtifact{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := json.Marshal(tt.tc)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if strings.Contains(string(b), "artifacts") {
				t.Errorf("reports without artifacts must not change on the wire, got %s", b)
			}
		})
	}
}

func TestTestCaseArtifactsRoundTrip(t *testing.T) {
	in := TestCase{
		Name: "login should fail",
		Artifacts: []TestArtifact{
			{Type: TestArtifactTypeTrace, Path: "login-chromium/trace.zip", FileName: "trace.zip", Size: 8524},
			{Type: TestArtifactTypeScreenshot, Path: "login-chromium-retry1/test-failed-1.png", FileName: "test-failed-1.png", Size: 5664, Attempt: 1},
		},
	}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var out TestCase
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("round trip mismatch:\n in: %+v\nout: %+v", in, out)
	}
}

func TestTestArtifactJSONShape(t *testing.T) {
	tests := []struct {
		name string
		in   TestArtifact
		want string
	}{
		{
			name: "first run omits attempt",
			in:   TestArtifact{Type: TestArtifactTypeVideo, Path: "p/video.webm", FileName: "video.webm", Size: 1},
			want: `{"type":"video","path":"p/video.webm","file_name":"video.webm","size":1}`,
		},
		{
			name: "retry keeps attempt",
			in:   TestArtifact{Type: TestArtifactTypeTrace, Path: "p/trace.zip", FileName: "trace.zip", Size: 2, Attempt: 2},
			want: `{"type":"trace","path":"p/trace.zip","file_name":"trace.zip","size":2,"attempt":2}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := json.Marshal(tt.in)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if string(b) != tt.want {
				t.Errorf("got %s, want %s", b, tt.want)
			}
		})
	}
}

func TestTestCaseDecodesReportWithoutArtifacts(t *testing.T) {
	// A report written before this field existed must still decode.
	var tc TestCase
	if err := json.Unmarshal([]byte(`{"name":"a","class_name":"c","duration_ms":3}`), &tc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if tc.Name != "a" || tc.DurationMs != 3 {
		t.Errorf("unexpected decode: %+v", tc)
	}
	if tc.Artifacts != nil {
		t.Errorf("expected nil artifacts, got %+v", tc.Artifacts)
	}
}
