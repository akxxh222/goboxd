# Load Testing Results

## Environment

The goboxd service was executed inside Docker with the following limits:

* 2 vCPU
* 2 GB RAM

The workload used was the provided `MemoryHog.java` program supplied in the challenge specification.

## Load Testing Tool

Load generation was performed using Vegeta.

Each load level was executed for 30 seconds with a request timeout of 10 seconds.

The following request rates were tested:

* 5 RPS
* 10 RPS
* 25 RPS
* 50 RPS
* 75 RPS
* 100 RPS
* 150 RPS
* 200 RPS
* 300 RPS
* 400 RPS

## Breaking Point

**Breaking Point: 5 RPS**

The first failed requests appeared at 5 RPS. Error rates increased rapidly as offered load increased and remained close to 100% for higher request rates.

### Mathematical Constraints (Why 5 RPS is the Breaking Point)

The theoretical maximum throughput of the system under the `2 vCPU / 2GB RAM` constraint is **~2.4 RPS**. 

1. **Duration:** `MemoryHog.java` sleeps for 1 second. Compiling (`javac`) and booting the JVM takes ~1-1.5 seconds. Minimum request duration = ~2.5 seconds.
2. **Memory:** The program allocates 150MB. JVM/compiler overhead adds ~100-150MB. Total memory per request = ~300MB.
3. **The Ceiling:** To prevent the 2GB container from crashing via `OOMKilled`, the application uses a strict concurrency semaphore limited to **6 simultaneous runs** (`6 * 300MB = 1.8GB`).

With a capacity of 6 runs every 2.5 seconds, the maximum throughput is ~2.4 RPS. When offered 5 RPS, the arrival rate immediately exceeds processing capacity. The internal semaphore queue fills up, and requests wait until they hit the 10-second context timeout.

## Observed Failure Mode

The primary failure mode was request timeout.

Most failed requests returned:

```text
context deadline exceeded
Client.Timeout exceeded while awaiting headers
```

This indicates that requests were waiting for execution capacity and exceeded the configured timeout before a response could be returned.

No container crashes were observed during testing.

No persistent service failure was observed after load was reduced.

## Latency Behaviour

Latency remained high throughout the test due to request queueing and timeout behaviour.

Observed latency characteristics:

* p50 latency approximately 5–6 seconds
* p95 latency approximately 12–13 seconds
* p99 latency approximately 14–15 seconds

The latency graph shows increasing delay as the service becomes saturated.

## Graceful Degradation

Under overload conditions the service degraded through request timeouts rather than crashing. Without the 6-slot concurrency semaphore, Vegeta sending 25+ RPS would have attempted to allocate 7+ GB of RAM, immediately blowing up the Docker container.

The service continued accepting connections and recovered after load decreased, indicating graceful degradation under excessive load.

## Reproducing the Test

1. Start the goboxd container with:

   * 2 vCPU
   * 2 GB RAM

2. Execute the load test:

```bash
./load-test.sh
```

3. Generate graphs:

```bash
python plot.py
```

4. Review:

* results.csv
* breaking-point.png
* latency.png

## Submitted Files

* results.csv
* breaking-point.png
* latency.png
* load-test.sh
* report-*.json
* plot.py
* run-request.json
* target.txt