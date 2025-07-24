// Copyright 2021 Harness Inc. All rights reserved.
// Use of this source code is governed by the PolyForm Free Trial 1.0.0 license
// that can be found in the licenses directory at the root of this repository, also available at
// https://polyformproject.org/wp-content/uploads/2020/05/PolyForm-Free-Trial-1.0.0.txt.

package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/harness/ti-client/types"
)

var _ Client = (*HTTPClient)(nil)

const (
	dbEndpoint            = "/reports/write?accountId=%s&orgId=%s&projectId=%s&pipelineId=%s&buildId=%s&stageId=%s&stepId=%s&report=%s&repo=%s&sha=%s&commitLink=%s"
	testEndpoint          = "/tests/select?accountId=%s&orgId=%s&projectId=%s&pipelineId=%s&buildId=%s&stageId=%s&stepId=%s&repo=%s&sha=%s&source=%s&target=%s"
	cgEndpoint            = "/tests/uploadcg?accountId=%s&orgId=%s&projectId=%s&pipelineId=%s&buildId=%s&stageId=%s&stepId=%s&repo=%s&sha=%s&source=%s&target=%s&timeMs=%d&schemaVersion=1.1"
	cgEndpointFailedTest  = "/tests/uploadcg?accountId=%s&orgId=%s&projectId=%s&pipelineId=%s&buildId=%s&stageId=%s&stepId=%s&repo=%s&sha=%s&source=%s&target=%s&timeMs=%d&hasFailedTests=true"
	uploadcgEndpoint      = "/v2/uploadcg"
	getTestsTimesEndpoint = "/tests/timedata?accountId=%s&orgId=%s&projectId=%s&pipelineId=%s&buildId=%s&stageId=%s&stepId=%s"
	agentEndpoint         = "/agents/link?accountId=%s&language=%s&os=%s&arch=%s&framework=%s&version=%s&buildenv=%s"
	commitInfoEndpoint    = "/vcs/commitinfo?accountId=%s&orgId=%s&projectId=%s&pipelineId=%s&buildId=%s&stageId=%s&stepId=%s&repo=%s&branch=%s"
	mlSelectTestsEndpoint = "/ml/tests/select?accountId=%s&orgId=%s&projectId=%s&pipelineId=%s&buildId=%s&stageId=%s&stepId=%s&repo=%s&sha=%s&source=%s&target=%s&mlKey=%s&commitLink=%s"
	summaryEndpoint       = "/reports/summary?accountId=%s&orgId=%s&projectId=%s&pipelineId=%s&buildId=%s&stageId=%s&stepId=%s&report=%s"
	testCasesEndpoint     = "/reports/test_cases?accountId=%s&orgId=%s&projectId=%s&pipelineId=%s&buildId=%s&stageId=%s&stepId=%s&report=%s&testCaseSearchTerm=%s&sort=%s&order=%s&pageIndex=%s&pageSize=%s&suite_name=%s"
	healthzEndpoint       = "/healthz"
	// savings
	savingsEndpoint = "/savings?accountId=%s&orgId=%s&projectId=%s&pipelineId=%s&buildId=%s&stageId=%s&stepId=%s&repo=%s&featureName=%s&featureState=%s&timeMs=%s"
)

