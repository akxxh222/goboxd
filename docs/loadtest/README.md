# Load Test Results

## Environment
- **Container Limits:** 2 vCPU, 2 GB RAM
- **Load Testing Tool:** Vegeta (`github.com/tsenart/vegeta`)
- **Workload:** `MemoryHog.java` (allocates 150MB, holds for 1s, plus JVM compile/run overhead)

## Results
- **Breaking Point:** < 5 RPS (Theoretical maximum throughput is ~2.4 RPS)

## Failure Analysis
The mathematical limit of this constraint was reached. `MemoryHog.java` sleeps for 1 second, and the Java compilation and JVM boot takes ~1-1.5 seconds. Each request therefore takes ~2.5 seconds minimum. 
To prevent the 2GB container from crashing via OOM, the system uses a strict concurrency semaphore limited to 6 simultaneous runs (6 runs * ~300MB = 1.8GB RAM). 
With a capacity of 6 runs every 2.5 seconds, the absolute maximum mathematical throughput is **~2.4 Requests Per Second**. 
When offered 5 RPS, the arrival rate exceeds the processing rate. The internal queue fills up instantly, and requests gracefully time out after 10 seconds.

**Failure Mode:** Graceful degradation via Queuing and Timeout. The server explicitly protected itself from an `OOMKilled` crash. Intense resource contention was safely managed by the concurrency limiter. Incoming requests queued up until they hit the strict 10-second context timeout, at which point they were cleanly rejected with HTTP 503 errors.

## Reproduction
1. Start the server with `docker compose up -d` to enforce the 2 CPU / 2GB RAM limits.
2. Run the attack script: `docker compose --profile tools run --rm tools bash -c "cd docs/loadtest && ./load-test.sh"`
3. Generate the graphs: `docker compose --profile tools run --rm tools bash -c "cd docs/loadtest && python3 plot.py"`