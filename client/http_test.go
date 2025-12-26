// Copyright 2021 Harness Inc. All rights reserved.
// Use of this source code is governed by the PolyForm Free Trial 1.0.0 license
// that can be found in the licenses directory at the root of this repository, also available at
// https://polyformproject.org/wp-content/uploads/2020/05/PolyForm-Free-Trial-1.0.0.txt.

package client

import (
	"testing"

	"github.com/harness/ti-client/types"
)

func TestHTTPClient_validateTiArgs(t *testing.T) {
	tests := []struct {
		name    string
		client  *HTTPClient
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid args",
			client: &HTTPClient{
				Endpoint: "https://ti-service.example.com",
				Token:    "test-token",
			},
			wantErr: false,
		},
		{
			name: "missing endpoint",
			client: &HTTPClient{
				Token: "test-token",
			},
			wantErr: true,
			errMsg:  "ti endpoint is not set",
		},
		{
			name: "missing token",
			client: &HTTPClient{
				Endpoint: "https://ti-service.example.com",
			},
			wantErr: true,
			errMsg:  "ti token is not set",
		},
		{
			name:    "missing both",
			client:  &HTTPClient{},
			wantErr: true,
			errMsg:  "ti endpoint is not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.client.validateTiArgs()
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTiArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("validateTiArgs() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestHTTPClient_validateBasicArgs(t *testing.T) {
	tests := []struct {
		name    string
		client  *HTTPClient
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid args",
			client: &HTTPClient{
				AccountID:  "account123",
				OrgID:      "org123",
				ProjectID:  "project123",
				PipelineID: "pipeline123",
			},
			wantErr: false,
		},
		{
			name: "missing accountID",
			client: &HTTPClient{
				OrgID:      "org123",
				ProjectID:  "project123",
				PipelineID: "pipeline123",
			},
			wantErr: true,
			errMsg:  "accountID is not set",
		},
		{
			name: "missing orgID",
			client: &HTTPClient{
				AccountID:  "account123",
				ProjectID:  "project123",
				PipelineID: "pipeline123",
			},
			wantErr: true,
			errMsg:  "orgID is not set",
		},
		{
			name: "missing projectID",
			client: &HTTPClient{
				AccountID:  "account123",
				OrgID:      "org123",
				PipelineID: "pipeline123",
			},
			wantErr: true,
			errMsg:  "projectID is not set",
		},
		{
			name: "missing pipelineID",
			client: &HTTPClient{
				AccountID: "account123",
				OrgID:     "org123",
				ProjectID: "project123",
			},
			wantErr: true,
			errMsg:  "pipelineID is not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.client.validateBasicArgs()
			if (err != nil) != tt.wantErr {
				t.Errorf("validateBasicArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("validateBasicArgs() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestHTTPClient_validateWriteArgs(t *testing.T) {
	tests := []struct {
		name    string
		client  *HTTPClient
		stepID  string
		report  string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid args",
			client: &HTTPClient{
				Endpoint:  "https://ti-service.example.com",
				Token:     "test-token",
				AccountID: "account123",
				OrgID:     "org123",
				ProjectID: "project123",
				PipelineID: "pipeline123",
				BuildID:   "build123",
				StageID:   "stage123",
			},
			stepID:  "step123",
			report:  "junit",
			wantErr: false,
		},
		{
			name: "missing stepID",
			client: &HTTPClient{
				Endpoint:  "https://ti-service.example.com",
				Token:     "test-token",
				AccountID: "account123",
				OrgID:     "org123",
				ProjectID: "project123",
				PipelineID: "pipeline123",
				BuildID:   "build123",
				StageID:   "stage123",
			},
			stepID:  "",
			report:  "junit",
			wantErr: true,
			errMsg:  "stepID is not set",
		},
		{
			name: "missing report",
			client: &HTTPClient{
				Endpoint:  "https://ti-service.example.com",
				Token:     "test-token",
				AccountID: "account123",
				OrgID:     "org123",
				ProjectID: "project123",
				PipelineID: "pipeline123",
				BuildID:   "build123",
				StageID:   "stage123",
			},
			stepID:  "step123",
			report:  "",
			wantErr: true,
			errMsg:  "report is not set",
		},
		{
			name: "missing buildID",
			client: &HTTPClient{
				Endpoint:  "https://ti-service.example.com",
				Token:     "test-token",
				AccountID: "account123",
				OrgID:     "org123",
				ProjectID: "project123",
				PipelineID: "pipeline123",
				StageID:   "stage123",
			},
			stepID:  "step123",
			report:  "junit",
			wantErr: true,
			errMsg:  "buildID is not set",
		},
		{
			name: "missing stageID",
			client: &HTTPClient{
				Endpoint:  "https://ti-service.example.com",
				Token:     "test-token",
				AccountID: "account123",
				OrgID:     "org123",
				ProjectID: "project123",
				PipelineID: "pipeline123",
				BuildID:   "build123",
			},
			stepID:  "step123",
			report:  "junit",
			wantErr: true,
			errMsg:  "stageID is not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.client.validateWriteArgs(tt.stepID, tt.report)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateWriteArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("validateWriteArgs() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestHTTPClient_validateDownloadLinkArgs(t *testing.T) {
	tests := []struct {
		name     string
		client   *HTTPClient
		language string
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid args",
			client: &HTTPClient{
				Endpoint: "https://ti-service.example.com",
				Token:    "test-token",
			},
			language: "java",
			wantErr:  false,
		},
		{
			name: "missing language",
			client: &HTTPClient{
				Endpoint: "https://ti-service.example.com",
				Token:    "test-token",
			},
			language: "",
			wantErr:  true,
			errMsg:   "language is not set",
		},
		{
			name: "missing endpoint",
			client: &HTTPClient{
				Token: "test-token",
			},
			language: "java",
			wantErr:  true,
			errMsg:   "ti endpoint is not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.client.validateDownloadLinkArgs(tt.language)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDownloadLinkArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("validateDownloadLinkArgs() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestHTTPClient_validateSelectTestsArgs(t *testing.T) {
	tests := []struct {
		name    string
		client  *HTTPClient
		stepID  string
		source  string
		target  string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid args",
			client: &HTTPClient{
				Endpoint:  "https://ti-service.example.com",
				Token:     "test-token",
				AccountID: "account123",
				OrgID:     "org123",
				ProjectID: "project123",
				PipelineID: "pipeline123",
				BuildID:   "build123",
				StageID:   "stage123",
			},
			stepID:  "step123",
			source:  "feature-branch",
			target:  "main",
			wantErr: false,
		},
		{
			name: "missing stepID",
			client: &HTTPClient{
				Endpoint:  "https://ti-service.example.com",
				Token:     "test-token",
				AccountID: "account123",
				OrgID:     "org123",
				ProjectID: "project123",
				PipelineID: "pipeline123",
				BuildID:   "build123",
				StageID:   "stage123",
			},
			stepID:  "",
			source:  "feature-branch",
			target:  "main",
			wantErr: true,
			errMsg:  "stepID is not set",
		},
		{
			name: "missing source",
			client: &HTTPClient{
				Endpoint:  "https://ti-service.example.com",
				Token:     "test-token",
				AccountID: "account123",
				OrgID:     "org123",
				ProjectID: "project123",
				PipelineID: "pipeline123",
				BuildID:   "build123",
				StageID:   "stage123",
			},
			stepID:  "step123",
			source:  "",
			target:  "main",
			wantErr: true,
			errMsg:  "source branch is not set",
		},
		{
			name: "missing target",
			client: &HTTPClient{
				Endpoint:  "https://ti-service.example.com",
				Token:     "test-token",
				AccountID: "account123",
				OrgID:     "org123",
				ProjectID: "project123",
				PipelineID: "pipeline123",
				BuildID:   "build123",
				StageID:   "stage123",
			},
			stepID:  "step123",
			source:  "feature-branch",
			target:  "",
			wantErr: true,
			errMsg:  "target branch is not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.client.validateSelectTestsArgs(tt.stepID, tt.source, tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSelectTestsArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("validateSelectTestsArgs() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestHTTPClient_validateUploadCgArgs(t *testing.T) {
	tests := []struct {
		name    string
		client  *HTTPClient
		stepID  string
		source  string
		target  string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid args",
			client: &HTTPClient{
				Endpoint:  "https://ti-service.example.com",
				Token:     "test-token",
				AccountID: "account123",
				OrgID:     "org123",
				ProjectID: "project123",
				PipelineID: "pipeline123",
				BuildID:   "build123",
				StageID:   "stage123",
			},
			stepID:  "step123",
			source:  "feature-branch",
			target:  "main",
			wantErr: false,
		},
		{
			name: "missing stepID",
			client: &HTTPClient{
				Endpoint:  "https://ti-service.example.com",
				Token:     "test-token",
				AccountID: "account123",
				OrgID:     "org123",
				ProjectID: "project123",
				PipelineID: "pipeline123",
				BuildID:   "build123",
				StageID:   "stage123",
			},
			stepID:  "",
			source:  "feature-branch",
			target:  "main",
			wantErr: true,
			errMsg:  "stepID is not set",
		},
		{
			name: "missing source",
			client: &HTTPClient{
				Endpoint:  "https://ti-service.example.com",
				Token:     "test-token",
				AccountID: "account123",
				OrgID:     "org123",
				ProjectID: "project123",
				PipelineID: "pipeline123",
				BuildID:   "build123",
				StageID:   "stage123",
			},
			stepID:  "step123",
			source:  "",
			target:  "main",
			wantErr: true,
			errMsg:  "source branch is not set",
		},
		{
			name: "missing target",
			client: &HTTPClient{
				Endpoint:  "https://ti-service.example.com",
				Token:     "test-token",
				AccountID: "account123",
				OrgID:     "org123",
				ProjectID: "project123",
				PipelineID: "pipeline123",
				BuildID:   "build123",
				StageID:   "stage123",
			},
			stepID:  "step123",
			source:  "feature-branch",
			target:  "",
			wantErr: true,
			errMsg:  "target branch is not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.client.validateUploadCgArgs(tt.stepID, tt.source, tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateUploadCgArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("validateUploadCgArgs() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestHTTPClient_validateCommitInfoArgs(t *testing.T) {
	tests := []struct {
		name    string
		client  *HTTPClient
		stepID  string
		branch  string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid args",
			client: &HTTPClient{
				Endpoint:  "https://ti-service.example.com",
				Token:     "test-token",
				AccountID: "account123",
				OrgID:     "org123",
				ProjectID: "project123",
				PipelineID: "pipeline123",
				BuildID:   "build123",
				StageID:   "stage123",
			},
			stepID:  "step123",
			branch:  "main",
			wantErr: false,
		},
		{
			name: "missing stepID",
			client: &HTTPClient{
				Endpoint:  "https://ti-service.example.com",
				Token:     "test-token",
				AccountID: "account123",
				OrgID:     "org123",
				ProjectID: "project123",
				PipelineID: "pipeline123",
				BuildID:   "build123",
				StageID:   "stage123",
			},
			stepID:  "",
			branch:  "main",
			wantErr: true,
			errMsg:  "stepID is not set",
		},
		{
			name: "missing branch",
			client: &HTTPClient{
				Endpoint:  "https://ti-service.example.com",
				Token:     "test-token",
				AccountID: "account123",
				OrgID:     "org123",
				ProjectID: "project123",
				PipelineID: "pipeline123",
				BuildID:   "build123",
				StageID:   "stage123",
			},
			stepID:  "step123",
			branch:  "",
			wantErr: true,
			errMsg:  "source branch is not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.client.validateCommitInfoArgs(tt.stepID, tt.branch)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCommitInfoArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("validateCommitInfoArgs() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestHTTPClient_validateSubmitChecksumsArgs(t *testing.T) {
	tests := []struct {
		name      string
		checksums map[string]uint64
		client    *HTTPClient
		wantErr   bool
		errCode   int
		errMsg    string
	}{
		{
			name: "valid checksums",
			checksums: map[string]uint64{
				"file1.go": 12345,
				"file2.go": 67890,
			},
			client: &HTTPClient{
				AccountID:  "account123",
				OrgID:      "org123",
				ProjectID:  "project123",
				PipelineID: "pipeline123",
			},
			wantErr: false,
		},
		{
			name:      "empty checksums",
			checksums: map[string]uint64{},
			client: &HTTPClient{
				AccountID:  "account123",
				OrgID:      "org123",
				ProjectID:  "project123",
				PipelineID: "pipeline123",
			},
			wantErr: true,
			errCode: 400,
			errMsg:  "checksums map cannot be empty",
		},
		{
			name: "empty filepath",
			checksums: map[string]uint64{
				"": 12345,
			},
			client: &HTTPClient{
				AccountID:  "account123",
				OrgID:      "org123",
				ProjectID:  "project123",
				PipelineID: "pipeline123",
			},
			wantErr: true,
			errCode: 400,
			errMsg:  "filepath cannot be empty",
		},
		{
			name: "zero checksum",
			checksums: map[string]uint64{
				"file1.go": 0,
			},
			client: &HTTPClient{
				AccountID:  "account123",
				OrgID:      "org123",
				ProjectID:  "project123",
				PipelineID: "pipeline123",
			},
			wantErr: true,
			errCode: 400,
			errMsg:  "checksum cannot be zero for file: file1.go",
		},
		{
			name: "missing accountID",
			checksums: map[string]uint64{
				"file1.go": 12345,
			},
			client: &HTTPClient{
				OrgID:      "org123",
				ProjectID:  "project123",
				PipelineID: "pipeline123",
			},
			wantErr: true,
			errMsg:  "accountID is not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.client.validateSubmitChecksumsArgs(tt.checksums)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSubmitChecksumsArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if tt.errCode > 0 {
					if clientErr, ok := err.(*Error); ok {
						if clientErr.Code != tt.errCode {
							t.Errorf("validateSubmitChecksumsArgs() error code = %v, want %v", clientErr.Code, tt.errCode)
						}
						if clientErr.Message != tt.errMsg {
							t.Errorf("validateSubmitChecksumsArgs() error message = %v, want %v", clientErr.Message, tt.errMsg)
						}
					} else if err.Error() != tt.errMsg {
						t.Errorf("validateSubmitChecksumsArgs() error = %v, want %v", err.Error(), tt.errMsg)
					}
				} else if err != nil && err.Error() != tt.errMsg {
					t.Errorf("validateSubmitChecksumsArgs() error = %v, want %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestHTTPClient_SetBasicArguments(t *testing.T) {
	tests := []struct {
		name           string
		client         *HTTPClient
		summaryRequest *types.SummaryRequest
		want           *types.SummaryRequest
	}{
		{
			name: "fill all empty fields",
			client: &HTTPClient{
				OrgID:      "org123",
				ProjectID:  "project123",
				PipelineID: "pipeline123",
				BuildID:    "build123",
			},
			summaryRequest: &types.SummaryRequest{},
			want: &types.SummaryRequest{
				OrgID:      "org123",
				ProjectID:  "project123",
				PipelineID: "pipeline123",
				BuildID:    "build123",
				ReportType: "junit",
			},
		},
		{
			name: "preserve existing values",
			client: &HTTPClient{
				OrgID:      "org123",
				ProjectID:  "project123",
				PipelineID: "pipeline123",
				BuildID:    "build123",
			},
			summaryRequest: &types.SummaryRequest{
				OrgID:      "existing-org",
				ProjectID:  "existing-project",
				PipelineID: "existing-pipeline",
				BuildID:    "existing-build",
				ReportType: "custom-report",
			},
			want: &types.SummaryRequest{
				OrgID:      "existing-org",
				ProjectID:  "existing-project",
				PipelineID: "existing-pipeline",
				BuildID:    "existing-build",
				ReportType: "custom-report",
			},
		},
		{
			name: "all stages clears stage and step",
			client: &HTTPClient{
				OrgID:      "org123",
				ProjectID:  "project123",
				PipelineID: "pipeline123",
				BuildID:    "build123",
			},
			summaryRequest: &types.SummaryRequest{
				AllStages: true,
				StageID:   "stage123",
				StepID:    "step123",
			},
			want: &types.SummaryRequest{
				OrgID:      "org123",
				ProjectID:  "project123",
				PipelineID: "pipeline123",
				BuildID:    "build123",
				ReportType: "junit",
				AllStages:  true,
				StageID:    "",
				StepID:     "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.client.SetBasicArguments(tt.summaryRequest)
			if tt.summaryRequest.OrgID != tt.want.OrgID {
				t.Errorf("SetBasicArguments() OrgID = %v, want %v", tt.summaryRequest.OrgID, tt.want.OrgID)
			}
			if tt.summaryRequest.ProjectID != tt.want.ProjectID {
				t.Errorf("SetBasicArguments() ProjectID = %v, want %v", tt.summaryRequest.ProjectID, tt.want.ProjectID)
			}
			if tt.summaryRequest.PipelineID != tt.want.PipelineID {
				t.Errorf("SetBasicArguments() PipelineID = %v, want %v", tt.summaryRequest.PipelineID, tt.want.PipelineID)
			}
			if tt.summaryRequest.BuildID != tt.want.BuildID {
				t.Errorf("SetBasicArguments() BuildID = %v, want %v", tt.summaryRequest.BuildID, tt.want.BuildID)
			}
			if tt.summaryRequest.ReportType != tt.want.ReportType {
				t.Errorf("SetBasicArguments() ReportType = %v, want %v", tt.summaryRequest.ReportType, tt.want.ReportType)
			}
			if tt.summaryRequest.StageID != tt.want.StageID {
				t.Errorf("SetBasicArguments() StageID = %v, want %v", tt.summaryRequest.StageID, tt.want.StageID)
			}
			if tt.summaryRequest.StepID != tt.want.StepID {
				t.Errorf("SetBasicArguments() StepID = %v, want %v", tt.summaryRequest.StepID, tt.want.StepID)
			}
		})
	}
}

func TestHTTPClient_validateWriteSavingsArgs(t *testing.T) {
	tests := []struct {
		name    string
		client  *HTTPClient
		stepID  string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid args",
			client: &HTTPClient{
				Endpoint:  "https://ti-service.example.com",
				Token:     "test-token",
				AccountID: "account123",
				OrgID:     "org123",
				ProjectID: "project123",
				PipelineID: "pipeline123",
				BuildID:   "build123",
				StageID:   "stage123",
			},
			stepID:  "step123",
			wantErr: false,
		},
		{
			name: "missing stepID",
			client: &HTTPClient{
				Endpoint:  "https://ti-service.example.com",
				Token:     "test-token",
				AccountID: "account123",
				OrgID:     "org123",
				ProjectID: "project123",
				PipelineID: "pipeline123",
				BuildID:   "build123",
				StageID:   "stage123",
			},
			stepID:  "",
			wantErr: true,
			errMsg:  "stepID is not set",
		},
		{
			name: "missing buildID",
			client: &HTTPClient{
				Endpoint:  "https://ti-service.example.com",
				Token:     "test-token",
				AccountID: "account123",
				OrgID:     "org123",
				ProjectID: "project123",
				PipelineID: "pipeline123",
				StageID:   "stage123",
			},
			stepID:  "step123",
			wantErr: true,
			errMsg:  "buildID is not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.client.validateWriteSavingsArgs(tt.stepID)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateWriteSavingsArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("validateWriteSavingsArgs() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestHTTPClient_validateGetTestTimesArgs(t *testing.T) {
	tests := []struct {
		name    string
		client  *HTTPClient
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid args",
			client: &HTTPClient{
				Endpoint:  "https://ti-service.example.com",
				Token:     "test-token",
				AccountID: "account123",
				OrgID:     "org123",
				ProjectID: "project123",
				PipelineID: "pipeline123",
			},
			wantErr: false,
		},
		{
			name: "missing endpoint",
			client: &HTTPClient{
				Token:     "test-token",
				AccountID: "account123",
				OrgID:     "org123",
				ProjectID: "project123",
				PipelineID: "pipeline123",
			},
			wantErr: true,
			errMsg:  "ti endpoint is not set",
		},
		{
			name: "missing accountID",
			client: &HTTPClient{
				Endpoint:  "https://ti-service.example.com",
				Token:     "test-token",
				OrgID:     "org123",
				ProjectID: "project123",
				PipelineID: "pipeline123",
			},
			wantErr: true,
			errMsg:  "accountID is not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.client.validateGetTestTimesArgs()
			if (err != nil) != tt.wantErr {
				t.Errorf("validateGetTestTimesArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("validateGetTestTimesArgs() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestHTTPClient_validateMLSelectTestArgs(t *testing.T) {
	tests := []struct {
		name    string
		client  *HTTPClient
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid args",
			client: &HTTPClient{
				Endpoint:  "https://ti-service.example.com",
				Token:     "test-token",
				AccountID: "account123",
				OrgID:     "org123",
				ProjectID: "project123",
				PipelineID: "pipeline123",
			},
			wantErr: false,
		},
		{
			name: "missing endpoint",
			client: &HTTPClient{
				Token:     "test-token",
				AccountID: "account123",
				OrgID:     "org123",
				ProjectID: "project123",
				PipelineID: "pipeline123",
			},
			wantErr: true,
			errMsg:  "ti endpoint is not set",
		},
		{
			name: "missing accountID",
			client: &HTTPClient{
				Endpoint:  "https://ti-service.example.com",
				Token:     "test-token",
				OrgID:     "org123",
				ProjectID: "project123",
				PipelineID: "pipeline123",
			},
			wantErr: true,
			errMsg:  "accountID is not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.client.validateMLSelectTestArgs()
			if (err != nil) != tt.wantErr {
				t.Errorf("validateMLSelectTestArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("validateMLSelectTestArgs() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

