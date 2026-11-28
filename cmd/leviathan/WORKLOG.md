# Leviathan 작업 내역과 코드 안내

Leviathan은 법원·공소청·중대범죄수사청의 기관 디렉터리를 gRPC로 제공하는 서비스입니다. 실행·API 설명은 [README.md](README.md)를 참고하고, 이 문서는 **지금까지 한 작업**과 **처음 코드를 읽는 분을 위한 안내**를 정리합니다.

## 지금까지 한 작업

| 날짜 | 커밋 / PR | 내용 |
|---|---|---|
| 2026-11-18 | `cff7607` / PR #1 | 전국 법원과 재판부를 원본 PDF에서 옮겨 적고, 원외재판부를 대법원 아래의 별도 고등법원으로 표현했습니다. Protobuf 스키마와 테스트를 처음 만들었습니다. |
| 2026-11-19 | `c7f225a` / PR #2 | 공소청(67곳)과 중대범죄수사청(본청+지방청 6곳) 목록을 추가했습니다. 시·군법원을 별표 2 기준 99곳으로 정리하고(없는 8곳 삭제 등), 부산가정법원 울산지원을 울산가정법원으로, 광주고등법원 전주·제주 원외재판부를 별도 고등법원으로 정정했습니다. 모든 기관에 우편번호와 정규화한 도로명 주소를 붙이고, 데이터 표를 파일별로 나눴습니다. |
| 2026-11-27 | `5745773` / PR #4 | 재판부구성표와 대조해 처음 옮길 때 빠진 재판부 약 570개를 보완했습니다(서울중앙지방법원 51→289 등). 울산가정법원 재판부를 울산지방법원에서 옮기고 원문 오타 2개를 바로잡았습니다. 법원별 재판부 수를 테스트(`expectedPanelCounts`)로 고정했습니다. |
| 2026-11-28 | `2cad1bb` | 하나였던 프로그램을 Aegis·Clio·Leviathan 세 서비스로 나눴습니다. 기관 데이터는 `internal/leviathan`으로 옮기고, gRPC 계약을 `api/proto/casepulse/leviathan/v1/leviathan.proto`로 따로 두었습니다. 서버 어댑터(`internal/leviathangrpc/server`)와 클라이언트(`internal/leviathangrpc`)를 분리하고 Dockerfile을 추가했습니다. |
| 2026-11-28 | `645750b` | 원본 PDF를 Git에서 빼고 공식 홈페이지 링크로 안내하도록 바꿨습니다. |
| 2026-11-28 | `f85c9da` / PR #8 | 표준 gRPC health 서비스, `-healthcheck` 옵션을 추가하고 health 요청은 로그에서 뺐습니다. Compose가 Leviathan이 healthy가 된 뒤 Clio를 시작합니다. |
| 2026-11-28 | (이번 작업) | 초급 개발자도 읽을 수 있도록 소스 코드 주석을 보강하고 이 문서를 추가했습니다. |

## 현재 제공하는 데이터

- 법원 215곳(대법원, 고등법원, 원외재판부, 특허법원, 지방·가정·행정·회생법원과 지원, 시·군법원)과 소속 재판부
- 공소청 1 + 광역공소청 6 + 지방공소청 18 + 지청 42
- 중대범죄수사청 본청 + 지방수사청 6

데이터는 소스 코드에 직접 적은 **정적 자료**입니다. 프로그램이 실행 중에 PDF나 웹을 읽지 않으며, 원본 자료가 바뀌면 코드를 고쳐야 합니다.

## 코드 읽는 순서

처음 읽을 때는 아래 순서를 추천합니다.

