# AGENTS.md - Context for AI Coding Assistants

This document provides essential context about the `ti-client` codebase to help AI coding assistants (like Cursor, Windsurf, etc.) understand the project structure, patterns, and conventions.

## Project Purpose

`ti-client` is a **Go client library** for the Harness Test Intelligence (TI) service. It provides a clean, type-safe interface for Go applications to communicate with the TI service backend.

**Key Context:**
- This is a **library/package**, not a standalone application
- It's imported by other Harness repositories: `harness-core` and `lite-engine`
- The TI service backend is in a separate repository: `harness-ti`
- The client enables CI/CD pipelines to use test intelligence features

## Core Concepts

### Test Intelligence (TI)
Test Intelligence automatically selects which tests to run based on:
- **Code changes**: Tests that depend on modified source files
- **New tests**: Tests introduced in the current branch/PR
- **Updated tests**: Existing tests that were modified
- **Previous failures**: Tests that failed in previous builds
- **Flaky tests**: Tests with inconsistent results

### Callgraph
A **callgraph** represents code dependencies - which source files are called by which tests. This is used to determine which tests need to run when source code changes.

- **V1 API**: Uses Avro-encoded binary format (`UploadCg`)
- **V2 API (Chrysalis)**: Uses JSON format (`UploadCgV2`) - newer, preferred approach

### Test Selection Flow
1. Client sends changed files and branch info to TI service
2. TI service analyzes callgraph and returns list of tests to run
3. Client receives `SelectTestsResp` with `RunnableTest` list
4. Tests are executed
5. Results are written back via `Write()`
6. Callgraph is uploaded via `UploadCg()` or `UploadCgV2()`

## Architecture Patterns

### Interface-Based Design
The client uses Go interfaces for abstraction:
- `client.Client` interface defines all operations
- `client.HTTPClient` implements the interface
- This allows for testing and potential alternative implementations

### Error Handling
- Custom `client.Error` type with Code and Message
- Errors are returned, not panicked
- HTTP errors are wrapped in `client.Error` with status codes

### Retry Logic
All HTTP operations use exponential backoff retry:
- Network errors: Retry with backoff
- 5xx server errors: Retry (configurable per operation)
- 4xx client errors: Not retried
- Different operations have different max retry times

### Validation Pattern
Each operation has a corresponding `validate*Args()` function:
- `validateWriteArgs()` for `Write()`
- `validateSelectTestsArgs()` for `SelectTests()`
- `validateUploadCgArgs()` for `UploadCg()`
- etc.

## Package Structure

```
ti-client/
├── client/                   # Core client implementation
│   ├── client.go             # Client interface (all methods)
│   └── http.go               # HTTPClient implementation
│
├── types/                    # Core type definitions
│   ├── types.go              # Main types (TestCase, SelectTestsReq/Resp, etc.)
│   ├── savings.go            # Savings tracking types
│   └── cache/                # Cache-related types
│       ├── buildcache/
│       ├── dlc/
│       ├── gradle/
│       └── maven/
│
├── chrysalis/                # V2 API types (newer API)
│   └── types/
│       ├── types.go          # UploadCgRequest, SkipTestsRequest
│       ├── chain.go          # Chain type (code dependency)
│       ├── test.go           # Test type
│       └── identifier.go     # Identifier type
│
└── clientUtils/               # Utility functions
    └── telemetryUtils/        # Telemetry helpers
```

## Key Files and Their Roles

### `client/client.go`
- Defines the `Client` interface
- All public methods that consumers use
- Custom `Error` type

### `client/http.go`
- `HTTPClient` struct: Implements `Client` interface
- `NewHTTPClient()`: Factory function to create client
- HTTP request/response handling
- Retry logic with exponential backoff
- TLS/mTLS configuration
- Validation functions

### `types/types.go`
- Core data structures:
  - `TestCase`: Test execution result
  - `RunnableTest`: Test to be executed
  - `SelectTestsReq/Resp`: Test selection request/response
  - `File`: Changed file information
  - `Status`, `FileStatus`, `Selection`: Enums/constants
- Environment variable constants
- Telemetry data structures

### `types/savings.go`
- Savings tracking for intelligence features
- `SavingsFeature`: BUILD_CACHE, TI, DLC
- `IntelligenceExecutionState`: FULL_RUN, OPTIMIZED, DISABLED

### `chrysalis/types/types.go`
- V2 API request types
- `UploadCgRequest`: JSON callgraph upload
- `SkipTestsRequest`: File checksum-based skip logic

## Code Conventions

### Naming
- **Interfaces**: Single word (e.g., `Client`)
- **Implementations**: Descriptive (e.g., `HTTPClient`)
- **Methods**: PascalCase, descriptive verbs (e.g., `SelectTests`, `UploadCg`)
- **Constants**: SCREAMING_SNAKE_CASE (e.g., `StatusPassed`, `FileModified`)

### Function Signatures
- Always takes `context.Context` as first parameter
- Returns `error` as last return value
- Request/response types are in `types` package
- V2 types are in `chrysalis/types` package

### HTTP Patterns
- Endpoints are defined as `const` strings with format specifiers
- Path construction uses `fmt.Sprintf()` with endpoint template
- Headers:
  - `X-Harness-Token`: Authentication
  - `X-Request-ID`: Request tracking (usually SHA)
- Request body: JSON-encoded
- Response body: JSON-decoded into response struct

