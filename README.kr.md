# 로글링

**Loggling**은 대규모 JSONL 로그를 고속으로 정제하기 위해 만든 초고속 JSONL 데이터 정제 엔진입니다. Zero-copy Single-scan 방식과 Go의 `sync.Pool`을 활용해 메모리 할당을 최소화하면서 수백만 건의 로그를 실시간으로 처리합니다.

---

## 성능

_테스트 환경: Apple M1 Mac (8GB RAM)_

**1. 단일 스레드 파싱** - 1,000만 줄 기준
| 지표 | Go 표준 라이브러리 (`json.Unmarshal`) | **Loggling (Single)** | 향상 수치 |
| :------------------ | :------------------------------------ | :-------------------- | :-------------- |
| 처리량 | ~40,000 logs/s | **730,000+ logs/s** | 18배 이상 빠름 |
| 메모리 할당 | 1,024+ B/op | **최소화 (1 Copy)** | Ultra-low alloc |
| 아키텍처 | 리플렉션 / 전수 파싱 | 비트를 훑는 단일 스캔 | CPU 효율 극대화 |

**2. 다중 파일 병렬 파싱** - 3,250만 줄 기준 (6 File K-Way Merge)
| 지표 | **Loggling (Parallel)** | 참고 사항 |
| :------------------ | :---------------------- | :---------------------------------- |
| 처리량 | **2,045,800+ logs/s** | 6개의 워커가 동시에 병렬 파싱 |
| Peak RAM 소모량 | **12.9 MB** | 3,250만 줄을 합치는 동안의 최대 RAM |
| 처리 시간 | **17.58 초** | 시간순 재정렬 100% 보장 |
| 아키텍처 | Go Worker Pool + Channels | CPU 4.2코어 동시 풀가동 |

---

## 주요 기능

- Zero-copy Single-scan 엔진  
  필드 위치를 바이트 단위로 한 번만 훑어서 찾아냅니다. 리플렉션이나 전체 언마셜링 같은 무거운 작업은 사용하지 않습니다.

- 데이터 다이어트
    - 필드 삭제: 불필요한 메타데이터 필드를 물리적으로 제거해서 저장·전송 비용을 최대 80%까지 줄입니다.
    - 정밀 마스킹: 비밀번호, 이메일 등 민감한 필드만 골라서 원본 바이트 위에서 바로 마스킹합니다.

- Atomic 메트릭 엔진  
  원자적 카운터로 처리량, 총 처리 건수, 드롭/에러 건수를 실시간으로 집계하면서도 성능 저하를 최소화합니다.

- 운영 안정성  
  Broken JSON 같은 비정상 로그가 들어와도 패닉 없이 복구해서 24시간 무중단 처리를 목표로 합니다.

---

## 설정 가이드

Loggling은 YAML 기반 설정으로 파이프라인을 정의합니다.

```yaml
# configs/config.yaml
default:
    input: ./data/input.log
    output: ./data/output.log

# 🚨 [NEW] 서버 게이트웨이 모드 (HTTP Ingress)
server:
    enabled: true # true로 켜면 로컬 파일 탐색을 건너뛰고 포트를 엽니다.
    port: 8080 # 수신 대기할 포트 (기본값: 8080)

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

### 설치

```bash
git clone https://github.com/yourname/loggling.git
cd loggling
go build -o loggling ./cmd/loggling/main.go
```

### 사용법

1. `configs/config.yaml` 파일에 파이프라인을 정의합니다.
2. 아래 명령으로 엔진을 실행합니다.

```bash
./loggling
```

---

## 최근 업데이트

- **무중단 핫 리로드 (Hot Config Reload)**: 엔진이 수백만 건의 파싱을 진행하는 도중에도 `config.yaml`의 변경점을 실시간 감시하고 메모리를 락-프리(Lock-free)로 교체합니다. 서버 재시작 없이 파이프라인 룰이 0.001초 만에 갱신됩니다.
- **네트워크 게이트웨이 모드 (HTTP Ingress)**: 로컬 파일을 읽는 것을 넘어 `POST /logs` 포트를 수신하는 무한 대기형 고속 중앙 로그 수집 서버로 작동합니다.
- **무중단 다운타임 지원 (Zero-downtime Log Rotation)**: 운영 중인 서버에서 `SIGHUP` 신호망을 쏴주면, 프로세스 재시작 없이 빈 파일로 안전하게 연결을 갈아입습니다 (`logrotate` 및 `newsyslog` 완벽 호환).
- **다중 코어 병렬 처리**: Go 패키지의 워커 풀을 이용해 디렉토리 내 파일들을 모든 CPU 코어가 동시에 파싱합니다.
- **TUI 모니터링**: Docker 커맨드라인처럼 한눈에 들어오는 멀티라인 진행률 바와 워커별 처리 속도 애니메이션을 터미널에서 실시간으로 확인 가능합니다.
- **Glob 경로 탐색**: 와일드카드(`*.log`)를 활용하여 다중 파일과 로그 디렉토리를 한 번에 처리할 수 있습니다.

---

## 라이선스

MIT License
