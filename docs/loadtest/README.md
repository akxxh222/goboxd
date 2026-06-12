# Load Test Results

## Environment
- **Container Limits:** 2 vCPU, 2 GB RAM
- **Load Testing Tool:** Vegeta (`github.com/tsenart/vegeta`)
- **Workload:** `MemoryHog.java` (allocates 150MB, holds for 1s, plus JVM compile/run overhead)

## Results
- **Breaking Point:** 10 RPS

## Failure Analysis
At 5 RPS, the service handled all 150 requests flawlessly with a p50 latency of ~2.01 seconds.

At 10 RPS, the system reached its breaking point, returning an 84% error rate. Because each `MemoryHog` instance allocates 150MB of RAM and holds it for 1 second, at 10 RPS the system attempts to allocate over 1.5GB of RAM concurrently, nearly maxing out the 2GB container limit. Additionally, spinning up 10 concurrent `javac` and `java` JVM processes heavily saturated the 2 vCPUs.

**Failure Mode:** Graceful degradation via Queuing and Timeout. The server did not hard-crash or get `OOMKilled`. Instead, intense resource contention caused request processing to slow down drastically. Incoming requests queued up until they hit the strict 10-second request timeout, at which point they cleanly failed.

## Reproduction
1. Start the server with `docker compose up -d` to enforce the 2 CPU / 2GB RAM limits.
2. Run the attack script: `docker compose --profile tools run --rm tools bash -c "cd docs/loadtest && ./load-test.sh"`
3. Generate the graphs: `docker compose --profile tools run --rm tools bash -c "cd docs/loadtest && python3 plot.py"`