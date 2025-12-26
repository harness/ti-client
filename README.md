# ti-client

A Go client library for communicating with the Harness Test Intelligence (TI) service. This client is designed to be imported and used by other Harness codebases such as `harness-core` and `lite-engine`, where binaries like `lite-engine` communicate with the `ti-service` backend.

## Overview

The `ti-client` provides a comprehensive interface for interacting with the TI service, enabling test intelligence features such as:

- **Intelligent Test Selection**: Automatically select tests to run based on code changes
- **Callgraph Management**: Upload and manage code callgraphs to track test dependencies
- **Test Result Reporting**: Write test execution results back to the TI service
- **ML-Based Test Selection**: Leverage machine learning models for optimized test selection
- **Time Savings Tracking**: Track and report time savings from various intelligence features
- **Agent Management**: Download test intelligence agent artifacts

## Architecture

The TI service (located in the `harness-ti` repository) provides the backend API, while this client library enables Go applications to interact with it. The client is typically used by:

- **lite-engine**: The execution engine that runs CI/CD steps
- **harness-core**: Core Harness platform components

## Package Structure

```
ti-client/
├── client/              # HTTP client implementation
│   ├── client.go        # Client interface definition
│   └── http.go          # HTTP client implementation with retry logic, mTLS support
├── types/               # Core type definitions
│   ├── types.go         # Test cases, selection requests/responses, telemetry
│   ├── savings.go       # Savings tracking types (build cache, TI, DLC)
│   └── cache/           # Cache-related types (buildcache, dlc, gradle, maven)
├── chrysalis/           # V2 API types (newer API version)
│   └── types/           # V2 types for uploadcg, skip tests, etc.
└── clientUtils/         # Utility functions
    └── telemetryUtils/  # Telemetry helper functions
```

## Key Features

### Test Intelligence

- **SelectTests**: Intelligently select tests to run based on source code changes, new tests, updated tests, and previous failures
- **MLSelectTests**: Use machine learning models for advanced test selection
- **GetSkipTests**: Determine which tests can be skipped based on file checksums

### Callgraph Management

- **UploadCg**: Upload Avro-encoded callgraphs to track code dependencies
- **UploadCgV2**: Upload JSON-encoded callgraphs using the newer V2 API
- **UploadCgFailedTest**: Upload callgraphs for failed tests without updating last successful commit

### Test Reporting

- **Write**: Write test execution results to the TI service
- **GetTestCases**: Retrieve test cases executed in a build with pagination and filtering
- **Summary**: Get test execution summary (total, passed, failed, skipped tests)
- **GetTestTimes**: Retrieve historical test timing data

### Savings Tracking

- **WriteSavings**: Track time savings from various intelligence features:
  - Build Cache
  - Test Intelligence
  - Docker Layer Caching (DLC)

### Additional Features

- **DownloadLink**: Get download links for TI agent artifacts
- **CommitInfo**: Get commit information for the last successful commit with callgraph
- **Healthz**: Health check endpoint for service availability
- **DownloadAgent**: Download agent files from remote storage


### Security Features

The client supports several security configurations:

- **mTLS (Mutual TLS)**: Client certificate authentication
  - Supports base64-encoded certificates or file paths
  - Default paths: `/etc/mtls/client.crt` and `/etc/mtls/client.key`

- **Custom Root CAs**: Load additional root certificates from a directory

- **TLS Verification**: Configurable TLS verification (useful for development/testing)

## Retry Logic

The client includes built-in retry logic with exponential backoff for:

- Network errors
- Server errors (5xx status codes)
- Configurable maximum elapsed time per operation

Different operations have different retry timeouts:
- Test selection: 10 minutes
- Callgraph upload: 45 minutes
- Test result writing: 10 minutes
- Agent download: 5 minutes

## Environment Variables

The client expects the following environment variables (defined in `types/types.go`):

- `HARNESS_ACCOUNT_ID`: Harness account ID
- `HARNESS_ORG_ID`: Organization ID
- `HARNESS_PROJECT_ID`: Project ID
- `HARNESS_PIPELINE_ID`: Pipeline ID
- `HARNESS_STAGE_ID`: Stage ID
- `HARNESS_STEP_ID`: Step ID
- `HARNESS_BUILD_ID`: Build ID
- `HARNESS_TI_SERVICE_ENDPOINT`: TI service endpoint URL
- `HARNESS_TI_SERVICE_TOKEN`: Authentication token for TI service

## Dependencies

- `github.com/cenkalti/backoff`: Exponential backoff for retries
- `github.com/cespare/xxhash/v2`: Hashing utilities
- `go.mongodb.org/mongo-driver`: MongoDB driver (for BSON types in chrysalis)

## Related Repositories

- **harness-ti**: The TI service backend that this client communicates with
- **harness-core**: Core Harness platform (uses this client)
- **lite-engine**: Execution engine (uses this client)