// defaultClient is the default http.Client.
var defaultClient = &http.Client{
	CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

// NewHTTPClient returns a new HTTPClient with optional mTLS and custom root certificates.
func NewHTTPClient(endpoint, token, accountID, orgID, projectID, pipelineID, buildID, stageID, repo, sha, commitLink string, skipverify bool, additionalCertsDir, base64MtlsClientCert, base64MtlsClientCertKey string) *HTTPClient {
	endpoint = strings.TrimSuffix(endpoint, "/")
	client := &HTTPClient{
		Endpoint:   endpoint,
		Token:      token,
		AccountID:  accountID,
		OrgID:      orgID,
		ProjectID:  projectID,
		PipelineID: pipelineID,
		BuildID:    buildID,
		StageID:    stageID,
		Repo:       repo,
		Sha:        sha,
		CommitLink: commitLink,
		SkipVerify: skipverify,
	}

	// Load mTLS certificates if available
	mtlsEnabled, mtlsCerts := loadMTLSCerts(base64MtlsClientCert, base64MtlsClientCertKey, "/etc/mtls/client.crt", "/etc/mtls/client.key")

	// Load custom root CAs if additional certificates directory is provided
	rootCAs := loadRootCAs(additionalCertsDir)

	// Only create HTTP client if needed (mTLS, additional certs, or skipverify)
	if skipverify || rootCAs != nil || mtlsEnabled {
		client.Client = clientWithTLSConfig(skipverify, rootCAs, mtlsEnabled, mtlsCerts)
	}

	return client
}

// loadMTLSCerts determines the source of mTLS certificates based on base64 strings or file paths
func loadMTLSCerts(base64Cert, base64Key, defaultCertFile, defaultKeyFile string) (bool, tls.Certificate) {

	// Attempt to load from base64 strings
	if base64Cert != "" && base64Key != "" {
		cert, err := loadCertsFromBase64(base64Cert, base64Key)
		if err == nil {
			return true, cert
		}
		fmt.Printf("failed to load mTLS certs from base64, error: %s\n", err)
	}

	// Fallback to default paths
	return loadMTLSCertsFromFiles(defaultCertFile, defaultKeyFile)
}

// loadCertsFromBase64 loads certificates from base64-encoded strings
func loadCertsFromBase64(certBase64, keyBase64 string) (tls.Certificate, error) {
	certBytes, err := base64.StdEncoding.DecodeString(certBase64)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to decode base64 certificate: %w", err)
	}
	keyBytes, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to decode base64 key: %w", err)
	}
	return tls.X509KeyPair(certBytes, keyBytes)
}

// loadMTLSCertsFromFiles loads mTLS certificates from file paths
func loadMTLSCertsFromFiles(certFile, keyFile string) (bool, tls.Certificate) {
	if fileExists(certFile) && fileExists(keyFile) {
		mtlsCerts, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			fmt.Printf("failed to load mTLS cert/key pair, error: %s\n", err)
			return false, tls.Certificate{}
		}
		return true, mtlsCerts
	}
	return false, tls.Certificate{}
}

// loadRootCAs loads custom root CAs from the provided directory
func loadRootCAs(additionalCertsDir string) *x509.CertPool {
	if additionalCertsDir == "" {
		return nil
	}

	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	fmt.Printf("additional certs dir to allow: %s\n", additionalCertsDir)

	files, err := os.ReadDir(additionalCertsDir)
	if err != nil {
		fmt.Printf("could not read directory %s, error: %s\n", additionalCertsDir, err)
		return rootCAs
	}

	// Go through all certs in this directory and add them to the global certs
	for _, f := range files {
		path := filepath.Join(additionalCertsDir, f.Name())
		fmt.Printf("trying to add certs at: %s to root certs\n", path)
		// Create TLS config using cert PEM
		rootPem, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("could not read certificate file (%s), error: %s\n", path, err.Error())
			continue
		}
		// Append certs to the global certs
		ok := rootCAs.AppendCertsFromPEM(rootPem)
		if !ok {
			fmt.Printf("error adding cert (%s) to pool, please check format of the certs provided.\n", path)
			continue
		}
		fmt.Printf("successfully added cert at: %s to root certs\n", path)
	}
	return rootCAs
}

// clientWithTLSConfig creates an HTTP client with the provided TLS settings
func clientWithTLSConfig(skipverify bool, rootCAs *x509.CertPool, mtlsEnabled bool, cert tls.Certificate) *http.Client {
	config := &tls.Config{
		InsecureSkipVerify: skipverify,
	}
	// Only use rootCAs if skipverify is false
	if !skipverify && rootCAs != nil {
		config.RootCAs = rootCAs
	}
	if mtlsEnabled {
		fmt.Println("setting mTLS Client Certs in TI Service Client")
		config.Certificates = []tls.Certificate{cert}
	}
	return &http.Client{
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: config,
		},
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	return err == nil && !info.IsDir()
}

