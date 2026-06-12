#!/bin/bash

# Exit on error
set -e

# Create Vegeta target file
echo "POST http://goboxd:8080/run" > target.txt
echo "Content-Type: application/json" >> target.txt
echo "@run-request.json" >> target.txt

# Initialize the CSV with headers
echo "target_rps,throughput_rps,duration_s,requests,success,failed,error_pct,p50_ms,p95_ms,p99_ms,max_ms" > results.csv

echo "Starting load test..."

for rate in 5 10 25 50 75 100 150 200 300 400; do
  echo "Testing at ${rate} RPS for 30s..."
  vegeta attack -rate=${rate}/1s -duration=30s -timeout=10s -targets=target.txt \
    | vegeta report -type=json > report-${rate}.json

  jq -r --arg r "$rate" '
    [ $r,
      .throughput,
      (.duration/1e9),
      .requests,
      (.status_codes["200"] // 0),
      (.requests - (.status_codes["200"] // 0)),
      ((1 - .success) * 100),
      (.latencies["50th"]/1e6),
      (.latencies["95th"]/1e6),
      (.latencies["99th"]/1e6),
      (.latencies.max/1e6)
    ] | @csv' report-${rate}.json >> results.csv
done

echo "Load test complete! Results saved to results.csv."