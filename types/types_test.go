// Copyright 2021 Harness Inc. All rights reserved.
// Use of this source code is governed by the PolyForm Free Trial 1.0.0 license
// that can be found in the licenses directory at the root of this repository, also available at
// https://polyformproject.org/wp-content/uploads/2020/05/PolyForm-Free-Trial-1.0.0.txt.

package types

import "testing"

func TestConvertToFileStatus(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected FileStatus
	}{
		{
			name:     "modified status",
			input:    FileModified,
			expected: FileModified,
		},
		{
			name:     "added status",
			input:    FileAdded,
			expected: FileAdded,
		},
		{
			name:     "deleted status",
			input:    FileDeleted,
			expected: FileDeleted,
		},
		{
			name:     "unknown status defaults to modified",
			input:    "unknown",
			expected: FileModified,
		},
		{
			name:     "empty string defaults to modified",
			input:    "",
			expected: FileModified,
		},
		{
			name:     "case sensitive - lowercase",
			input:    "modified",
			expected: FileModified,
		},
		{
			name:     "case sensitive - uppercase",
			input:    "MODIFIED",
			expected: FileModified, // Still defaults because it doesn't match exactly
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertToFileStatus(tt.input)
			if got != tt.expected {
				t.Errorf("ConvertToFileStatus(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant Status
		expected string
	}{
		{
			name:     "StatusPassed",
			constant: StatusPassed,
			expected: "passed",
		},
		{
			name:     "StatusSkipped",
			constant: StatusSkipped,
			expected: "skipped",
		},
		{
			name:     "StatusFailed",
			constant: StatusFailed,
			expected: "failed",
		},
		{
			name:     "StatusError",
			constant: StatusError,
			expected: "error",
		},
		{
			name:     "StatusSkippedByTI",
			constant: StatusSkippedByTI,
			expected: "skipped_by_ti",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.constant) != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestSelectionConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant Selection
		expected string
	}{
		{
			name:     "SelectSourceCode",
			constant: SelectSourceCode,
			expected: "source_code",
		},
		{
			name:     "SelectNewTest",
			constant: SelectNewTest,
			expected: "new_test",
		},
		{
			name:     "SelectUpdatedTest",
			constant: SelectUpdatedTest,
			expected: "updated_test",
		},
		{
			name:     "SelectPreviousFailure",
			constant: SelectPreviousFailure,
			expected: "previous_failure",
		},
		{
			name:     "SelectFlakyTest",
			constant: SelectFlakyTest,
			expected: "flaky_test",
		},
		{
			name:     "SelectAlwaysRunTest",
			constant: SelectAlwaysRunTest,
			expected: "always_run_test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.constant) != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestFileStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant FileStatus
		expected string
	}{
		{
			name:     "FileModified",
			constant: FileModified,
			expected: "modified",
		},
		{
			name:     "FileAdded",
			constant: FileAdded,
			expected: "added",
		},
		{
			name:     "FileDeleted",
			constant: FileDeleted,
			expected: "deleted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.constant) != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestEnvironmentVariableConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{
			name:     "AccountIDEnv",
			constant: AccountIDEnv,
			expected: "HARNESS_ACCOUNT_ID",
		},
		{
			name:     "OrgIDEnv",
			constant: OrgIDEnv,
			expected: "HARNESS_ORG_ID",
		},
		{
			name:     "ProjectIDEnv",
			constant: ProjectIDEnv,
			expected: "HARNESS_PROJECT_ID",
		},
		{
			name:     "PipelineIDEnv",
			constant: PipelineIDEnv,
			expected: "HARNESS_PIPELINE_ID",
		},
		{
			name:     "StageIDEnv",
			constant: StageIDEnv,
			expected: "HARNESS_STAGE_ID",
		},
		{
			name:     "StepIDEnv",
			constant: StepIDEnv,
			expected: "HARNESS_STEP_ID",
		},
		{
			name:     "BuildIDEnv",
			constant: BuildIDEnv,
			expected: "HARNESS_BUILD_ID",
		},
		{
			name:     "TiSvcEp",
			constant: TiSvcEp,
			expected: "HARNESS_TI_SERVICE_ENDPOINT",
		},
		{
			name:     "TiSvcToken",
			constant: TiSvcToken,
			expected: "HARNESS_TI_SERVICE_TOKEN",
		},
		{
			name:     "InfraEnv",
			constant: InfraEnv,
			expected: "HARNESS_INFRA",
		},
		{
			name:     "HarnessInfra",
			constant: HarnessInfra,
			expected: "VM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

