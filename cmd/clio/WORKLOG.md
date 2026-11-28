# Clio 작업 내역과 코드 안내

Clio는 사건의 기본 정보(사건번호·사건명·종류·당사자·계속 법원·재판부)를 검증하고 저장하는 gRPC 서비스입니다. 실행·API 설명은 [README.md](README.md)를 참고하고, 이 문서는 **지금까지 한 작업**과 **처음 코드를 읽는 분을 위한 안내**를 정리합니다.

## 지금까지 한 작업

| 날짜 | 커밋 / PR | 내용 |
|---|---|---|
| 2026-11-27 | `5a7644c` / PR #3 | `Case` 자료형과 검사 함수를 추가했습니다. 사건번호는 연도+사건부호+일련번호(`ParseCaseNumber`, `String`), 당사자는 지위·이름을 직접 입력, 모든 사건에 법원과 재판부가 필수입니다. 재판부는 ID가 바뀔 수 있어 이름으로 저장합니다. |
| 2026-11-27 | `8c0070c` / PR #4 | 당사자마다 의뢰인 여부(`Party.Client`)를 표시하고, 의뢰인이 한 명 이상 있어야 하며 지위가 다른 당사자를 함께 의뢰인으로 하면 이해충돌로 거부합니다(`PartySide`, `Case.Clients`). |
| 2026-11-27 | `8749ea7` / PR #4 | 웹 서비스와 함께 더미 사건 240건 생성기(당시 `cmd/casepulse-web/dummy.go`)를 만들었습니다. |
| 2026-11-28 | `2cad1bb` | 서비스를 나누면서 사건 코드를 `internal/clio`로 옮기고 Clio gRPC 서비스(`cmd/clio`)를 만들었습니다. 계약은 `api/proto/casepulse/clio/v1/clio.proto`, 서버 어댑터는 `internal/cliogrpc/server`, 클라이언트는 `internal/cliogrpc`입니다. 시작할 때 Leviathan에서 법원 목록을 받아 법원·재판부 소속을 검사합니다. 요청 로그에는 개인정보 보호를 위해 메서드명만 남깁니다. |
| 2026-11-28 | `f85c9da` / PR #8 | 표준 gRPC health 서비스와 `-healthcheck` 옵션을 추가했습니다. |
| 2026-11-28 | `bc14b20` / PR #9 | 개발용 PostgreSQL 19 Compose 서비스(`postgres`)와 `.env.example`을 추가했습니다. |
| 2026-11-28 | `b5a45c1` / PR #10 | 저장소를 PostgreSQL로 바꿨습니다. `Store`(검증)와 `Repository`(저장: 메모리·PostgreSQL)를 나누고, pgx v5 저장소와 goose 마이그레이션(`cases`, `case_parties`)을 추가했습니다. 사건 ID는 `uuidv7()`, 같은 법원의 같은 사건번호는 고유 제약으로 막습니다. 저장소가 비어 있을 때만 더미를 넣고(`CLIO_SEED_DUMMY`), 입력 오류는 `InvalidArgument`, 저장소 장애는 `Unavailable`로 구분합니다. 두 저장소가 같은 동작을 하는지 공통 테스트(`repotest`)로 확인합니다. |
| 2026-11-28 | (이번 작업) | 초급 개발자도 읽을 수 있도록 소스 코드 주석을 보강하고 이 문서를 추가했습니다. |

## 구조 한눈에 보기

```text
Aegis ──gRPC──▶ cliogrpc/server ──▶ clio.Store ──▶ clio.Repository
                (요청↔자료형 변환)   (검증, 법원 정보 붙이기)   ├─ MemoryRepository   (DATABASE_URL 없음)
                                                          └─ postgres.Repository (DATABASE_URL 있음)
Leviathan ──gRPC(시작할 때 한 번)──▶ 법원 목록 → Store가 검증에 사용
```

계층을 나눈 이유는 각자 한 가지 일만 하게 하기 위해서입니다.

- **gRPC 어댑터**: 네트워크 메시지 ↔ Go 자료형 변환, 오류 → gRPC 상태 코드 변환
- **Store(서비스 계층)**: 입력 검사, 법원·재판부 소속 확인, 조회 결과에 법원 정보 붙이기
- **Repository(저장소)**: 저장과 조회만. 메모리든 DB든 같은 인터페이스를 지킵니다.

## 코드 읽는 순서

