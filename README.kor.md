# 로글링 (Loggling)

**Loggling**은 대규모 JSONL 로그를 고속으로 정제하기 위해 만든 초고속 JSONL 데이터 정제 엔진입니다. Zero-copy Single-scan 방식과 Go의 `sync.Pool`을 활용해 메모리 할당을 거의 없애면서도 수백만 건의 로그를 실시간으로 처리합니다.

---

## 성능 (Benchmark)

1,000만 줄짜리 복잡한 JSONL 로그를 단일 스레드로 처리했을 때의 결과입니다.

| 지표                | Go 표준 라이브러리 (`json.Unmarshal`) | **Loggling (V1)**     | 향상 수치       |
| :------------------ | :------------------------------------ | :-------------------- | :-------------- |
| 처리량 (Throughput) | ~40,000 logs/s                        | **730,000+ logs/s**   | 18배 이상 빠름  |
| 메모리 할당         | 1,024+ B/op                           | **~0 B/op**           | Zero-allocation |
| 아키텍처            | 리플렉션 / 전수 파싱                  | 비트를 훑는 단일 스캔 | CPU 효율 극대화 |

---

## 주요 기능

- Zero-copy Single-scan 엔진  
  필드 위치를 바이트 단위로 한 번만 훑어서 찾아냅니다. 리플렉션이나 전체 언마셜링 같은 무거운 작업은 사용하지 않습니다.

- 데이터 다이어트 (Stripper & Masker)
    - 필드 삭제 (Stripping): 불필요한 메타데이터 필드를 물리적으로 제거해서 저장·전송 비용을 최대 80%까지 줄입니다.
    - 정밀 마스킹 (Masking): 비밀번호, 이메일 등 민감한 필드만 골라서 원본 바이트 위에서 바로 마스킹합니다.

- Atomic 메트릭 엔진  
  원자적 카운터로 처리량(LPS), 총 처리 건수, 드롭/에러 건수를 실시간으로 집계하면서도 성능 저하를 최소화합니다.

- 운영 안정성  
  Broken JSON 같은 비정상 로그가 들어와도 패닉 없이 복구(Recover)해서 24시간 무중단 처리를 목표로 합니다.

---

## 설정 가이드 (YAML)

Loggling은 YAML 기반 설정으로 파이프라인을 정의합니다.

```yaml
# configs/config.yaml
default:
    input: ./data/input.log
    output: ./data/output.log

pipeline:
    # 1. 성능 중심 필터링 (가장 먼저 수행해서 뒤쪽 연산을 줄입니다)
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

1. `configs/config.yaml` 파일에 파이프라인을 정의합니다.
2. 아래 명령으로 엔진을 실행합니다.

```bash
./loggling
```

---

## 로드맵 (Roadmap)

- [x] V1.0: 단일 스레드 기반 Zero-copy 정제 엔진 (배포 준비 완료)
- [ ] V1.5: 멀티 코어 병렬 처리 (Worker Pool)
- [ ] V2.0: 네트워크 게이트웨이 모드 (gRPC / HTTP 입력)
- [ ] V2.5: 핫 리로드 (무중단 설정 반영)

---

## 라이선스 (License)

MIT License
