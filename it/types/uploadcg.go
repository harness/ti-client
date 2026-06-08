// Copyright 2026 Harness Inc. All rights reserved.
// Use of this source code is governed by the PolyForm Free Trial 1.0.0 license
// that can be found in the licenses directory at the root of this repository, also available at
// https://polyformproject.org/wp-content/uploads/2020/05/PolyForm-Free-Trial-1.0.0.txt.

// Package types defines the wire format for integration test (IT) call graph
// uploads. The IT graph is a multi-service nested structure produced by hcli's
// collection phase after a cross-service test run completes.
package types

// CollectionStatus indicates the outcome of collecting graph data from a service.
type CollectionStatus string

const (
	CollectionStatusSuccess CollectionStatus = "success"
	CollectionStatusPartial CollectionStatus = "partial"
	CollectionStatusFailed  CollectionStatus = "failed"
)

// MaxRecursionDepth is the hard cap on nested downstream depth accepted by the
// upload endpoint. Trees deeper than this are rejected.
const MaxRecursionDepth = 10

// UploadITGraphRequest is the top-level payload for POST /it/uploadcg.
//
// Platform identifiers (accountId, orgId, projectId, parentUniqueId, uniqueId)
// are sent as URL query params, matching the V2 contract. Only the graph body
// lives in the JSON payload.
type UploadITGraphRequest struct {
	// ExecutionID is agent-generated; carried for trace correlation only.
	// Not part of the storage doc key.
	ExecutionID string `json:"execution_id"`

	// Service identifies the service that produced this graph block.
	Service ServiceBlock `json:"service"`

	// Entries are per-test-case tracked-source lists for this service.
	Entries []Entry `json:"entries"`

	// Downstream are graphs collected from services this service called.
	Downstream []UploadITGraphRequest `json:"downstream,omitempty"`

	// CollectionStatus indicates whether collection from this service (and its
	// downstream) succeeded.
	CollectionStatus CollectionStatus `json:"collection_status"`
}

// ServiceBlock identifies a service in the deployed TestEnv.
//
// UUID is the build-phase anchor: TI service uses it to look up
// (repo, commitSHA, source files, class mappings) registered when the
// artifact was built.
//
// Name is the human-readable handle used as a join key in storage.
//
// Address is debug-only metadata (not used for joining or selection).
type ServiceBlock struct {
	UUID    string `json:"uuid"`
	Name    string `json:"name"`
	Address string `json:"address,omitempty"`
}

// Entry represents one per-test-case tracked execution within a service.
type Entry struct {
	// ContextID is the per-test-case ID propagated via X-TI-Context-ID.
	// Used to join graph segments across services for the same test.
	ContextID string `json:"context_id"`

	// Root is the entry-point handler/servlet method that received the request.
	Root Node `json:"root"`

	// Nodes are all source files/classes/methods touched while processing
	// the request keyed by ContextID.
	Nodes []Node `json:"nodes"`
}

// Node is one source location touched during request processing.
//
// File is required. Class and Method are optional — not all languages or
// frameworks expose method-level granularity (e.g. config classes,
// model/POJO classes, scripting languages).
type Node struct {
	File   string `json:"file"`
	Class  string `json:"class,omitempty"`
	Method string `json:"method,omitempty"`
}

// UploadITGraphResponse is the body returned on successful 202 Accepted.
type UploadITGraphResponse struct {
	ParentUniqueID   string      `json:"parentUniqueId"`
	UniqueID         string      `json:"uniqueId"`
	ServicesAccepted int         `json:"services_accepted"`
	Warnings         []ITWarning `json:"warnings,omitempty"`
}

// ITWarning is a per-service validation issue that did not cause the upload to
// be rejected (e.g. one service had malformed nodes, others were fine).
type ITWarning struct {
	Service string `json:"service"`
	Msg     string `json:"msg"`
}
