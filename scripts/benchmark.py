import argparse
import json
import random
import statistics
import time
from concurrent.futures import ThreadPoolExecutor, as_completed
from typing import List, Tuple
from urllib import request


def load_payloads(path: str) -> List[dict]:
    with open(path, "r", encoding="utf-8") as f:
        return json.load(f)


def post_json(url: str, payload: dict, timeout: float) -> Tuple[int, float]:
    body = json.dumps(payload).encode("utf-8")
    req = request.Request(url, data=body, method="POST")
    req.add_header("Content-Type", "application/json")
    start = time.perf_counter()
    try:
        with request.urlopen(req, timeout=timeout) as resp:
            status = resp.status
            resp.read()
    except Exception:
        status = 0
    elapsed_ms = (time.perf_counter() - start) * 1000.0
    return status, elapsed_ms


def percentile(data: List[float], p: float) -> float:
    if not data:
        return 0.0
    data_sorted = sorted(data)
    k = (len(data_sorted) - 1) * p
    f = int(k)
    c = min(f + 1, len(data_sorted) - 1)
    if f == c:
        return data_sorted[f]
    return data_sorted[f] + (data_sorted[c] - data_sorted[f]) * (k - f)


def main() -> None:
    parser = argparse.ArgumentParser(description="Simple API benchmark")
    parser.add_argument("--url", default="http://localhost:9999/fraud-score")
    parser.add_argument("--payloads", default="resources/example-payloads.json")
    parser.add_argument("--requests", type=int, default=500)
    parser.add_argument("--concurrency", type=int, default=20)
    parser.add_argument("--timeout", type=float, default=2.0)
    parser.add_argument("--seed", type=int, default=42)
    args = parser.parse_args()

    payloads = load_payloads(args.payloads)
    if not payloads:
        raise SystemExit("no payloads loaded")

    random.seed(args.seed)
    jobs = [random.choice(payloads) for _ in range(args.requests)]

    latencies: List[float] = []
    status_ok = 0
    status_err = 0

    start_all = time.perf_counter()
    with ThreadPoolExecutor(max_workers=args.concurrency) as executor:
        futures = [executor.submit(post_json, args.url, p, args.timeout) for p in jobs]
        for fut in as_completed(futures):
            status, elapsed_ms = fut.result()
            latencies.append(elapsed_ms)
            if status == 200:
                status_ok += 1
            else:
                status_err += 1
    total_ms = (time.perf_counter() - start_all) * 1000.0

    p50 = percentile(latencies, 0.50)
    p95 = percentile(latencies, 0.95)
    p99 = percentile(latencies, 0.99)

    print("requests:", args.requests)
    print("concurrency:", args.concurrency)
    print("ok:", status_ok)
    print("error:", status_err)
    print("total_ms:", round(total_ms, 2))
    print("avg_ms:", round(statistics.mean(latencies), 2))
    print("p50_ms:", round(p50, 2))
    print("p95_ms:", round(p95, 2))
    print("p99_ms:", round(p99, 2))


if __name__ == "__main__":
    main()
