// Copyright 2026 Harness Inc. All rights reserved.
// Use of this source code is governed by the PolyForm Free Trial 1.0.0 license
// that can be found in the licenses directory at the root of this repository, also available at
// https://polyformproject.org/wp-content/uploads/2020/05/PolyForm-Free-Trial-1.0.0.txt.

// Package types defines the wire format for integration test (IT) call graph
// uploads.
//
// The upload is a tests-first JSON array. Each top-level entry is a service
// block listing the tests that touched that service and the source files
// they reached. The chain-stitching consumer uses service.uuid to fetch the
// build manifest from GCS and stamp per-source content checksums onto the
// stored chain (V2-equivalent reference for selection-time comparison).
//
// Platform identifiers (accountId, orgId, projectId, parentUniqueId,
// uniqueId) ride as URL query params on POST /it/uploadcg, not in the body.
package types

// UploadITGraphRequest is the top-level upload payload — a flat array of
// service blocks.
//
// Each block describes one service that the test run touched. Sources within
// a block belong to that service (the file_path is relative to that
// service's repo). No recursion or call-hierarchy is encoded; selection only
// needs the union of (service, file) touches per test.
type UploadITGraphRequest []ServiceBlock

// ServiceBlock identifies one service and the tests that exercised it.
//
// host + port is the canonical service address — the load-balancer DNS hcli
// uses for discovery. Pod IPs and internal aliases must be normalized to
// this canonical form by hcli before upload so the chain's stored address
// matches discovery's address at selection time.
//
// uuid is the build-phase anchor: TI service uses it to fetch the build
// manifest from GCS at chain-write time and stamp per-source content
// checksums onto the chain.
//
// service_name is optional human-readable metadata; selection joins by
// (host, port), not by name.
type ServiceBlock struct {
	Service Service `json:"service"`
	Tests   []Test  `json:"tests"`
}

// Service identifies a deployed service in the test environment.
type Service struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	UUID        string `json:"uuid"`
	ServiceName string `json:"service_name,omitempty"`
}

// Test is one test case that ran against (and touched code in) the parent
// service block.
//
// test_repo_url + test_file_path identify the test (the test repo is
// typically a separate IT repo, not a deployed service). test_checksum is
// xxhash64 of the test file's bytes, computed by hcli at upload time. There
// is no build manifest for the test repo, so the upload is the only path
// for test_checksum to reach storage. Selection compares the chain's stored
// test_checksum against a freshly-computed checksum to detect test-side
// changes.
type Test struct {
	TestRepoURL  string   `json:"test_repo_url"`
	TestFilePath string   `json:"test_file_path"`
	TestChecksum string   `json:"test_checksum"`
	Sources      []Source `json:"sources"`
}

// Source is one source file/class/method that the test exercised within the
// parent service.
//
// file_path is required and is the path relative to the service's repo
// root. class (FQCN) and method are optional debug/UI metadata; selection
// matches by file_path.
//
// Source-file content checksums are NOT carried here. They live in the
// build manifest in GCS, indexed by service.uuid; the chain-stitching
// consumer fetches them at write time. Keeping sources content-free keeps
// the upload payload small and makes the build manifest the single source
// of truth for file fingerprints.
type Source struct {
	FilePath string `json:"file_path"`
	Class    string `json:"class,omitempty"`
	Method   string `json:"method,omitempty"`
}

// UploadITGraphResponse is the body returned on successful 200/202 from
// POST /it/uploadcg.
type UploadITGraphResponse struct {
	Status string `json:"status"`
	Path   string `json:"path"`
}
