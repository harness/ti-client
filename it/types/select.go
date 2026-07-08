// Copyright 2026 Harness Inc. All rights reserved.
// Use of this source code is governed by the PolyForm Free Trial 1.0.0 license
// that can be found in the licenses directory at the root of this repository, also available at
// https://polyformproject.org/wp-content/uploads/2020/05/PolyForm-Free-Trial-1.0.0.txt.

package types

// SelectITRequest is the body of POST /it/select. Platform identifiers
// (accountId) ride as URL query params, like /it/uploadcg.
//
// hcli sends:
//   - the test repo it's about to run (URL + the local test files with
//     their content checksums), and
//   - the set of deployed services it discovered, each identified by
//     (repo_url, repo_uuid).
//
// ti-service fetches each repo_uuid's build manifest from GCS, extracts the
// per-file content checksums, encodes every source path as
// "#{repo_url}#{file_path}", merges those with the local test-file
// checksums into one map, and runs selection against it_chains. The
// encoding must match what the stitching consumer wrote, so repo_url — not
// repo_uuid — is the stable key.
type SelectITRequest struct {
	// TestRepoURL identifies the IT test repo (the selection identity's
	// repo). Matches UploadITGraphRequest.TestRepoURL.
	TestRepoURL string `json:"test_repo_url"`

	// Files maps a local test-repo file path to its content checksum
	// (hcli-computed). These are the test files themselves; cross-repo
	// source checksums are resolved server-side from manifests.
	Files map[string]string `json:"files"`

	// Deployment is the set of services hcli discovered, each anchoring a
	// build manifest. ServiceName is intentionally absent: it is sometimes
	// unavailable and selection joins by repo, not by name.
	Deployment []DeployedRepo `json:"deployment"`

	// ExecutionContext scopes the selection identity (matches the identity's
	// extraInfo). Optional; empty means the account/repo default identity.
	ExecutionContext map[string]string `json:"execution_context,omitempty"`
}

// DeployedRepo identifies one repo build backing a deployed service.
//
// RepoURL is the stable cross-build identity used to encode source paths.
// RepoUUID anchors the specific build manifest in GCS
// (manifests/{accountId}/{repo_uuid}/manifest.json.gz) whose checksums are
// current for this deployment. Both are required: the UUID finds the
// manifest, the URL keys the encoding.
type DeployedRepo struct {
	RepoURL  string `json:"repo_url"`
	RepoUUID string `json:"repo_uuid"`
}