1. **[`internal/clio/cases.go`](../../internal/clio/cases.go)** — 사건 자료형과 검사 규칙. DB·네트워크와 관계없는 순수한 코드라 가장 먼저 읽기 좋습니다.
2. **[`internal/clio/store.go`](../../internal/clio/store.go)** — `Repository` 인터페이스와 `Store`의 `Create`·`List`·`Get` 흐름.
3. **[`internal/clio/memory.go`](../../internal/clio/memory.go)** — 가장 간단한 저장소 구현. 잠금(`sync.RWMutex`)과 "검사 후 저장" 패턴을 볼 수 있습니다.
4. **[`internal/clio/postgres/postgres.go`](../../internal/clio/postgres/postgres.go)** — 같은 인터페이스의 PostgreSQL 구현. 트랜잭션, `CopyFrom`, 고유 제약 오류 처리를 봅니다.
5. **[`internal/cliogrpc/server/server.go`](../../internal/cliogrpc/server/server.go)** — gRPC 요청을 Store로 연결하고 오류 코드를 정합니다.
6. **[`cmd/clio/main.go`](main.go)** — 위 부품을 조립해 서버를 시작합니다.

## 파일별 역할

| 파일 | 역할 |
|---|---|
| `cmd/clio/main.go` | 시작점: 법원 목록 받기 → 저장소 열기 → 더미 넣기 → gRPC 서버 |
| `internal/clio/cases.go` | `Case`, `CaseNumber`, `Party`, `CourtAssignment`와 검사 함수 |
| `internal/clio/court.go` | Leviathan에서 받은 법원 정보를 담는 Clio용 간단한 자료형 |
| `internal/clio/store.go` | 서비스 계층(`Store`), `Repository` 인터페이스, 오류 값 |
| `internal/clio/memory.go` | 메모리 저장소 |
| `internal/clio/dummy.go` | 고정 시드 더미 사건 240건 생성기 |
| `internal/clio/postgres/postgres.go` | PostgreSQL 저장소와 마이그레이션 실행 |
| `internal/clio/postgres/migrations/` | 스키마 SQL(goose). 기존 파일은 고치지 않고 다음 번호 파일을 추가합니다. |
| `internal/clio/repotest/repotest.go` | 메모리·PostgreSQL 공통 저장소 테스트 |
| `internal/cliogrpc/server/server.go` | gRPC 서버 어댑터 |
| `internal/cliogrpc/client.go` | Aegis가 쓰는 클라이언트 |

## 오류가 전달되는 길

| Clio 내부 오류 | gRPC 상태 코드 | Aegis HTTP 응답 |
|---|---|---|
| `ValidationError`(입력 오류) | `InvalidArgument` | 400 + 오류 메시지 목록 |
| `ErrDuplicateCase` | `AlreadyExists` | 409 |
| `ErrCaseNotFound` | `NotFound` | 404 |
| 그 밖의 저장소 오류(DB 중단 등) | `Unavailable` | 503 (자세한 내용은 Clio 로그에만) |

## 알아 두면 좋은 개념

- **인터페이스**: `Repository`는 메서드 목록만 정한 약속입니다. 메모리와 PostgreSQL 저장소가 같은 약속을 지키므로 `Store` 코드는 바꾸지 않고 저장 방식을 바꿀 수 있습니다.
- **오류 모으기**: `Validate`는 문제를 하나 찾고 멈추지 않고 `errors.Join`으로 모두 모아 돌려줍니다. 화면에 고칠 곳을 한 번에 보여 주기 위해서입니다.
- **트랜잭션**: 여러 사건을 한 번에 넣을 때 하나라도 실패하면 전부 취소합니다. 메모리 저장소도 "전부 검사한 뒤 저장"으로 같은 동작을 흉내 냅니다.
- **context**: 모든 저장소 메서드는 `ctx`를 받아, 요청이 취소되거나 시간이 지나면 DB 작업도 멈춥니다.
- **고정 시드 난수**: 더미 생성기는 시드를 고정해 실행할 때마다 같은 240건을 만듭니다.

## 남은 과제

- 사건 수정·삭제 RPC
- 목록 페이지 나누기(현재는 전체 목록을 한 번에 돌려줌)
- 사건 진행내역·송달·문서, 변호사·의뢰인 관리(아직 범위 밖)
- 법원 목록 갱신: 지금은 시작할 때 받은 스냅샷이라 Leviathan 데이터가 바뀌면 재시작해야 합니다.