// HTTPClient provides an http service client.
type HTTPClient struct {
	Client     *http.Client
	Endpoint   string // Example: http://localhost:port
	Token      string
	AccountID  string
	OrgID      string
	ProjectID  string
	PipelineID string
	BuildID    string
	StageID    string
	Repo       string
	Sha        string
	CommitLink string
	SkipVerify bool
}

// Write writes test results to the TI server
func (c *HTTPClient) Write(ctx context.Context, stepID, report string, tests []*types.TestCase) error {
	if err := c.validateWriteArgs(stepID, report); err != nil {
		return err
	}
	path := fmt.Sprintf(dbEndpoint, c.AccountID, c.OrgID, c.ProjectID, c.PipelineID, c.BuildID, c.StageID, stepID, report, c.Repo, c.Sha, c.CommitLink)
	backoff := createBackoff(10 * 60 * time.Second)
	_, err := c.retry(ctx, c.Endpoint+path, "POST", c.Sha, &tests, nil, false, false, backoff) //nolint:bodyclose
	return err
}

// DownloadLink returns a list of links where the relevant agent artifacts can be downloaded
func (c *HTTPClient) DownloadLink(ctx context.Context, language, os, arch, framework, version, env string) ([]types.DownloadLink, error) {
	var resp []types.DownloadLink
	if err := c.validateDownloadLinkArgs(language); err != nil {
		return resp, err
	}
	path := fmt.Sprintf(agentEndpoint, c.AccountID, language, os, arch, framework, version, env)
	backoff := createBackoff(5 * 60 * time.Second)
	_, err := c.retry(ctx, c.Endpoint+path, "GET", "", nil, &resp, false, true, backoff) //nolint:bodyclose
	return resp, err
}

// SelectTests returns a list of tests which should be run intelligently
func (c *HTTPClient) SelectTests(ctx context.Context, stepID, source, target string, in *types.SelectTestsReq, failedTestRerunEnabled bool) (types.SelectTestsResp, error) {
	var resp types.SelectTestsResp
	if err := c.validateSelectTestsArgs(stepID, source, target); err != nil {
		return resp, err
	}
	path := fmt.Sprintf(testEndpoint, c.AccountID, c.OrgID, c.ProjectID, c.PipelineID, c.BuildID, c.StageID, stepID, c.Repo, c.Sha, source, target)
	if failedTestRerunEnabled {
		path += "&failedTestRerunEnabled=true"
	}
	backoff := createBackoff(10 * 60 * time.Second)
	_, err := c.retry(ctx, c.Endpoint+path, "POST", c.Sha, in, &resp, false, false, backoff) //nolint:bodyclose
	return resp, err
}

func (c *HTTPClient) UploadCgFailedTest(ctx context.Context, stepID, source, target string, timeMs int64, cg []byte) error {
	return c.uploadCGInternal(ctx, stepID, source, target, timeMs, cg, cgEndpointFailedTest)
}

// UploadCg uploads avro encoded callgraph to server
func (c *HTTPClient) UploadCg(ctx context.Context, stepID, source, target string, timeMs int64, cg []byte, failedTestRerunEnabled bool) error {
	cgEndpointFF := cgEndpoint
	if failedTestRerunEnabled {
		cgEndpointFF = cgEndpoint + "&failedTestRerunEnabled=true"
	}

	return c.uploadCGInternal(ctx, stepID, source, target, timeMs, cg, cgEndpointFF)
}

