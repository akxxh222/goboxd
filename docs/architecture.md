# Architecture

## High-level overview

goboxd is a Linux-only code execution service built in Go. It accepts HTTP requests, validates them, writes source files to a temporary directory, and executes the code inside `nsjail` sandboxes. The service currently supports Python 3 and C++.

## Request lifecycle

Client
→ HTTP API
→ Validation
→ Runner
→ Sandbox
→ Language executor
→ Result aggregation
→ Response

1. Client sends a request to `/run`.
2. `internal/httpapi` decodes JSON and applies request validation.
3. `internal/runner` writes source and test data to a temporary directory.
4. The runner invokes `nsjail` with sandbox options.
5. The language executor runs the appropriate command inside the sandbox.
6. Output is captured, truncated if needed, and returned as JSON.
7. The temporary directory is deleted after the request completes.

## Component responsibilities

- `cmd/goboxd`
  - Application entry point.
  - Verifies runtime OS is Linux.
  - Verifies `nsjail` is available in PATH.
  - Cleans stale temporary directories older than 30 minutes.
  - Starts the HTTP server on port `:8080`.

- `internal/httpapi`
  - Exposes HTTP endpoints: `/healthz`, `/readyz`, `/info`, `/run`.
  - Validates HTTP methods.
  - Limits request body size to 1 MiB.
  - Serializes JSON responses.
  - Creates and removes per-request temporary directories.

- `internal/security`
  - Validates required request fields.
  - Enforces source size and test limits.
  - Enforces maximum input and expected output sizes.

- `internal/config`
  - Defines timeouts and limits used by the service.
  - Configures sandbox rlimits for CPU, address space, file size, open files, and processes.

- `internal/types`
  - Defines request and response payloads.
  - Defines build and test result structures.

- `internal/runner`
  - Dispatches language execution based on request language.
  - Writes source files and artifacts in a temp directory.
  - Constructs `nsjail` command line arguments.
  - Captures and truncates process output.
  - Logs sandbox diagnostics separately from returned program stderr.

## Package structure

```
.
├── cmd/
│   └── goboxd/          application entry point
├── internal/
│   ├── config/          runtime and sandbox limits
│   ├── httpapi/         HTTP endpoints and response handling
│   ├── runner/          sandbox command assembly and language execution
│   ├── security/        request validation logic
│   └── types/           API payload definitions
├── docs/               generated documentation
├── tests/              integration tests
├── Dockerfile
├── docker-compose.yml
└── Makefile
```

## Execution flow

- `cmd/goboxd/main.go` starts the server and cleans stale temp dirs.
- `internal/httpapi.NewMux` registers endpoints.
- `POST /run` is handled by `internal/httpapi.run`.
- `run` parses JSON, validates with `internal/security.ValidateRunRequest`, and creates a temp directory with `os.MkdirTemp`.
- `runner.Run` dispatches based on `language`.
- `internal/runner/sandboxedCommand` builds the `nsjail` command line.
- The language-specific runner writes source to a fixed filename and executes inside the sandbox.
- Output is captured using `cappedBuffer` and returned in the response.
- `defer os.RemoveAll(tempDir)` removes the workspace at the end of the request.

## Temp directory lifecycle

- A per-request temporary directory is created under the system temp directory with prefix `goboxd-*`.
- The directory is chmodded to `0755` before execution.
- The service defers removal of the entire directory after each request.
- At startup, `cmd/goboxd` scans the system temp directory and removes stale `goboxd-*` directories older than 30 minutes.

## nsjail integration

- `internal/runner/sandboxedCommand` uses `nsjail` when available and when not running on Windows.
- It passes sandbox options including:
  - `-Mo`
  - `--user 65534`
  - `--group 65534`
  - `--time_limit`
  - `--rlimit_cpu`
  - `--rlimit_as`
  - `--rlimit_fsize`
  - `--rlimit_nofile`
  - `--rlimit_nproc`
  - `--chroot /`
  - `-E PATH`
  - `--cwd` set to the request temp directory
  - `--log` set to `nsjail.log` inside the temp directory
- If `nsjail` is not installed or the environment is Windows, the service falls back to executing commands directly.

## Build phase

- Only C++ submissions have an explicit build phase.
- The C++ build command is:
  - `g++ solution.cpp -o solution`
- The build runs inside `nsjail` with larger sandbox limits for memory, file size, and open files.
- Build metadata is returned in the `build` field of the `/run` response.

## Execution phase

- Python execution runs:
  - `python3 solution.py`
- C++ execution runs:
  - `./solution`
- Each test case is executed separately in its own sandbox invocation.
- Standard input is provided from `tests[].stdin`.
- Output is compared to `tests[].expected_stdout` using trimmed whitespace equality.

## Cleanup phase

- After each request, the temporary directory is removed.
- The `nsjail.log` file is read and logged by the service before removal.
- Stale temp directories older than 30 minutes are removed on startup.
