# Security Audit

This document reviews the security controls implemented in goboxd and maps them against the security issues listed in the hackathon specification.

## Security Summary

| Security Issue                 | Status    |
| ------------------------------ | --------- |
| Path traversal via filename    | Fixed     |
| Shell-style directory commands | Fixed     |
| Compiler flag injection        | Fixed     |
| Request size limits            | Fixed     |
| UID collisions under load      | Mitigated |
| Unbounded child output         | Fixed     |
| Stale jail directories         | Fixed     |

The service executes untrusted code inside `nsjail` sandboxes, applies request validation before execution, limits captured output, isolates execution environments, and automatically cleans up temporary workspaces.

---

# Path Traversal via Filename

## Threat Description

If user-controlled filenames are written directly to disk, an attacker can attempt directory traversal using paths such as:

```text
../../etc/passwd
```

or absolute paths to escape the intended workspace and access host files.

## Status

**Fixed**

## Implementation Location

* `internal/types/types.go`
* `internal/httpapi/handlers.go`
* `internal/runner/python.go`
* `internal/runner/cpp.go`

## Mitigation

The API does not accept user-controlled filenames.

Source files are written using fixed server-controlled names:

* `solution.py`
* `solution.cpp`

Each execution receives a dedicated temporary workspace created by the server.

Because the filename is never supplied by the client, directory traversal through filenames is not possible.

## Remaining Limitations

None identified in the current implementation.

---

# Shell-Style Directory Commands

## Threat Description

Constructing shell commands with user-controlled input can lead to command injection.

Examples include:

```sh
rm -rf $USER_INPUT
mkdir $USER_INPUT
```

If user input is interpolated into shell commands, arbitrary command execution becomes possible.

## Status

**Fixed**

## Implementation Location

* `internal/runner/python.go`
* `internal/runner/cpp.go`
* `internal/runner/sandbox.go`

## Mitigation

The service does not invoke a shell.

Process execution uses:

```go
exec.CommandContext(...)
```

with command arguments passed separately.

Filesystem operations are performed through Go's standard library APIs instead of shell utilities.

## Remaining Limitations

The implementation assumes trusted installations of:

* `nsjail`
* `python3`
* `g++`

Compromise of those binaries is outside the application's threat model.

---

# Compiler Flag Injection

## Threat Description

Allowing users to supply arbitrary compiler flags can introduce dangerous behavior.

Examples include:

```text
-B
-fplugin
-Wl,...
@response_file
--specs
```

These options may trigger arbitrary code execution during compilation or alter toolchain behavior.

## Status

**Fixed**

## Implementation Location

* `internal/runner/cpp.go`
* `internal/runner/sandbox.go`

## Mitigation

Compilation commands are fully controlled by the server.

The build command is fixed and does not accept compiler arguments from requests.

Example:

```text
g++ solution.cpp -o solution
```

Requests cannot modify compiler options.

## Remaining Limitations

If user-configurable compiler flags are added in future stages, a strict allow-list must be implemented.

---

# Request Size Limits

## Threat Description

Unbounded request sizes can cause excessive memory consumption, CPU usage, or denial-of-service conditions.

Affected fields include:

* source code
* test count
* stdin
* expected output
* request body size

## Status

**Fixed**

## Implementation Location

* `internal/httpapi/handlers.go`
* `internal/security/validation.go`
* `internal/config/config.go`

## Mitigation

The service enforces multiple limits:

| Resource                 | Limit   |
| ------------------------ | ------- |
| Request body             | 1 MiB   |
| Source code              | 256 KiB |
| Test cases               | 25      |
| stdin per test           | 64 KiB  |
| expected_stdout per test | 64 KiB  |

Requests exceeding these limits are rejected before execution.

## Remaining Limitations

The current API only accepts JSON requests.

Any future binary-upload functionality would require separate validation controls.

---

# UID Collisions Under Load

## Threat Description

Shared execution identities can increase the risk of cross-request interference when multiple sandboxes execute simultaneously.

The reference implementation used a small UID allocation pool, creating potential collisions.

## Status

**Mitigated**

## Implementation Location

* `internal/runner/sandbox.go`
* Temporary workspace creation logic

## Mitigation

The implementation does not allocate host-side user accounts.

