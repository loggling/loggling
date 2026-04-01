# Loggling

Loggling은 JSONL 형식의 로그를 효율적으로 정제하기 위한 데이터 파이프라인 엔진입니다. Zero-copy 싱글 스캔 방식과 Go의 `sync.Pool`을 활용해 메모리 할당을 최소화하여 다량의 로그를 처리합니다.

---

## 성능 지표

_테스트 환경: Apple M1 Mac (8GB RAM)_

**1. 단일 스레드 파싱** (1,000만 줄 기준)
| 지표 | Go 표준 라이브러 (`json.Unmarshal`) | **Loggling (Single)** | 비고 |
| :--- | :--- | :--- | :--- |
| 처리량 | ~40,000 logs/s | **730,000+ logs/s** | 약 18배 향상 |
| 메모리 할당 | 1,024+ B/op | **최소화 (1 Copy)** | 할당 최적화 |
| 아키텍처 | 리플렉션, 전체 파싱 | 단일 패스 스캔 | 연산 효율 개선 |

**2. 다중 파일 병렬 파싱** (3,250만 줄 기준, 6 파일 병합)
| 지표 | **Loggling (Parallel)** | 비고 |
| :--- | :--- | :--- |
| 처리량 | **2,045,800+ logs/s** | 6개 워커 병렬 처리 |
| 최대 RAM 소모량 | **12.9 MB** | 3,250만 줄 병합 시 최대 사용량 |
| 처리 시간 | **17.58 초** | 시간순 정렬 보장 |
| 아키텍처 | Go Worker Pool + Channels | 다중 코어 활용 |

---

## 주요 기능

- **Zero-copy 단일 스캔 엔진**  
  필드 위치를 바이트 단위로 한 번만 탐색합니다. 큰 리플렉션이나 전체 언마셜링 작업을 수행하지 않습니다.

- **데이터 정제 및 마스킹**
  - **필드 삭제**: 불필요한 메타데이터 필드를 제거하여 저장 및 처리 용량을 줄입니다.
  - **데이터 마스킹**: 패스워드나 이메일 등 민감한 정보를 원본 바이트 배열 내에서 변경합니다.

- **실시간 메트릭 수집**  
  원자적(Atomic) 변수를 사용하여 처리량, 드롭 수, 에러 건수를 실시간으로 집계하며, 락(Lock) 지연을 방지합니다.

- **오류 복구**  
  손상된 JSON이 입력되어도 패닉을 발생시키지 않고 안정적으로 처리를 이어갑니다.

---

## 설정 가이드

Loggling은 YAML 파일을 통해 파이프라인 작동 방식을 정의합니다.

```yaml
# configs/config.yaml
default:
    input: ./data/input.log
    output: ./data/output.log

# 서버 환경설정
server:
    enabled: true
    port: 8080

pipeline:
    # 1. 필터링 구문
    filter:
        - { field: "level", value: "DEBUG" }

    # 2. 불필요한 필드 삭제
    stripper:
        - { field: "metadata" }
        - { field: "id" }

    # 3. 데이터 마스킹
    masker:
        - { field: "password", preset: "password" }
```

---

## 시작하기

### 빌드 및 실행

```bash
git clone https://github.com/yourname/loggling.git
cd loggling
go build -o loggling ./cmd/loggling/main.go
./loggling
```

---

## 버전별 업데이트 명세

- **DLQ (Dead Letter Queue)**: 문법 오류가 있는 손상된 JSON이 유입될 경우, 이를 탐지하여 별도의 dlq 지정 파일에 원본을 보존합니다.
- **설정 파일 핫 리로드**: 런타임 중 `config.yaml` 내용이 변경되면, 서비스를 재시작하지 않고 무잠금 방식으로 즉시 새 규칙을 반영합니다.
- **네트워크 게이트웨이 모드 (HTTP Ingress)**: `POST /logs` 엔드포인트를 열어 HTTP 요청 기반의 로그 수집 기능을 제공합니다.
- **로테이션 지원 (Zero-downtime Log Rotation)**: 운영 중 `SIGHUP` 시그널을 수신하면 프로세스 재시작 없이 새로운 로그 파일 핸들로 전환합니다. `logrotate` 도구와 호환 가능합니다.
- **병렬 처리 아키텍처**: Go 워커 풀을 구성하여 가용 프로세서를 활용해 다수의 로그 파일을 동시에 파싱합니다.
- **TUI 모니터링**: 터미널 UI를 통해 진행률과 워커별 처리 속도 정보를 제공합니다.
- **Glob 경로 지원**: 와일드카드(`*.log`)를 지정하여 하나의 디렉토리 내 복수의 파일을 처리합니다.

---

## 라이선스

MIT License
