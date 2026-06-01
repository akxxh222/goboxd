# Architectural Decision Records

## Use nsjail for sandboxing

**Context:**

The service needs to execute untrusted code safely.

**Options considered:**

1. Run programs directly on the host.
2. Use Docker containers for every request.
3. Use nsjail.

**Decision:**

Use nsjail.

**Rationale:**

It provides process isolation and resource controls while keeping execution lightweight.

---

## Use fixed source filenames

**Context:**

User-controlled filenames can introduce path traversal risks.

**Options considered:**

1. Accept filenames from requests.
2. Generate filenames on the server.

**Decision:**

Use fixed filenames such as solution.py and solution.cpp.

**Rationale:**

This removes filename validation complexity and prevents traversal through filenames.

---

## Create a temporary workspace per request

**Context:**

Multiple requests may execute at the same time.

**Options considered:**

1. Shared workspace.
2. Separate workspace for every request.

**Decision:**

Use a separate temporary directory for every request.

**Rationale:**

Requests remain isolated and cleanup becomes easier.

---

## Separate build and execution stages

**Context:**

Compiled languages require a build step before execution.

**Options considered:**

1. Handle everything in one step.
2. Keep build and execution separate.

**Decision:**

Use separate build and execution stages.

**Rationale:**

This makes status reporting clearer and matches the API specification.