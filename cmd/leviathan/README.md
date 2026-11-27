# Leviathan

Leviathan은 법원·공소청·중대범죄수사청의 기관 디렉터리를 소유하고 gRPC 조회 API로 제공합니다. 다른 서비스는 이 자료를 복사해 관리하지 않고 Leviathan의 계약을 통해 조회합니다.

## 제공 데이터

| 자료 | 수록 내용 |
|---|---|
| 법원 | 법원 215곳, 상위 법원·단계·종류·관할, 재판부, 소재지와 주소 |
| 공소청 | 공소청 본청과 소속 기관 66곳(광역공소청 6, 지방공소청 18, 지청 42) |
| 중대범죄수사청 | 본청과 지방청 6곳, 관할 지역과 주소 |

기관의 정적 자료와 검증 코드는 [`internal/leviathan`](../../internal/leviathan)에 있습니다. 원본 PDF는 저장소 용량을 줄이기 위해 Git에 포함하지 않으며, 최신 법령·별표와 주소는 공식 기관 홈페이지에서 확인합니다. 주소는 원문 표기를 기준으로 정규화하며, 공통 주소 적용 로직은 [`addresses.go`](../../internal/leviathan/addresses.go)에 있습니다.

시작 시 기관 관계와 주소 표가 실제 기관 목록과 맞는지 확인합니다. 없는 상위 기관을 참조하거나 주소 표가 기관 자료와 맞지 않으면 조용히 누락시키지 않고 초기화 오류로 알립니다. 데이터는 코드에 포함된 정적 자료이며, DB 쓰기나 기관 변경 RPC는 제공하지 않습니다.

## gRPC API

계약은 [`leviathan.proto`](../../api/proto/casepulse/leviathan/v1/leviathan.proto)에 있습니다.

| RPC | 기능 |
|---|---|
| `ListCourts` / `GetCourt` | 법원 목록과 ID 조회 |
| `ListProsecutionOffices` / `GetProsecutionOffice` | 공소청 및 소속 기관 목록과 ID 조회 |
| `ListInvestigationAgencies` / `GetInvestigationAgency` | 중대범죄수사청 기관 목록과 ID 조회 |

모든 목록·단건 응답은 기관의 주소와 제공 가능한 부가 정보를 포함합니다. 조회 대상 ID가 없으면 `NotFound`를 반환합니다. gRPC 서버 어댑터는 [`internal/leviathangrpc/server`](../../internal/leviathangrpc/server)에, 다른 서비스가 사용하는 클라이언트는 [`internal/leviathangrpc`](../../internal/leviathangrpc)에 분리되어 있습니다.

## 로그와 주소

요청이 들어오면 RPC 메서드와 요청 메시지를 터미널(표준 출력)에 기록합니다. 로그 파일은 만들지 않습니다. 표준 gRPC health 서비스(`grpc.health.v1.Health`)도 제공하며, 주기적인 health 요청은 로그에 남기지 않습니다.

| 설정 | 기본값 | 용도 |
|---|---|---|
| `-addr` | `:9090` | gRPC listen 주소 |
| `-healthcheck` | — | 서버를 켜지 않고 `-addr`의 gRPC health 상태를 확인한 뒤 종료(성공 0, 실패 1) |

## 실행과 테스트

저장소 루트에서 실행합니다.

```sh
go run ./cmd/leviathan
```

Compose 네트워크 주소는 `leviathan:9090`입니다. gRPC 포트는 호스트에 공개하지 않습니다.

```sh
docker compose up -d --build leviathan
docker compose logs -f leviathan
```

로컬에서 Aegis나 Clio를 별도로 실행할 때는 Leviathan 기본 주소 `localhost:9090`을 사용합니다.

```sh
go test ./internal/leviathan ./internal/leviathangrpc/... ./cmd/aegis
```