Instead, isolation is provided through:

* nsjail namespaces
* per-request temporary workspaces
* process isolation
* filesystem isolation

Each request receives an independent workspace and execution environment.

No UID allocation pool exists.

## Remaining Limitations

The implementation relies on namespace isolation rather than unique host-side user identities.

This is sufficient for the current threat model but differs from systems that assign dedicated host UIDs per execution.

---

# Unbounded Child Output

## Threat Description

A malicious program can continuously write to stdout or stderr and exhaust host memory if output is buffered without limits.

## Status

**Fixed**

## Implementation Location

* `internal/runner/runner.go`
* `internal/runner/python.go`
* `internal/runner/cpp.go`

## Mitigation

Captured output is stored in a bounded buffer.

Current limit:

```text
64 KiB
```

When the limit is reached:

* additional output is discarded
* truncation is recorded in the response

This prevents excessive memory growth caused by child processes.

## Remaining Limitations

Output generation by the child process is not prevented; only captured output is bounded.

Resource limits and sandbox controls provide additional protection.

---

# Stale Jail Directories

## Threat Description

Unexpected crashes or incomplete cleanup can leave temporary directories behind.

Accumulated workspaces can:

* consume disk space
* retain source code
* expose execution artifacts

## Status

**Fixed**

## Implementation Location

* `cmd/goboxd/main.go`
* `internal/runner/cleanup.go`

## Mitigation

Each request creates a dedicated temporary workspace.

Cleanup occurs through:

```go
defer os.RemoveAll(...)
```

ensuring removal on normal request completion.

Additionally, startup cleanup removes stale directories matching:

```text
goboxd-*
```

older than 30 minutes.

## Remaining Limitations

Orphaned directories may remain until the next service restart.

The startup sweep removes them automatically.

---

# Sandbox Logging

## Status

**Implemented**

## Description

Sandbox diagnostic output is separated from user-visible execution results.

The service no longer returns nsjail diagnostic logs in API responses.

User-facing:

* `stdout` contains program output only
* `stderr` contains program stderr only

Sandbox diagnostics are retained separately for operational debugging.

This prevents leakage of sandbox internals through execution results.

---

# Additional Security Controls

## nsjail Isolation

The service executes untrusted code through `nsjail`.

Sandboxing includes:

* mount isolation
* PID isolation
* network isolation
* user namespace isolation
* process limits
* memory limits
* execution time limits

## Namespace Usage

Isolation is delegated to `nsjail`.

The sandbox creates separate namespaces for:

* processes
* mounts
* networking
* users
* IPC
* UTS

This limits visibility into host resources.

## Capability Usage

The runtime image grants the capabilities required by `nsjail`.

Current deployment also uses privileged container execution during development.

This simplifies sandbox operation but increases host-level trust requirements.

## Temporary Workspace Handling

Each execution receives:

* a dedicated temporary directory
* isolated source files
* isolated build artifacts

The workspace is removed after execution.

## Output Truncation

Both stdout and stderr are capped.

Responses indicate when truncation occurs.

## Request Validation

Validation occurs before any code execution.

Rejected requests include:

* invalid JSON
* missing required fields
* unsupported languages
* oversized payloads
* invalid test configurations

## Cleanup Mechanisms

Cleanup occurs at two levels:

1. Per-request workspace cleanup.
2. Startup cleanup of stale workspaces.

---

# Security Assumptions and Known Limitations

The current implementation assumes:

* Linux host environment
* Functional nsjail installation
* Trusted compiler and interpreter binaries

The following are intentionally out of scope:

* Authentication
* Authorization
* Rate limiting
* Persistent storage
* Multi-tenant production deployment

Additional observations:

* Docker Compose currently runs with elevated privileges to support sandbox operation.
* Readiness checks verify tool availability but do not execute a full sandbox test.
* Language support is implemented in code rather than a dynamic registry.
* Load testing and sustained concurrency benchmarking are planned for later stages.

---

# Security Status

The implementation addresses all seven security issues identified in the hackathon specification through a combination of request validation, sandbox isolation, bounded resource usage, controlled process execution, workspace cleanup, and output limiting.

The remaining risks primarily relate to deployment configuration and operational hardening rather than direct vulnerabilities in the request execution path.