// UploadCgV2 uploads JSON payload to /uploadcg endpoint
func (c *HTTPClient) UploadCgV2(ctx context.Context, jsonPayload interface{}) error {
	if err := c.validateTiArgs(); err != nil {
		return err
	}
	backoff := createBackoff(45 * 60 * time.Second)

	if payloadStr, ok := jsonPayload.(string); ok {
		// If the payload is a string, treat it as raw JSON and pass it as an io.Reader.
		reader := strings.NewReader(payloadStr)
		_, err := c.retry(ctx, c.Endpoint+uploadcgEndpoint, "POST", "", reader, nil, true, true, backoff) //nolint:bodyclose
		return err
	}

	// For other types, use the existing behavior to JSON-encode the payload.

	return errors.New("payload type not supported")
}

func (c *HTTPClient) uploadCGInternal(ctx context.Context, stepID, source, target string, timeMs int64, cg []byte, endpoint string) error {
	if err := c.validateUploadCgArgs(stepID, source, target); err != nil {
		return err
	}
	path := fmt.Sprintf(endpoint, c.AccountID, c.OrgID, c.ProjectID, c.PipelineID, c.BuildID, c.StageID, stepID, c.Repo, c.Sha, source, target, timeMs)
	backoff := createBackoff(45 * 60 * time.Second)
	_, err := c.retry(ctx, c.Endpoint+path, "POST", c.Sha, &cg, nil, false, true, backoff) //nolint:bodyclose
	return err
}

// GetTestTimes gets test timing data
func (c *HTTPClient) GetTestTimes(ctx context.Context, stepID string, in *types.GetTestTimesReq) (types.GetTestTimesResp, error) {
	var resp types.GetTestTimesResp
	if err := c.validateGetTestTimesArgs(); err != nil {
		return resp, err
	}
	path := fmt.Sprintf(getTestsTimesEndpoint, c.AccountID, c.OrgID, c.ProjectID, c.PipelineID, c.BuildID, c.StageID, stepID)
	backoff := createBackoff(10 * 60 * time.Second)
	_, err := c.retry(ctx, c.Endpoint+path, "POST", "", in, &resp, false, true, backoff) //nolint:bodyclose
	return resp, err
}

// UploadCg uploads avro encoded callgraph to server
func (c *HTTPClient) CommitInfo(ctx context.Context, stepID, branch string) (types.CommitInfoResp, error) {
	var resp types.CommitInfoResp
	if err := c.validateCommitInfoArgs(stepID, branch); err != nil {
		return resp, err
	}
	path := fmt.Sprintf(commitInfoEndpoint, c.AccountID, c.OrgID, c.ProjectID, c.PipelineID, c.BuildID, c.StageID, stepID, c.Repo, branch)
	backoff := createBackoff(5 * 60 * time.Second)
	_, err := c.retry(ctx, c.Endpoint+path, "GET", "", nil, &resp, false, true, backoff) //nolint:bodyclose
	return resp, err
}

// UploadCg uploads avro encoded callgraph to server
func (c *HTTPClient) MLSelectTests(ctx context.Context, stepID, mlKey, source, target string, in *types.MLSelectTestsRequest) (types.SelectTestsResp, error) {
	var resp types.SelectTestsResp
	if err := c.validateMLSelectTestArgs(); err != nil {
		return resp, err
	}
	path := fmt.Sprintf(mlSelectTestsEndpoint, c.AccountID, c.OrgID, c.ProjectID, c.PipelineID, c.BuildID, c.StageID, stepID, c.Repo, c.Sha, source, target, mlKey, c.CommitLink)
	_, err := c.do(ctx, c.Endpoint+path, "POST", "", in, &resp) //nolint:bodyclose
	return resp, err
}

