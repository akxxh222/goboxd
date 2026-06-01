# Plan Evolution

## Initial understanding

**What we thought we'd do:**

At first I thought this was basically an online compiler project.

**What we actually did:**

After reading the specification more carefully, I realized the focus is sandboxing and secure execution, not just compiling code.

**Why it changed:**

The nsjail requirement and security section showed that process isolation is a major part of the project.

---

## First implementation

**What we thought we'd do:**

Build everything inside one Go file and get the API working first.

**What we actually did:**

Started with a simple version and later split the code into separate packages.

**Why it changed:**

The project became harder to maintain as more functionality was added.

---

## Testing strategy

**What we thought we'd do:**

Only use unit tests.

**What we actually did:**

Added integration tests for the /run endpoint.

**Why it changed:**

The specification asks for end-to-end testing and integration tests provide better coverage.

---

## Output handling

**What we thought we'd do:**

Return all process output directly.

**What we actually did:**

Limited captured output and separated sandbox logs from user output.

**Why it changed:**

Large outputs can consume memory and sandbox logs should not be exposed in API responses.