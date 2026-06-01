# goboxd

goboxd is a Go HTTP service that executes untrusted code inside `nsjail` sandboxes and returns structured build and test results over HTTP.

The current implementation supports Python 3 and C++ execution, request validation, resource limits, and sandboxed program execution through `nsjail`.

## Supported Languages

* Python 3 (`py3`)
* C++ (`cpp`)

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

## Project Structure

```text
cmd/            Service entrypoint
internal/       Application packages
docs/           Project documentation
tests/          Integration tests
```

## Documentation

Additional documentation is available under `docs/`:

* `docs/api.md` — API endpoints and request/response formats
* `docs/architecture.md` — System design and request flow
* `docs/security.md` — Security controls and mitigations
* `docs/languages.md` — Language execution model
* `docs/benchmarks.md` — Benchmarking and load-testing notes

## Current Status

Implemented:

* `GET /healthz`
* `GET /readyz`
* `GET /info`
* `POST /run`
* Python 3 execution
* C++ compilation and execution
* Request validation
* Output size limits and truncation
* Sandbox execution using `nsjail`
* Unit and integration tests
* Docker-based local development workflow

See the documentation in `docs/` for implementation details.