### Error Patterns
```go
// Custom error with code
return &Error{Code: res.StatusCode, Message: out.Message}

// Validation errors
return fmt.Errorf("stepID is not set")

// Context errors are not retried
if err := ctx.Err(); err != nil {
    return res, err
}
```

## Common Tasks

### Adding a New API Endpoint

1. **Add method to interface** (`client/client.go`):
   ```go
   // NewMethod does something
   NewMethod(ctx context.Context, param string) (Response, error)
   ```

2. **Add endpoint constant** (`client/http.go`):
   ```go
   newEndpoint = "/api/new?accountId=%s&param=%s"
   ```

3. **Implement method** (`client/http.go`):
   ```go
   func (c *HTTPClient) NewMethod(ctx context.Context, param string) (Response, error) {
       if err := c.validateNewMethodArgs(param); err != nil {
           return Response{}, err
       }
       path := fmt.Sprintf(newEndpoint, c.AccountID, param)
       backoff := createBackoff(5 * 60 * time.Second)
       var resp Response
       _, err := c.retry(ctx, c.Endpoint+path, "GET", "", nil, &resp, false, true, backoff)
       return resp, err
   }
   ```

4. **Add validation function**:
   ```go
   func (c *HTTPClient) validateNewMethodArgs(param string) error {
       if err := c.validateTiArgs(); err != nil {
           return err
       }
       if param == "" {
           return fmt.Errorf("param is not set")
       }
       return nil
   }
   ```

5. **Add types** (`types/types.go` or new file):
   ```go
   type Response struct {
       Field string `json:"field"`
   }
   ```

### Modifying Existing Endpoints

- **Endpoint URL changes**: Update the endpoint constant
- **Request/response changes**: Update types in `types/` package
- **New parameters**: Add to function signature, update validation
- **Retry behavior**: Adjust `createBackoff()` timeout

### Adding New Types

- **Core types**: Add to `types/types.go` or create new file in `types/`
- **V2 API types**: Add to `chrysalis/types/`
- **Cache types**: Add to `types/cache/{category}/`
- Use JSON tags for serialization
- Use BSON tags for MongoDB (in chrysalis types)

## Security Considerations

### mTLS Support
- Certificates can be provided as:
  - Base64-encoded strings (preferred for containers)
  - File paths (default: `/etc/mtls/client.crt`, `/etc/mtls/client.key`)
- mTLS is optional - only enabled if certificates are provided

### TLS Configuration
- `SkipVerify`: For development/testing only
- Custom root CAs: Loaded from directory
- System cert pool: Used as base, additional certs appended

### Authentication
- Token-based: `X-Harness-Token` header
- Token is required for all operations
- No token refresh logic (assumed to be valid for request lifetime)

## Testing Patterns

When adding tests (if test files exist):
- Use table-driven tests for multiple scenarios
- Mock HTTP client for unit tests
- Test validation functions separately
- Test retry logic with mock servers

## Dependencies

### External
- `github.com/cenkalti/backoff`: Exponential backoff for retries
- `github.com/cespare/xxhash/v2`: Hashing (used internally)
- `go.mongodb.org/mongo-driver`: BSON types for chrysalis (MongoDB ObjectID)

### Internal
- All types are in `types/` or `chrysalis/types/`
- No circular dependencies
- `client` depends on `types`, not vice versa

## Common Workflows

### Test Selection Workflow
1. Get changed files from git/VCS
2. Create `SelectTestsReq` with files and branch info
3. Call `SelectTests()` or `MLSelectTests()`
4. Receive `SelectTestsResp` with tests to run
5. Execute tests
6. Write results via `Write()`
7. Upload callgraph via `UploadCg()` or `UploadCgV2()`

### Callgraph Upload Workflow
1. Generate callgraph (done by TI agent, not this client)
2. Encode as Avro (V1) or JSON (V2)
3. Call `UploadCg()` or `UploadCgV2()`
4. Service stores callgraph for future test selection

### Savings Tracking Workflow
1. Track time taken and time saved for intelligence features
2. Collect metrics (Gradle, Maven, DLC)
3. Call `WriteSavings()` with metrics
4. Service aggregates and reports savings

## Important Notes

1. **No State Management**: Client is stateless - each request is independent
2. **Context Usage**: Always respect context cancellation/timeout
3. **Retry Behavior**: Different operations have different retry timeouts based on expected duration
4. **V2 API**: Chrysalis (V2) is the newer API - prefer when possible
5. **Backward Compatibility**: V1 API still supported for existing integrations
6. **Error Handling**: Always check errors, don't ignore them
7. **Validation**: All public methods validate inputs before making requests

## Extension Points

### Adding New Intelligence Features
1. Add feature constant to `types.SavingsFeature`
2. Add metrics type if needed (e.g., in `types/cache/`)
3. Update `SavingsRequest` if new metrics needed
4. Add endpoint/method if new API needed

### Supporting New Test Frameworks
- No client changes needed - framework-agnostic
- TI service handles framework-specific logic
- Client just passes language/framework info

### Adding New Telemetry
- Add fields to `TelemetryData` in `types/types.go`
- Use `telemetryUtils` for common calculations
- Write telemetry via existing `Write()` method

## Related Codebases

- **harness-ti**: TI service backend (API server)
- **harness-core**: Uses this client in pipeline execution
- **lite-engine**: Uses this client to communicate with TI service

When making changes, consider:
- Backward compatibility with existing consumers
- API contract with `harness-ti` service
- Impact on `lite-engine` and `harness-core` integrations

