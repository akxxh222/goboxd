# API Reference

This document describes the HTTP endpoints implemented by goboxd.

## GET /healthz

Purpose

- Basic liveness probe.

Request format

- No request body.
- Only HTTP GET is accepted.

Validation rules

- If the request method is not GET, the endpoint responds with `405 Method Not Allowed`.

Response format

- `200 OK`
- JSON object:
  - `status`: string, always `ok`

Example request

```http
GET /healthz HTTP/1.1
Host: localhost:8080
```

Example response

```json
{
  "status": "ok"
}
```

Error responses

- `405 Method Not Allowed` when the request method is not GET.
  - Response body: `{"error":"method not allowed"}`

## GET /readyz

Purpose

- Readiness probe that verifies required runtime binaries are available.

Request format

- No request body.
- Only HTTP GET is accepted.

Validation rules

- If the request method is not GET, the endpoint responds with `405 Method Not Allowed`.

Response format

- `200 OK` when all required binaries are present.
- `503 Service Unavailable` when one or more required binaries are missing.
- JSON object:
  - `status`: `ready` or `not_ready`
  - `checks`: object with keys `nsjail`, `python`, and `g++`
    - Each value is `ok` or `missing`.

Example request

```http
GET /readyz HTTP/1.1
Host: localhost:8080
```

Example response

```json
{
  "status": "ready",
  "checks": {
    "nsjail": "ok",
    "python": "ok",
    "g++": "ok"
  }
}
```

Example failure response

```json
{
  "status": "not_ready",
  "checks": {
    "nsjail": "ok",
    "python": "missing",
    "g++": "ok"
  }
}
```

Error responses

- `405 Method Not Allowed` when the request method is not GET.
  - Response body: `{"error":"method not allowed"}`

## GET /info

Purpose

- Returns service metadata and configurable limits.

Request format

- No request body.
- Only HTTP GET is accepted.

Validation rules

- If the request method is not GET, the endpoint responds with `405 Method Not Allowed`.

Response format

- `200 OK`
- JSON object containing:
  - `name`: service name, `goboxd`
  - `languages`: array of supported languages
  - `endpoints`: array of available endpoints
  - `limits`: configured runtime limits and sandbox settings

Example request

```http
GET /info HTTP/1.1
Host: localhost:8080
```

Example response

```json
{
  "name": "goboxd",
  "languages": [
    { "id": "py3", "name": "Python 3" },
    { "id": "cpp", "name": "C++" }
  ],
  "endpoints": [
    "GET /healthz",
    "GET /readyz",
    "GET /info",
    "POST /run"
  ],
  "limits": {
    "build_timeout_ms": 10000,
    "run_timeout_ms": 3000,
    "max_request_body_bytes": 1048576,
    "max_source_bytes": 262144,
    "max_tests": 25,
    "max_test_input_bytes": 65536,
    "max_expected_bytes": 65536,
    "max_captured_output": 65536,
    "sandbox": {
      "cpu_seconds": "3",
      "address_space_mb": "256",
      "file_size_mb": "1",
      "open_files": "32",
      "processes": "32"
    }
  }
}
```

Error responses

- `405 Method Not Allowed` when the request method is not GET.
  - Response body: `{"error":"method not allowed"}`

## POST /run

Purpose

- Compile and execute submitted code inside an isolated sandbox.
- Run one or more test cases against the submitted program.

Request format

- Content-Type: `application/json`
- JSON object with the following fields:
  - `language` (string, required)
  - `source` (string, required)
  - `tests` (array of test objects, required)

Request field descriptions

- `language`: runtime identifier. Supported values are `py3` and `cpp`.
- `source`: source code string to execute.
- `tests`: array of test cases. Each test case contains:
  - `stdin`: input string to provide on standard input.
  - `expected_stdout`: expected output string on standard output.

Validation rules

- The request body is limited to 1 MiB.
- `language` is required.
- `source` is required and may be at most 256 KiB.
- `tests` is required and must contain at least one element.
- No more than 25 tests are allowed.
- Each `stdin` value may be at most 64 KiB.
- Each `expected_stdout` value may be at most 64 KiB.

Response format

- `200 OK` on a valid request.
- JSON object with:
  - `status`: run-level status
  - `build`: optional build metadata for C++ submissions
  - `tests`: array of test result objects

Build result fields

- `status`: `ok`, `failed`, or `time_limit_exceeded`
- `stdout`: captured standard output from the build command
- `stderr`: captured standard error from the build command
- `stdout_truncated`: true if output exceeded 64 KiB
- `stderr_truncated`: true if output exceeded 64 KiB
- `duration_ms`: build duration in milliseconds

Test result fields

- `status`: `accepted`, `wrong_output`, `runtime_error`, `time_limit_exceeded`, or `not_executed`
- `stdout`: captured standard output from the program run
- `stderr`: captured standard error from the program run
- `stdout_truncated`: true if output exceeded 64 KiB
- `stderr_truncated`: true if output exceeded 64 KiB
- `duration_ms`: test duration in milliseconds

Example request

```http
POST /run HTTP/1.1
Host: localhost:8080
Content-Type: application/json

{
  "language": "py3",
  "source": "print(input())",
  "tests": [
    {
      "stdin": "hello\n",
      "expected_stdout": "hello"
    }
  ]
}
```

Example response for Python

```json
{
  "status": "accepted",
  "tests": [
    {
      "status": "accepted",
      "stdout": "hello\n",
      "stderr": "",
      "stdout_truncated": false,
      "stderr_truncated": false,
      "duration_ms": 50
    }
  ]
}
```

Example response for C++

```json
{
  "status": "accepted",
  "build": {
    "status": "ok",
    "stdout": "",
    "stderr": "",
    "stdout_truncated": false,
    "stderr_truncated": false,
    "duration_ms": 150
  },
  "tests": [
    {
      "status": "accepted",
      "stdout": "hello\n",
      "stderr": "",
      "stdout_truncated": false,
      "stderr_truncated": false,
      "duration_ms": 35
    }
  ]
}
```

Error responses

- `400 Bad Request` for validation failures.
  - Example: `{"error":"language is required"}`
  - Example: `{"error":"source is too large"}`
  - Example: `{"error":"too many tests"}`
  - Example: `{"error":"unknown language"}`
- `400 Bad Request` for invalid JSON or request bodies larger than 1 MiB.
  - Example: `{"error":"invalid JSON request"}`
  - Example: `{"error":"request body too large"}`
- `500 Internal Server Error` if the server fails to create or prepare a temporary work directory.
  - Example: `{"error":"failed to create temp directory"}`

Status vocabulary mapping

- Top-level `status` values:
  - `accepted`: all tests passed.
  - `wrong_output`: program exited successfully but output did not match expected stdout.
  - `build_failed`: C++ compilation failed.
  - `runtime_error`: runtime execution failed for a test.
  - `time_limit_exceeded`: execution or build exceeded the configured timeout.
  - `internal_error`: an internal error occurred while preparing the run.
- Build `status` values:
  - `ok`
  - `failed`
  - `time_limit_exceeded`
- Test `status` values:
  - `accepted`
  - `wrong_output`
  - `runtime_error`
  - `time_limit_exceeded`
  - `not_executed`

Notes

- Supported languages are hardcoded in `internal/runner/runner.go`.
- There is no YAML-driven language registry in the current implementation.
- The `/readyz` endpoint verifies only binary availability, not sandbox capability.
