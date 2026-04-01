# Loggling

**Loggling** is a high-performance JSONL log pre-processing engine designed for large-scale workloads. By using **Zero-copy Single-scan** techniques and Go's `sync.Pool`, it can process millions of log lines with near-zero memory allocation.

---

## Performance

Benchmark: processing 10 million lines of complex JSONL logs on a single thread.

| Metric       | Standard Library (`json.Unmarshal`) | **Loggling (V1)**   | Improvement        |
| :----------- | :---------------------------------- | :------------------ | :----------------- |
| Throughput   | ~40,000 logs/s                      | **730,000+ logs/s** | 18x faster         |
| Memory Alloc | 1,024+ B/op                         | **1 copy/op**       | Ultra-low alloc    |
| Architecture | Reflection / Full Parse             | Single-pass lexer   | Efficiency-focused |

---

## Key Features

- Zero-copy Single-scan engine  
  Identifies required fields in a single byte-level pass, without reflection or full unmarshaling.

- Data refining (Stripper & Masker)
    - Field stripping: Physically removes unnecessary fields to reduce storage and transfer costs by up to 80%.
    - Surgical masking: In-place byte masking for sensitive fields such as passwords and emails.

- Atomic metrics engine  
  Tracks LPS, total processed, dropped, and errored logs using atomic counters for real-time visibility.

- Robustness in production  
  Handles malformed or broken JSON logs with panic recovery to keep the process running.

---

## Configuration (YAML)

Loggling uses a YAML configuration file to define the processing pipeline.

```yaml
# configs/config.yaml
default:
    input: ./data/input.log
    output: ./data/output.log

pipeline:
    # 1. Performance-first filtering
    filter:
        - { field: "level", value: "DEBUG" }

    # 2. Cost-saving field stripping
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

1. Configure your pipeline in `configs/config.yaml`.
2. Run the engine:

```bash
./loggling
```

---

## Recent Updates

- **Multi-core Parallel Processing**: Process massive glob directories efficiently utilizing all CPU cores with Go Worker Pools.
- **TUI Progress Monitoring**: View beautiful Docker-style, multi-line progress bars and TPS speeds directly in your terminal.
- **Glob Path Integration**: Easily tail and match multiple target files recursively.

---

## License

MIT License
