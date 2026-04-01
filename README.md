# Loggling

**Loggling** is an elite, high-performance JSONL log pre-processing engine built for extreme scale. By leveraging **Zero-copy Single-scan** technology and Go's `sync.Pool`, it manages to process millions of logs with near-zero memory allocation.

---

## Performance

Measured processing of 10 million lines of complex JSONL logs on a single thread.

| Metric | Standard Library (`json.Unmarshal`) | **Loggling (V1)** | Improvement |
| :--- | :--- | :--- | :--- |
| **Throughput** | ~40,000 logs/s | **730,000+ logs/s** | **18x Faster** |
| **Memory Alloc** | 1,024+ B/op | **~0 B/op** | **Zero-allocation** |
| **Architecture** | Reflection / Full Parse | **Single-pass Lexer** | **Efficiency Focus** |

---

## Key Features

- **Zero-copy Single-scan Engine**: Identifies all fields in a single byte-level pass. No expensive reflection or full unmarshaling.
- **Data Refining (Stripper & Masker)**: 
  - **Field Stripping**: Physically remove unnecessary fields to reduce storage costs by up to 80%.
  - **Surgical Masking**: In-place byte masking for sensitive data (Passwords, Emails).
- **Atomic Metrics Engine**: Real-time performance monitoring (LPS, Total, Dropped, Errored) with zero-cost atomic counters.
- **Production Ready Robustness**: Built-in panic recovery for malformed logs, ensuring 24/7 stability.

---

## Configuration (YAML)

Loggling uses a flexible YAML configuration to define your processing pipeline.

```yaml
# configs/config.yaml
default:
  input: ./data/input.log
  output: ./data/output.log

pipeline:
  # 1. Performance-first Filtering
  filter:
    - { field: "level", value: "DEBUG" }
  
  # 2. Cost-saving Field Stripping
  stripper:
    - { field: "metadata" }
    - { field: "id" }

  # 3. Security Masking
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
1. Configure your pipeline in `configs/config.yaml`.
2. Run the engine:
```bash
./loggling
```

---

## Roadmap

- [x] **V1.0**: Single-threaded Zero-copy Engine (Surgical Refining)
- [ ] **V1.5**: Multi-core Parallel Processing (Go Worker Pools)
- [ ] **V2.0**: Network Gateway Mode (gRPC / HTTP Ingress)
- [ ] **V2.5**: Hot Reload (Dynamic Config Update without Restart)

---

## License
MIT License
