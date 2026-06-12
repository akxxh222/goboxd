# goboxd

goboxd is a Go HTTP service that executes untrusted code inside `nsjail` sandboxes and returns structured build and test results over HTTP.

The current implementation supports dynamic language configuration via YAML, request validation, resource limits, graceful degradation, and sandboxed program execution through `nsjail`.

## Supported Languages

| Language | Identifier |
| :--- | :--- |
| Bash | `bash` |
| C | `c` |
| C++ | `cpp` |
| Java | `java` |
| JavaScript (Node.js) | `node` |
| OCaml | `ocaml` |
| Python 3 | `py3` |
| R | `r` |
| Verilog | `verilog` |

## Running Locally

Build the Docker image:

```sh
make build
```

Start the service:

```sh
make run
```

Verify that the service is running:

```sh
curl http://localhost:8080/healthz
```

Expected response:

```json
{"status":"ok"}
```

Run unit tests:

```sh
make test
```

Run integration tests:

```sh
make integration
```

The service listens on port `8080`.

## Architecture Overview

goboxd implements the **Strangler Fig Pattern** to elegantly migrate from hardcoded executors to a dynamic plug-and-play system without dropping backward compatibility.

```mermaid
flowchart TD
    Req([Incoming POST /run]) --> Route[runner.Run()]
    Route --> CheckRegistry{Is language defined in<br>languages.yaml?}
    
    CheckRegistry -- "Yes (Stage 2/3)" --> GenericRunner[Generic YAML Runner]
    GenericRunner --> Sandbox[nsjail Sandbox Isolation]
    
    CheckRegistry -- "No (Stage 1 Fallback)" --> Switch{Legacy Switch Statement}
    Switch -- "case: py3" --> PyRunner[Python Runner]
    Switch -- "case: cpp" --> CppRunner[C++ Runner]
    Switch -- "case: java" --> JavaRunner[Java Runner]
    Switch -- "default" --> Fail[Reject Request]
    
    PyRunner --> Sandbox
    CppRunner --> Sandbox
    JavaRunner --> Sandbox
    
    Sandbox --> Res([HTTP Response])
    
    style GenericRunner fill:#2ea043,stroke:#fff,stroke-width:2px,color:#fff
    style Switch fill:#8b949e,stroke:#fff,stroke-width:2px,color:#fff
```

## Project Structure

| Directory | Description |
| :--- | :--- |
| `cmd/` | Service entrypoint |
| `internal/` | Application packages |
| `docs/` | Project documentation |
| `tests/` | Integration tests |

## Documentation

Additional documentation is available under `docs/`:

| Document | Description |
| :--- | :--- |
| `docs/api.md` | API endpoints and request/response formats |
| `docs/architecture.md` | System design and request flow |
| `docs/security.md` | Security controls and mitigations |
| `docs/languages.md` | Language execution model |
| `docs/benchmarks.md` | Benchmarking and load-testing notes |
| `docs/loadtest/README.md` | Load testing results and graceful degradation details |

## Current Status

Implemented features:

* `GET /healthz`
* `GET /readyz`
* `GET /info`
* `POST /run`
* Dynamic YAML plug-and-play language registry
* Graceful degradation under heavy load via concurrency semaphores
* Request validation
* Output size limits and truncation
* Sandbox execution using `nsjail`
* Unit and End-to-End integration tests
* Docker-based local development workflow

See the documentation in `docs/` for implementation details.
