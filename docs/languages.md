# Language Support

This document describes the languages supported by the current goboxd implementation.

## Supported languages

- `py3`: Python 3
- `cpp`: C++

## Language registry

- The repository does not use a YAML-driven language registry.
- Supported languages are hardcoded in `internal/runner/runner.go`.
- Adding a new language requires modifying the runner dispatch logic and implementing a new handler.

## Python 3

Identifier

- `py3`

Build step

- No explicit build step.
- The source code is written to `solution.py`.

Run step

- The executor runs `python3 solution.py` inside the sandbox.

Interpreter path

- `python3` on Linux.
- On Windows, the service falls back to `python` if it is used in a non-sandboxed mode.

Source filename strategy

- The source file is written to a fixed name: `solution.py`.

Artifact strategy

- No compiled artifact is produced.
- The source file is executed directly by the interpreter.

Sandbox execution flow

1. Create a temporary directory.
2. Write `solution.py` into that directory.
3. Invoke `nsjail` with sandbox options and `python3 solution.py`.
4. Capture stdout and stderr into bounded buffers.
5. Compare trimmed stdout to `tests[].expected_stdout`.
6. Remove the temporary directory and any `nsjail.log` file.

Supported request fields

- `language`
- `source`
- `tests[].stdin`
- `tests[].expected_stdout`

## C++

Identifier

- `cpp`

Build step

- The source code is written to `solution.cpp`.
- The build command is:
  - `g++ solution.cpp -o solution`
- The build runs in `nsjail` with larger sandbox limits.

Run step

- The compiled binary is executed as `./solution` inside the sandbox.

Compiler path

- `g++`

Source filename strategy

- The source file is written to a fixed name: `solution.cpp`.

Artifact strategy

- The binary is written to a fixed name: `solution`.
- The compiled executable is executed from the temporary working directory.

Sandbox execution flow

1. Create a temporary directory.
2. Write `solution.cpp` into that directory.
3. Build using `nsjail` and `g++`.
4. If the build fails, return `build_failed` and `not_executed` test results.
5. If the build succeeds, execute each test with `./solution`.
6. Capture stdout and stderr for each test.
7. Remove the temporary directory and any `nsjail.log` file.

Supported request fields

- `language`
- `source`
- `tests[].stdin`
- `tests[].expected_stdout`

## Extending language support

A new language implementation requires:

1. Adding a new case in `internal/runner/runner.go`.
2. Implementing source file creation and any build step.
3. Adding a sandboxed execution path in `internal/runner`.
4. Updating any API documentation and the `/info` response if desired.

The current implementation does not provide a generic plugin registry for languages.
