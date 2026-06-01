# AI Prompts Log

## Understanding the project

**Prompt:**
What exactly is goboxd and what am I supposed to build for Stage 1?

**Response summary:**
Explained that goboxd is a Go service that runs untrusted code inside nsjail sandboxes and returns execution results through an HTTP API.

**What we used / didn't use:**
Used the explanation to understand the project scope and Stage 1 requirements.

---

## Building the HTTP endpoints

**Prompt:**
How should I implement /healthz, /readyz and /info in Go?

**Response summary:**
Suggested creating separate handlers and returning JSON responses.

**What we used / didn't use:**
Used the handler structure and adjusted the responses to match the specification.

---

## Running Python code

**Prompt:**
How can I execute Python code from Go and capture stdout and stderr?

**Response summary:**
Suggested using exec.CommandContext and temporary files.

**What we used / didn't use:**
Used the execution approach and adapted it for sandbox execution.

---

## Integrating nsjail

**Prompt:**
How do I run programs through nsjail from Go?

**Response summary:**
Suggested wrapping the execution command with an nsjail command and using a temporary workspace.

**What we used / didn't use:**
Used the nsjail integration approach and modified it to match project requirements.

---

## Security review

**Prompt:**
Review the repository against the hackathon security requirements.

**Response summary:**
Identified potential issues around output limits, cleanup, request validation and file handling.

**What we used / didn't use:**
Used the findings to verify and improve existing protections.

---

## Integration tests

**Prompt:**
How should end-to-end tests be structured for the /run endpoint?

**Response summary:**
Suggested testing the API through HTTP requests and validating the returned results.

**What we used / didn't use:**
Used the general approach and implemented integration tests for Python and C++.