func (c *HTTPClient) Summary(ctx context.Context, summaryRequest types.SummaryRequest) (types.SummaryResponse, error) {
	var resp types.SummaryResponse
	if err := c.validateMLSelectTestArgs(); err != nil {
		return resp, err
	}

	c.SetBasicArguments(&summaryRequest)

	path := fmt.Sprintf(summaryEndpoint, c.AccountID, summaryRequest.OrgID, summaryRequest.ProjectID, summaryRequest.PipelineID, summaryRequest.BuildID, summaryRequest.StageID, summaryRequest.StepID, summaryRequest.ReportType)
	backoff := createBackoff(5 * 60 * time.Second)
	_, err := c.retry(ctx, c.Endpoint+path, "GET", "", nil, &resp, false, true, backoff) //nolint:bodyclose
	return resp, err
}

func (c *HTTPClient) GetTestCases(ctx context.Context, testCasesRequest types.TestCasesRequest) (types.TestCases, error) {
	var resp types.TestCases
	if err := c.validateMLSelectTestArgs(); err != nil {
		return resp, err
	}

	c.SetBasicArguments(&testCasesRequest.BasicInfo)

	path := fmt.Sprintf(testCasesEndpoint, c.AccountID, testCasesRequest.BasicInfo.OrgID, testCasesRequest.BasicInfo.ProjectID, testCasesRequest.BasicInfo.PipelineID, testCasesRequest.BasicInfo.BuildID, testCasesRequest.BasicInfo.StageID, testCasesRequest.BasicInfo.StepID, testCasesRequest.BasicInfo.ReportType, testCasesRequest.TestCaseSearchTerm, testCasesRequest.Sort, testCasesRequest.Order, testCasesRequest.PageIndex, testCasesRequest.PageSize, testCasesRequest.SuiteName)
	backoff := createBackoff(5 * 60 * time.Second)
	_, err := c.retry(ctx, c.Endpoint+path, "GET", "", nil, &resp, false, true, backoff) //nolint:bodyclose
	return resp, err
}

// WriteSavings writes time savings for a step/feature to TI server
func (c *HTTPClient) WriteSavings(ctx context.Context, stepID string, featureName types.SavingsFeature, featureState types.IntelligenceExecutionState, timeTakenMs int64, savingsRequest types.SavingsRequest) error {
	if err := c.validateWriteSavingsArgs(stepID); err != nil {
		return err
	}
	timeTakenMsStr := strconv.Itoa(int(timeTakenMs))
	path := fmt.Sprintf(savingsEndpoint, c.AccountID, c.OrgID, c.ProjectID, c.PipelineID, c.BuildID, c.StageID, stepID, c.Repo, string(featureName), string(featureState), timeTakenMsStr)
	_, err := c.do(ctx, c.Endpoint+path, "POST", "", savingsRequest, nil) //nolint:bodyclose
	return err
}

// Healthz pings the healthz endpoint
func (c *HTTPClient) Healthz(ctx context.Context) error {
	response, err := c.do(ctx, c.Endpoint+healthzEndpoint, "GET", "", nil, nil)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("TI Healthz Ping failed. Status Code:%s", response.Status)
	}
	return nil
}

// DownloadAgent downloads the agent file from remote storage.
func (c *HTTPClient) DownloadAgent(ctx context.Context, path string) (io.ReadCloser, error) {
	resp, err := c.open(ctx, path, "GET", nil)
	return resp.Body, err
}

func (c *HTTPClient) retry(ctx context.Context, method, path, sha string, in, out interface{}, isOpen, retryOnServerErrors bool, b backoff.BackOff) (*http.Response, error) {
	for {
		var res *http.Response
		var err error
		if !isOpen {
			res, err = c.do(ctx, method, path, sha, in, out)
		} else {
			res, err = c.open(ctx, method, path, in.(io.Reader))
		}

		// do not retry on Canceled or DeadlineExceeded
		if err := ctx.Err(); err != nil {
			// Context cancelled
			return res, err
		}

		duration := b.NextBackOff()

		if res != nil {
			// Check the response code. We retry on 5xx-range
			// responses to allow the server time to recover, as
			// 5xx's are typically not permanent errors and may
			// relate to outages on the server side.
			if res.StatusCode >= 500 && retryOnServerErrors {
				// TI server error: Reconnect and retry
				if duration == backoff.Stop {
					return nil, err
				}
				time.Sleep(duration)
				continue
			}
		} else if err != nil {
			// Request error: Retry
			if duration == backoff.Stop {
				return nil, err
			}
			time.Sleep(duration)
			continue
		}
		return res, err
	}
}