1. **[`cmd/leviathan/main.go`](main.go)** — 프로그램 시작점. 옵션을 읽고 gRPC 서버를 켭니다.
2. **[`api/proto/casepulse/leviathan/v1/leviathan.proto`](../../api/proto/casepulse/leviathan/v1/leviathan.proto)** — 다른 서비스와 약속한 RPC와 메시지 모양입니다. `gen/go/leviathan/v1/`은 이 파일에서 자동 생성된 코드이므로 직접 고치지 않습니다.
3. **[`internal/leviathangrpc/server/server.go`](../../internal/leviathangrpc/server/server.go)** — RPC 요청이 오면 내부 데이터를 찾아 Protobuf 메시지로 바꿔 돌려줍니다.
4. **[`internal/leviathan/casepulse.go`](../../internal/leviathan/casepulse.go)** — 기관 자료형(`Institution`, `Court`, `ProsecutionOffice`, `InvestigationAgency`) 정의.
5. **[`internal/leviathan/courts.go`](../../internal/leviathan/courts.go)** — 여러 표를 합쳐 법원 목록을 만드는 6단계 과정(`buildCourtDirectory`).
6. 나머지 데이터 파일은 필요할 때 봅니다.

## 파일별 역할

| 파일 | 역할 |
|---|---|
| `cmd/leviathan/main.go` | 시작점, 요청 로그(인터셉터) |
| `internal/leviathan/casepulse.go` | 패키지 설명과 자료형·상수 |
| `internal/leviathan/courts.go` | 법원 목록 조립 로직과 구조 표 |
| `internal/leviathan/courts_municipal.go` | 시·군법원 이름(시/군 구분) 표 |
| `internal/leviathan/courts_panels.go` | 법원별 재판부 이름 표(가장 큰 파일, 데이터만 있음) |
| `internal/leviathan/courts_addresses.go` | 법원 주소 표 |
| `internal/leviathan/prosecutions.go`, `prosecutions_addresses.go` | 공소청 목록과 주소 |
| `internal/leviathan/investigations.go` | 중대범죄수사청 목록(주소 포함) |
| `internal/leviathan/addresses.go` | 주소 표를 기관에 채우고 누락·오타를 검사하는 공통 함수 |
| `internal/leviathangrpc/server/server.go` | gRPC 서버 어댑터 |
| `internal/leviathangrpc/client.go` | Clio·Aegis가 쓰는 클라이언트 |
| `internal/healthcheck/healthcheck.go` | health 서비스 등록과 `-healthcheck` 점검 함수(세 서비스 공용) |
| `internal/leviathan/*_test.go` | 기관 수, 상위 ID, 주소, 재판부 수 등 데이터 검사 |

## 알아 두면 좋은 개념

- **패키지 초기화 때 한 번 만들기**: `var courtDirectory = buildCourtDirectory()`처럼 함수 밖의 변수는 `main`보다 먼저 초기화됩니다. 데이터가 서로 맞지 않으면 이때 `panic`이 나서 잘못된 데이터로 서버가 뜨지 않습니다.
- **복사본 돌려주기**: `Courts()`는 원본이 아니라 깊은 복사본을 돌려줍니다. 호출한 쪽이 슬라이스를 고쳐도 원본이 바뀌지 않습니다.
- **어댑터**: 내부 자료형(`leviathan.Court`)과 네트워크 자료형(`leviathanv1.Court`)을 나눠 두고 `courtToProto` 같은 함수로 옮겨 담습니다. 내부 구조를 바꿔도 다른 서비스와의 약속(proto)은 그대로 둘 수 있습니다.
- **인터셉터**: `grpc.UnaryInterceptor(logGRPCRequest)`는 모든 요청 앞에 끼어드는 함수로, 여기서는 요청 로그를 남깁니다.

## 데이터를 고칠 때

- 재판부를 고치면 `courts_test.go`의 `expectedPanelCounts`도 함께 고칩니다.
- 기관 이름을 바꾸면 주소 표의 이름도 같이 바꿔야 합니다(이름으로 주소를 찾습니다).
- `go test -v ./internal/leviathan -run PrintsTree`로 법원·공소청 트리를 출력해 눈으로 검토할 수 있습니다.
- Leviathan 데이터를 바꾼 뒤에는 Clio와 Aegis도 재시작해야 합니다(두 서비스는 시작할 때 받은 법원 목록을 씁니다).

## 남은 과제

- 경찰·등기소 기관 데이터(`InstitutionTypePolice`, `InstitutionTypeRegistry`는 종류만 정의됨)
- 전주가정법원과 그 지원의 주소(개원 후 확인)
- 수원지방수사청 이전(2027. 5. 이후) 주소 갱신
