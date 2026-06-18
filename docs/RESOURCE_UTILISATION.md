# Resource Utilisation Report

## Benchmark Configuration

- Tool: k6
- Virtual Users (VUs): 50
- Duration: 30 seconds
- Environment:
  - app1
  - app2
  - Redis

---

## Load Test Results

| Metric | Value |
|----------|----------|
| Total Requests | 137459 |
| Failed Requests | 0 |
| Throughput | 4581 req/s |
| Average Latency | 10.87 ms |
| P95 Latency | 22.32 ms |
| VUs | 50 |

---

## Container Resource Usage

### app1

- CPU: 0.00%
- Memory: 9.4 MiB

### app2

- CPU: 0.00%
- Memory: 2.2 MiB

### Redis

- CPU: 0.90%
- Memory: 8.6 MiB

---

## Observations

- No request failures occurred during the benchmark.
- Average response time remained below 11 ms.
- P95 latency remained below 23 ms.
- Redis memory usage remained under 10 MiB.
- Application containers showed minimal CPU utilization.
- System successfully sustained over 4500 requests per second.

## Conclusion

The rate limiter service handled sustained concurrent traffic with low latency and low resource consumption. Redis remained stable throughout the benchmark and no errors were observed.