// do is a helper function that posts a signed http request with
// the input encoded and response decoded from json.
func (c *HTTPClient) do(ctx context.Context, path, method, sha string, in, out interface{}) (*http.Response, error) { //nolint:unparam
	var r io.Reader

	if in != nil {
		buf := new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(in); err != nil {
			return nil, err
		}
		r = buf
	}

	req, err := http.NewRequestWithContext(ctx, method, path, r)
	if err != nil {
		return nil, err
	}

	// the request should include the secret shared between
	// the agent and server for authorization.
	req.Header.Add("X-Harness-Token", c.Token)
	// adding sha as request-id for logging context
	if sha != "" {
		req.Header.Add("X-Request-ID", sha)
	}
	res, err := c.client().Do(req)
	if res != nil {
		defer func() {
			// drain the response body so we can reuse
			// this connection.
			if _, cerr := io.Copy(io.Discard, io.LimitReader(res.Body, 4096)); cerr != nil {
			}
			res.Body.Close()
		}()
	}
	if err != nil {
		return res, err
	}

	// if the response body return no content we exit
	// immediately. We do not read or unmarshal the response
	// and we do not return an error.
	if res.StatusCode == http.StatusNoContent {
		return res, nil
	}

	// else read the response body into a byte slice.
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return res, err
	}

	if res.StatusCode >= http.StatusMultipleChoices {
		// if the response body includes an error message
		// we should return the error string.
		if len(body) != 0 {
			out := new(Error)
			if err := json.Unmarshal(body, out); err == nil {
				return res, &Error{Code: res.StatusCode, Message: out.Message}
			}
			return res, &Error{Code: res.StatusCode, Message: string(body)}
		}
		// if the response body is empty we should return
		// the default status code text.
		return res, errors.New(
			http.StatusText(res.StatusCode),
		)
	}
	if out == nil {
		return res, nil
	}
	return res, json.Unmarshal(body, out)
}

// client is a helper function that returns the default client
// if a custom client is not defined.
func (c *HTTPClient) client() *http.Client {
	if c.Client == nil {
		return defaultClient
	}
	return c.Client
}

// helper function to open an http request
func (c *HTTPClient) open(ctx context.Context, path, method string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-Harness-Token", c.Token)
	return c.client().Do(req)
}

func createInfiniteBackoff() *backoff.ExponentialBackOff {
	return createBackoff(0)
}

func createBackoff(maxElapsedTime time.Duration) *backoff.ExponentialBackOff {
	exp := backoff.NewExponentialBackOff()
	exp.MaxElapsedTime = maxElapsedTime
	return exp
}

func (c *HTTPClient) validateTiArgs() error {
	if c.Endpoint == "" {
		return fmt.Errorf("ti endpoint is not set")
	}
	if c.Token == "" {
		return fmt.Errorf("ti token is not set")
	}
	return nil
}

func (c *HTTPClient) validateBasicArgs() error {
	if c.AccountID == "" {
		return fmt.Errorf("accountID is not set")
	}
	if c.OrgID == "" {
		return fmt.Errorf("orgID is not set")
	}
	if c.ProjectID == "" {
		return fmt.Errorf("projectID is not set")
	}
	if c.PipelineID == "" {
		return fmt.Errorf("pipelineID is not set")
	}
	return nil
}

func (c *HTTPClient) validateWriteArgs(stepID, report string) error {
	if err := c.validateTiArgs(); err != nil {
		return err
	}
	if err := c.validateBasicArgs(); err != nil {
		return err
	}
	if c.BuildID == "" {
		return fmt.Errorf("buildID is not set")
	}
	if c.StageID == "" {
		return fmt.Errorf("stageID is not set")
	}
	if stepID == "" {
		return fmt.Errorf("stepID is not set")
	}
	if report == "" {
		return fmt.Errorf("report is not set")
	}
	return nil
}

