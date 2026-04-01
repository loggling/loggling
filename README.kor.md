# 로글링 (Loggling)

**Loggling**은 대규모 로그 처리를 극대화하기 위해 설계된 초고속 JSONL 데이터 정제 엔진입니다. **Zero-copy Single-scan** 기술과 Go의 `sync.Pool`을 활용하여 메모리 할당을 최소화하면서도 수백만 건의 로그를 실시간으로 가공합니다

---

## 성능 (Benchmark)

1,000만 줄의 복잡한 JSONL 로그를 단일 스레드에서 처리한 결과입니다.

| 지표 | Go 표준 라이브러리 (`json.Unmarshal`) | **Loggling (V1)** | 향상 수치 |
| :--- | :--- | :--- | :--- |
| **처리량 (Throughput)** | ~40,000 logs/s | **730,000+ logs/s** | **18배 이상 빠름** |
| **메모리 할당** | 1,024+ B/op | **~0 B/op** | **Zero-allocation** |
| **아키텍처** | 리플렉션 / 전수 파싱 | **비트를 훑는 단일 스캔** | **CPU 효율 극대화** |

---

## 주요 기능

- **Zero-copy Single-scan 엔진**: 필드 위치를 바이트 단위로 단 한 번만 훑어 찾아냅니다. 무거운 리플렉션이나 전체 언마셜링을 배제했습니다.
- **데이터 다이어트 (Stripper & Masker)**: 
  - **필드 삭제 (Stripping)**: 불필요한 메타데이터 필드를 물리적으로 제거하여 데이터 저장 및 전송 비용을 **최대 80%** 절감합니다.
  - **정밀 마스킹 (Masking)**: 비밀번호, 이메일 등 민감한 데이터만 원본 바이트 위에서 즉시 마스킹합니다.
- **Atomic 메트릭 엔진**: 성능 저하 없는 원자적 카운터를 통해 실시간 처리량(LPS), 총 처리량, 드롭/에러 건수를 실시간 모니터링합니다.
- **운영 안정성**: 비정상적인 로그(Broken JSON) 입력 시에도 패닉 없이 복구(Recover)하여 24시간 중단 없는 처리를 보장합니다.

---

## 설정 가이드 (YAML)

Loggling은 유연한 YAML 설정을 통해 데이터 파이프라인을 정의합니다.

```yaml
# configs/config.yaml
default:
  input: ./data/input.log
  output: ./data/output.log

pipeline:
  # 1. 성능 중심 필터링 (가장 먼저 수행하여 뒤쪽 연산을 절약)
  filter:
    - { field: "level", value: "DEBUG" }
  
  # 2. 비용 절감을 위한 필드 삭제
  stripper:
    - { field: "metadata" }
    - { field: "id" }

  # 3. 보안 마스킹
  masker:
    - { field: "password", preset: "password" }
```

---

## 시작하기

### 설치 (Installation)
```bash
git clone https://github.com/yourname/loggling.git
cd loggling
go build -o loggling ./cmd/loggling/main.go
```

### 사용법 (Usage)
1. `configs/config.yaml` 파일에 내 파이프라인 설정을 구성합니다.
2. 엔진을 실행합니다:
```bash
./loggling
```

---

## 로드맵 (Roadmap)

- [x] **V1.0**: 단일 스레드 기반 Zero-copy 정제 엔진 (배포 준비 완료)
- [ ] **V1.5**: 멀티 코어 병렬 처리 지원 (Worker Pool 도입)
- [ ] **V2.0**: 네트워크 게이트웨이 모드 (gRPC / HTTP 입력 지원)
- [ ] **V2.5**: 핫 리로드 (서버 중단 없는 실시간 설정 반영)

---

## 라이선스 (License)
MIT License
