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

---

## Verifying Python execution

**Prompt:**
How can I verify that Python execution is fully working and not just returning successful responses?

**Response summary:**
Suggested testing accepted cases, wrong output cases, runtime errors and stdin handling.

**What we used / didn't use:**
Used the suggested test cases and manually verified all Python execution paths through the API.

---

## Verifying C++ execution

**Prompt:**
How can I verify that C++ execution is fully working?

**Response summary:**
Suggested testing successful compilation, build failures, runtime errors and stdin handling.

**What we used / didn't use:**
Used the suggested tests and manually verified each execution path using API requests.

---

## Runtime error testing

**Prompt:**
How can I test runtime_error handling for C++ and Python?

**Response summary:**
Suggested programs that intentionally crash or raise exceptions.

**What we used / didn't use:**
Used the examples to verify that runtime errors were correctly mapped to runtime_error responses.

---

## Investigating sandbox logs in API responses

**Prompt:**
Why are nsjail logs appearing inside stderr responses and how should they be handled?

**Response summary:**
Explained that sandbox diagnostic logs should be separated from user program stderr.

**What we used / didn't use:**
Used the recommendation and updated the implementation so API responses only contain user stdout and stderr.

---

## Manual verification of execution results

**Prompt:**
Review the execution results and verify whether the implementation behaves according to the specification.

**Response summary:**
Reviewed API responses and suggested additional validation steps.

**What we used / didn't use:**
Used the review as guidance, but final verification was performed manually through repeated API testing because some execution paths could not be fully validated automatically.

---

## Testing build failures

**Prompt:**
How should build_failed responses be tested for compiled languages?

**Response summary:**
Suggested submitting intentionally invalid C++ code and verifying build status handling.

**What we used / didn't use:**
Used the approach to confirm build failures were reported correctly without crashing the service.

---

## Cleaning API responses

**Prompt:**
How should stdout and stderr be presented in API responses?

**Response summary:**
Suggested exposing only program output and hiding sandbox internals.

**What we used / didn't use:**
Used the recommendation and verified that successful runs returned clean responses without nsjail diagnostic output.

---

## Final Stage 1 verification

**Prompt:**
Check whether the implementation satisfies the Stage 1 requirements.

**Response summary:**
Reviewed the specification and compared it against the current implementation.

**What we used / didn't use:**
Used the checklist as a guide but manually validated endpoints, execution flows, tests and Docker functionality before considering Stage 1 complete.

---

## YAML Registry Migration

**Prompt:**
How can we make adding a language just one YAML block and one PR?

**Response summary:**
Suggested a dynamic YAML registry parsed at startup, combined with a generic `nsjail` runner that interprets the YAML configurations.

**What we used / didn't use:**
Implemented the YAML registry but intentionally kept the old Go files as a Stage 1 fallback (Strangler Fig pattern).

---

## Payload Testing Errors

**Prompt:**
R and OCaml are crashing with "Fatal error: the limit on the number of open files is too low" during blind payload testing.

**Response summary:**
Explained that R and OCaml aggressively open shared libraries and require higher limits. Suggested bumping generic default limits to "max".

**What we used / didn't use:**
Applied the generous sandbox defaults to the generic YAML runner to prevent OOM/file-descriptor crashes.

---

## Load Testing Analysis

**Prompt:**
Review the Vegeta load test CSV results. We are getting 100% errors past 5 RPS.

**Response summary:**
Analyzed the physics of the 2vCPU/2GB constraint vs the MemoryHog program. Proved the theoretical max throughput is ~2.4 RPS. Suggested implementing a 6-slot concurrency semaphore.

**What we used / didn't use:**
Implemented the semaphore to provide graceful degradation via request queuing and timeouts instead of container crashes. Documented the mathematical proof.
