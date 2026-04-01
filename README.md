# Loggling

**Loggling** is an ultra-fast JSONL data refining engine built to process massive JSONL logs at high speed. It utilizes a Zero-copy Single-scan approach and Go's `sync.Pool` to minimize memory allocation while processing millions of logs in real-time.

---

## Benchmark

_Test Environment: Apple M1 Mac (8GB RAM)_

**1. Single-threaded Engine (V0.1.0)** - 10M Lines
| Metric | Standard Library (`json.Unmarshal`) | **Loggling (Single)** | Improvement |
| :------------------ | :---------------------------------- | :-------------------- | :-------------- |
| Throughput | ~40,000 logs/s | **730,000+ logs/s** | 18x faster |
| Memory Alloc | 1,024+ B/op | **1 copy/op** | Ultra-low alloc |
| Architecture | Reflection / Full Parse | Single-pass lexer | Efficiency-focused |

**2. Multi-core Parallel Processing (V0.2.0)** - 32.5M Lines (6 File K-Way Merge)
| Metric | **Loggling (Parallel)** | Notes |
| :------------------ | :---------------------- | :---------------------------------- |
| Throughput | **2,045,800+ logs/s** | 6 concurrent Go workers |
| Peak RAM | **12.9 MB** | Max RSS processing 32.5M lines |
| Total Time | **17.58 secs** | 100% sequential chronological sort |
| Architecture | Worker Pool + Channels | Sustained 4.2 cores utilization |

---

## Key Features

- **Zero-copy Single-scan Engine**  
  Identifies all fields in a single byte-level pass. No expensive reflection or full unmarshaling.

- **Data Diet (Stripper & Masker)**
    - Stripping: Physically removes unnecessary metadata fields, reducing storage and transfer costs by up to 80%.
    - Masking: Precisely masks sensitive fields (like passwords, emails) directly on the original bytes.

- **Atomic Metrics Engine**  
  Real-time aggregation of throughput (LPS), total processed count, drops, and errors using atomic counters with minimal performance degradation.

- **Operational Reliability**  
  Aims for 24/7 uninterrupted processing by recovering from panics even when abnormal logs like Broken JSON are encountered.

---

## Configuration Guide (YAML)

Loggling defines its pipeline using YAML-based configurations.

```yaml
# configs/config.yaml
default:
    input: ./data/input.log
    output: ./data/output.log

pipeline:
    # 1. Performance-focused filtering (executed first to reduce downstream overhead)
    filter:
        - { field: "level", value: "DEBUG" }

    # 2. Field removal for cost reduction
    stripper:
        - { field: "metadata" }
        - { field: "id" }

    # 3. Security masking
    masker:
        - { field: "password", preset: "password" }
```

---

## Getting Started

### Installation

```bash
git clone https://github.com/yourname/loggling.git
cd loggling
go build -o loggling ./cmd/loggling/main.go
```

### Usage

1. Define your pipeline in the `configs/config.yaml` file.
2. Run the engine with the following command:

```bash
./loggling
```

---

## Recent Updates

- **Network Gateway Mode (HTTP Ingress)**: Beyond parsing local files, Loggling can act as a persistent high-speed log ingestion server listening to `POST /logs`.
- **Zero-Downtime Log Rotation**: Perfectly handles `SIGHUP` signals to rotate log file handles seamlessly without restarting the process, natively supporting `logrotate` and `newsyslog`.
- **Multi-core Parallel Processing**: Utilizes Go Worker Pools to concurrently parse files within directories using all CPU cores.
- **TUI Progress Monitoring**: Check multi-line progress bars and per-worker TPS animations in real-time right in your terminal, similar to Docker.
- **Glob Path Integration**: Process multiple files and log directories simultaneously using wildcards (`*.log`).

---

## License

MIT License
