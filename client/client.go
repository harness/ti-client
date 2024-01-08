package client

import (
	"context"
	"fmt"

	"github.com/harness/ti-client/types"
)

// Error is a custom error struct
type Error struct {
	Code    int
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

// Client defines a TI service client.
type Client interface {
	// Write test cases to DB
	Write(ctx context.Context, step, report string, tests []*types.TestCase) error

	// SelectTests returns list of tests which should be run intelligently
	SelectTests(ctx context.Context, step, source, target string, in *types.SelectTestsReq) (types.SelectTestsResp, error)

	// UploadCg uploads avro encoded callgraph to ti server
	UploadCg(ctx context.Context, step, source, target string, timeMs int64, cg []byte) error

	// DownloadLink returns a list of links where the relevant agent artifacts can be downloaded
	DownloadLink(ctx context.Context, language, os, arch, framework, version, env string) ([]types.DownloadLink, error)

	// GetTestTimes returns the test timing data
	GetTestTimes(ctx context.Context, in *types.GetTestTimesReq) (types.GetTestTimesResp, error)

	// CommitInfo returns the commit id of the last successful commit of a branch for which there is a callgraph
	CommitInfo(ctx context.Context, stepID, branch string) (types.CommitInfoResp, error)

	// MLSelectTests returns list of tests which should be run intelligently using ML Based TI
	MLSelectTests(ctx context.Context, stepID, mlKey, source, target, branch string, in *types.MLSelectTestsRequest) (types.MLSelectTestsResponse, error)

	// Summary returns the summary about test execution information for a build
	Summary(ctx context.Context, summaryRequest types.SummaryRequest) (types.SummaryResponse, error)

	// GetTestCases returns the testcases executed in a build
	GetTestCases(ctx context.Context, testCasesRequest types.TestCasesRequest) (types.TestCases, error)

	//Healthz pings the healthz endpoint
	Healthz(ctx context.Context) error
}
