import csv
import matplotlib.pyplot as plt

rows = list(csv.DictReader(open("results.csv")))
rps  = [float(r["target_rps"]) for r in rows]

plt.figure(); plt.plot(rps, [float(r["error_pct"]) for r in rows], marker="o", color="red")
plt.xscale('log')
plt.xticks(rps, [str(int(x)) for x in rps])
plt.grid(True, which="both", ls="--", alpha=0.5)
plt.xlabel("Offered RPS"); plt.ylabel("Error rate (%)"); plt.title("Breaking point")
plt.savefig("breaking-point.png", dpi=150, bbox_inches="tight")

plt.figure()
for k, lbl in [("p50_ms","p50"), ("p95_ms","p95"), ("p99_ms","p99")]:
    plt.plot(rps, [float(r[k]) for r in rows], marker="o", label=lbl)
plt.xscale('log')
plt.xticks(rps, [str(int(x)) for x in rps])
plt.grid(True, which="both", ls="--", alpha=0.5)
plt.xlabel("Offered RPS"); plt.ylabel("Latency (ms)"); plt.title("RPS vs latency"); plt.legend()
plt.savefig("latency.png", dpi=150, bbox_inches="tight")