func (c *HTTPClient) validateWriteSavingsArgs(stepID string) error {
	if err := c.validateTiArgs(); err != nil {
		return err
	}
	if err := c.validateBasicArgs(); err != nil {
		return err
	}
	if c.BuildID == "" {
		return fmt.Errorf("buildID is not set")
	}
	if c.StageID == "" {
		return fmt.Errorf("stageID is not set")
	}
	if stepID == "" {
		return fmt.Errorf("stepID is not set")
	}
	return nil
}

func (c *HTTPClient) validateDownloadLinkArgs(language string) error {
	if err := c.validateTiArgs(); err != nil {
		return err
	}
	if language == "" {
		return fmt.Errorf("language is not set")
	}
	return nil
}

func (c *HTTPClient) validateSelectTestsArgs(stepID, source, target string) error {
	if err := c.validateTiArgs(); err != nil {
		return err
	}
	if err := c.validateBasicArgs(); err != nil {
		return err
	}
	if c.BuildID == "" {
		return fmt.Errorf("buildID is not set")
	}
	if c.StageID == "" {
		return fmt.Errorf("stageID is not set")
	}
	if stepID == "" {
		return fmt.Errorf("stepID is not set")
	}
	if source == "" {
		return fmt.Errorf("source branch is not set")
	}
	if target == "" {
		return fmt.Errorf("target branch is not set")
	}
	return nil
}

func (c *HTTPClient) validateUploadCgArgs(stepID, source, target string) error {
	if err := c.validateTiArgs(); err != nil {
		return err
	}
	if err := c.validateBasicArgs(); err != nil {
		return err
	}
	if c.BuildID == "" {
		return fmt.Errorf("buildID is not set")
	}
	if c.StageID == "" {
		return fmt.Errorf("stageID is not set")
	}
	if stepID == "" {
		return fmt.Errorf("stepID is not set")
	}
	if source == "" {
		return fmt.Errorf("source branch is not set")
	}
	if target == "" {
		return fmt.Errorf("target branch is not set")
	}
	return nil
}

func (c *HTTPClient) validateGetTestTimesArgs() error {
	if err := c.validateTiArgs(); err != nil {
		return err
	}
	return c.validateBasicArgs()
}

func (c *HTTPClient) validateCommitInfoArgs(stepID, branch string) error {
	if err := c.validateTiArgs(); err != nil {
		return err
	}
	if err := c.validateBasicArgs(); err != nil {
		return err
	}
	if c.BuildID == "" {
		return fmt.Errorf("buildID is not set")
	}
	if c.StageID == "" {
		return fmt.Errorf("stageID is not set")
	}
	if stepID == "" {
		return fmt.Errorf("stepID is not set")
	}
	if branch == "" {
		return fmt.Errorf("source branch is not set")
	}
	return nil
}

func (c *HTTPClient) validateMLSelectTestArgs() error {
	if err := c.validateTiArgs(); err != nil {
		return err
	}
	return c.validateBasicArgs()
}

func (c *HTTPClient) SetBasicArguments(summaryRequest *types.SummaryRequest) {
	if summaryRequest.OrgID == "" {
		summaryRequest.OrgID = c.OrgID
	}
	if summaryRequest.ProjectID == "" {
		summaryRequest.ProjectID = c.ProjectID
	}
	if summaryRequest.PipelineID == "" {
		summaryRequest.PipelineID = c.PipelineID
	}
	if summaryRequest.BuildID == "" {
		summaryRequest.BuildID = c.BuildID
	}
	if summaryRequest.ReportType == "" {
		summaryRequest.ReportType = "junit"
	}

	if summaryRequest.AllStages {
		summaryRequest.StageID = ""
		summaryRequest.StepID = ""
	}
}
