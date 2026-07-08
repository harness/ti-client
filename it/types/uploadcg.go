// Copyright 2026 Harness Inc. All rights reserved.
// Use of this source code is governed by the PolyForm Free Trial 1.0.0 license
// that can be found in the licenses directory at the root of this repository, also available at
// https://polyformproject.org/wp-content/uploads/2020/05/PolyForm-Free-Trial-1.0.0.txt.

// Package types defines the wire format for integration test (IT) call graph
// uploads.
//
// The upload is a JSON object with the IT test repo URL hoisted to the top
// level (invariant within one run) and a `graph` array of service blocks.
// Each block lists the tests that touched that service and the sources they
// reached. Each source carries its own repo_uuid that anchors back to a
// per-build manifest in GCS — a deployed service typically contains classes
// from multiple jars, and each jar is built from a different repo, so the
// build identity is per-source, not per-service.
//
// The chain-stitching consumer (separate ticket) reads each source's
// repo_uuid, fetches the matching manifest, stamps per-source content
// checksums onto the stored chain (V2-equivalent reference for
// selection-time comparison).
//
// Platform identifiers (accountId, cgId) ride as URL query params on
// POST /it/uploadcg, not in the body.
//
// The body is gzipped on the wire and stored verbatim in GCS at
// it_callgraphs/{accountId}/{cgId}/callgraph.json.gz. ti-service does not
// decompress or parse the body; the consumer does that.
package types

// UploadITGraphRequest is the top-level upload payload.
//
// test_repo_url is hoisted because it is invariant within a single IT run
// (one hcli invocation runs tests from one IT repo checkout). Hoisting
// avoids repeating the URL on every test entry.
type UploadITGraphRequest struct {
	// TestRepoURL is the URL of the IT test repo. Applies to every test in
	// `Graph[*].Tests[*]`.
	TestRepoURL string `json:"test_repo_url"`

	// Graph is the list of service blocks contributing to this run.
	Graph []ServiceBlock `json:"graph"`
}

// ServiceBlock identifies one service and the tests that exercised it.
//
// host + port is the canonical service address — the load-balancer DNS hcli
// uses for discovery. Pod IPs and internal aliases must be normalized to
// this canonical form by hcli before upload so the chain's stored address
// matches discovery's address at selection time.
type ServiceBlock struct {
	Service Service `json:"service"`
	Tests   []Test  `json:"tests"`
}

// Service identifies a deployed service in the test environment.
//
// This block carries only the deployed-instance identity (host, port,
// optional human-readable name). Build identity (RepoUUID) lives per-source
// because a single deployed service typically loads classes from many jars
// — each jar built from a different repo — and each jar has its own build
// manifest. See Source.RepoUUID.
//
// ServiceName is optional human-readable metadata; selection joins by
// (Host, Port), not by name.
type Service struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	ServiceName string `json:"service_name,omitempty"`
}

// Test is one test case that ran against (and touched code in) the parent
// service block.
//
// TestFilePath is the path of the test source file relative to the
// top-level TestRepoURL. TestChecksum is xxhash64 of the test file's
// bytes, computed by hcli at upload time. There is no build manifest for
// the test repo, so the upload is the only path for the checksum to reach
// storage. Selection compares the chain's stored TestChecksum against a
// freshly-computed checksum to detect test-side changes.
//
// State is the test's outcome in this run (SUCCESS / FAILURE / FLAKY /
// UNKNOWN — the same vocabulary as the chrysalis chain state). The
// chain-stitching consumer stamps it onto every chain row it writes for
// this test, so selection can bucket a matched chain into skip vs. must-run
// exactly like V2: a stored chain whose state is FAILURE forces a re-run
// even when nothing changed; SUCCESS (and other non-FAILURE states) allow a
// skip when the recomputed chain checksum still matches. Empty is treated
// as UNKNOWN by the consumer.
type Test struct {
	TestFilePath string   `json:"test_file_path"`
	TestChecksum string   `json:"test_checksum"`
	State        string   `json:"state,omitempty"`
	Sources      []Source `json:"sources"`
}

// Source is one source file/class/method that the test exercised within
// the parent service.
//
// RepoUUID is the build-phase anchor identifying the repo+commit that THIS
// SOURCE was built from. Lives on the source (not the service) because a
// deployed service typically contains classes from multiple jars, each
// built from a different repo (the service's own code, internal shared
// libraries, etc.). The consumer fetches the manifest at
// manifests/{accountId}/{RepoUUID}/manifest.json.gz to resolve file paths
// and content checksums for this source.
//
// JVM agents emit Jar + Class (FQCN) — that pair joins directly with the
// build manifest's class_to_file map (keyed by "{jar}:{FQCN}") so the
// consumer can resolve to a file path and look up the file's content
// checksum. Non-JVM agents emit FilePath directly. At least one of
// (Jar AND Class) or FilePath must be present per source; both is fine
// and recommended for JVM where both are reliably available.
//
// Third-party / non-instrumented jars (no RepoUUID available) should be
// dropped by the agent — they have no manifest and contribute nothing to
// selection.
//
// Source-file content checksums are NOT carried here. They live in the
// build manifest in GCS, indexed by RepoUUID; the chain-stitching consumer
// fetches them at write time. Keeping sources content-free keeps the
// upload small and makes the build manifest the single source of truth
// for file fingerprints.
type Source struct {
	RepoUUID string `json:"repo_uuid"`
	Jar      string `json:"jar,omitempty"`
	Class    string `json:"class,omitempty"`
	Method   string `json:"method,omitempty"`
	FilePath string `json:"file_path,omitempty"`
}
