# Benchmarks

Extensive load testing has been performed to determine the mathematical breaking point and verify graceful degradation under strict constraints (2 vCPU, 2GB RAM).

## Load Test Summary

| Metric | Result |
| :--- | :--- |
| **Container Limits** | 2 vCPU, 2 GB RAM |
| **Target Workload** | `MemoryHog.java` (150MB RAM, 1s sleep per request) |
| **Theoretical Max Throughput** | ~2.4 RPS |
| **Observed Breaking Point** | < 5 RPS |
| **Failure Mode** | Graceful Degradation (HTTP 503 Timeouts via queuing) |
| **Concurrency Semaphore** | 6 simultaneous executions |

## Mathematical Constraints

The target workload requires ~300MB of RAM and takes ~2.5 seconds to execute (including compilation and JVM boot time). 
To prevent Docker from killing the container due to Out-Of-Memory (OOM) errors under the 2GB limit, `goboxd` employs a strict concurrency semaphore set to 6 simultaneous runs (`6 * 300MB = 1.8GB`).

With a capacity of 6 runs every 2.5 seconds, the absolute maximum mathematical throughput is **~2.4 RPS**. Any load exceeding this rate (such as 5 RPS) will immediately fill the queue and trigger graceful timeouts.

## Graceful Degradation

Instead of hard-crashing or consuming unbounded memory under heavy load, the service queues incoming requests. Once a request waits in the queue for more than 10 seconds, it is safely rejected with an `HTTP 503 Service Unavailable` error. This guarantees that the service remains stable and responsive to new connections even when completely saturated.

---
Please see `docs/loadtest/README.md` for full benchmark results, latency graphs, and failure analysis.