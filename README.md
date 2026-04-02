# Loggling

Loggling is a high-performance JSONL data processing engine. It utilizes a Zero-copy single-scan approach and Go's `sync.Pool` to process large volumes of logs while minimizing memory allocations.

---

## Benchmark

_Test Environment: Apple M1 Mac (8GB RAM)_

**1. Single-threaded Processing** (10M Lines)
| Metric | Standard Library (`json.Unmarshal`) | **Loggling (Single)** | Improvement |
| :--- | :--- | :--- | :--- |
| Throughput | ~40,000 logs/s | **730,000+ logs/s** | ~18x faster |
| Memory Alloc | 1,024+ B/op | **Minimal (1 Copy)** | Allocation optimized |
| Architecture | Reflection / Full Unmarshaling | Single-pass lexer | Lower CPU usage |

**2. Multi-core Parallel Processing** (32.5M Lines, 6 File Merge)
| Metric | **Loggling (Parallel)** | Notes |
| :--- | :--- | :--- |
| Throughput | **2,045,800+ logs/s** | 6 concurrent Go workers |
| Peak RAM | **12.9 MB** | Maximum usage during 32.5M line merge |
| Total Time | **17.58 secs** | Maintains chronological order |
| Architecture | Worker Pool + Channels | Multi-core utilization |

---

## Key Features

- **Zero-copy Single-scan Engine**  
  Identifies JSON fields in a single byte-level pass, avoiding reflective parsing and full unmarshaling.

- **Data Filtering and Masking**
    - Stripping: Removes metadata fields to reduce storage footprint.
    - Masking: Obfuscates sensitive fields directly on the byte array.

- **Real-time Metrics Tracking**  
  Aggregates throughput, processed limits, drops, and errors using atomic variables to avoid lock contention.

- **Fault Tolerance**  
  Continues operations and avoids process panic even when parsing malformed JSON payloads.

---

## Configuration Guide

Loggling defines its runtime behavior via a YAML file.

```yaml
# configs/config.yaml
default:
    inputs:
        - "./data/*.log"
    output: ./data/output.log
    dlq: ./data/error_dlq.log

# Server mode configuration
server:
    enabled: true
    port: 8080

pipeline:
    # 1. Filter setup
    filter:
        - { field: "level", value: "DEBUG" }

    # 2. Targeted field removal
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
./loggling
```

---

## Recent Updates

- **DLQ (Dead Letter Queue)**: Detects structurally corrupted JSON payloads and isolates them to a designated file.
- **Hot Config Reloading**: Watches for modifications in `config.yaml` and replaces parsing rules at runtime safely, avoiding process restarts.
- **Network Gateway Mode (HTTP Ingress)**: Adds basic log ingestion server capabilities listening on `POST /logs`.
- **Log Rotation Support**: Rotates output file handles securely upon receiving a `SIGHUP` signal. Fully compatible with `logrotate`.
- **Parallel Processing Pool**: Allocates Go routines to parse files concurrently across available CPU cores.
- **Monitoring TUI**: Provides a console-based display measuring worker throughput and task progression.
- **Glob Path Integration**: Supports wildcard mappings (`*.log`) to parse entire directories automatically.

---

## License

MIT License
