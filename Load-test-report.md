# **Aegis Load Testing Report**
## **Setup**
- Redis + Aegis deployed via Docker (same host)
- Workload: `redis-benchmark`
- Total requests: 10k
- Concurrency: 50 / 100 VUs

---
## **1. Baseline (Redis Direct)**
- **Throughput:** ~42k–45k req/s
- **Avg Latency:** ~0.6–1.2 ms
- **p99 Latency:** ~2–4 ms

```text
Represents upper bound of achievable performance
```

---
## **2. Aegis Performance**

### **50 VUs (GET / SET)**
- **Throughput:** ~23k–26k req/s
- **Avg Latency:** ~2–3 ms
- **p99 Latency:** ~5–10 ms
---
### **100 VUs (GET / SET / DEL)**
- **Throughput:** ~27k–31k req/s
- **Avg Latency:** ~2.5–3.2 ms
- **p99 Latency:** ~7–12 ms

```text
Aegis sustains ~65–75% of native Redis throughput under load
```

---
## **3. Feature Impact Analysis**
- **All features disabled:** ~28k–30k req/s
- **Hotkey tracking enabled:** ~24k–26k req/s
- **Tag processing enabled:** marginal additional overhead

```text
Hotkey tracking introduces synchronization overhead (mutex contention) under concurrency
```
---
## **4. Optimization Findings**
- **Worker scaling / channel sizing:** negligible impact
- **Mutex (hotkey tracking):** measurable contention under load
- **RESP encoding optimizations:** inconsistent gains (CPU ↓ vs allocations ↑)
- **Buffered I/O (bufio):** minimal improvement

```text
Performance is not limited by CPU-bound computation or syscall overhead
```
---
## **5. Blank Proxy Benchmark (Critical Control Test)**
- **Throughput:** ~30k req/s
    
- **Avg Latency:** ~4–5 ms

```text
A zero-logic TCP proxy achieves performance close to Aegis
```

---
## **6. Root Cause Analysis**

```text
Primary bottleneck = network + proxy overhead, not application logic
```

Breakdown:
- Additional **TCP hop (proxy layer)**
- **Docker networking overhead**
- Kernel-level packet handling & context switching
- Go runtime scheduling variability (secondary factor)

```text
Data movement dominates cost, not request processing
```

---
## **7. Expected Performance in Production (Cloud / Native Linux)**
- **Estimated throughput:** ~32k–38k req/s
- (~70–85% of Redis baseline)

Reasons:
- Reduced container/network virtualization overhead
- More stable kernel networking stack
- Better CPU scheduling characteristics
---
## **Final Conclusion**

```text
Aegis operates close to the practical performance ceiling of a Redis proxy in a containerized environment.
```
- Business logic overhead is minimal
- Majority latency is introduced by networking and proxy layering

---
## **Final One-Liner**
```text
Aegis sustains ~65–75% of native Redis throughput in Docker, with overhead dominated by inherent proxy networking rather than implementation inefficiencies.
```


---

### Benchmark Results, from multiple runs:
Scenario:  `redis-benchmark -t get,set,del -n 10000 -p 6380 -c 100`

| Test Scenario                     | Description          | Throughput (RPS) | p50 Latency | p99 Latency |
| --------------------------------- | -------------------- | ---------------- | ----------- | ----------- |
| **Baseline (Redis)**              | Direct connection    | **42,700**       | 1.26 ms     | 3.1 ms      |
| **Aegis (Standard Run)**          | All features enabled | **26,900**       | 2.50 ms     | 9.30 ms     |
| **Aegis (Features Disabled)**     | All features off     | **27,200**       | 2.4 ms      | 9.10ms      |
| **Aegis (Hotkeys Enabled Only)**  | Hotkeys on, tags off | **26,750**       | 2.70 ms     | 13.8 ms     |
| **Aegis (Higher Contention Run)** | Higher latency run   | **22,700**       | 2.95 ms     | 12.06 ms    |
| **Aegis (Alternate Stable Run)**  | Another stable run   | **27,000**       | 2.54 ms     | 11.07 ms    |
| **RESP Writer Optimisation**      | Manual byte write    | **28,050**       | 2.45 ms     | 9.00 ms     |
| **Blank Proxy (Control)**         | Raw TCP forward      | **29,560**       | 2.37 ms     | 8.80ms      |


---



(Yes i know, the tests seem inconsistent, but the numbers are exactly what I observed, tested multiple times)