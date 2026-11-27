# Clio

Clio는 사건의 기본 정보를 소유하고 gRPC API로 제공합니다. 사건번호·사건명·종류·당사자·계속 법원·재판부를 관리하며, 화면이나 기관 원본 자료는 소유하지 않습니다. 사건은 PostgreSQL에 저장하며, DB 주소를 주지 않으면 메모리에 보관합니다. 사건 진행내역·송달·문서는 아직 구현 범위에 포함되지 않습니다.

## 서비스 흐름

1. 시작할 때 Leviathan의 `ListCourts` RPC를 호출해 법원·재판부 목록을 받습니다.
2. 받은 목록을 메모리에 보관하고 사건 저장소를 엽니다. `DATABASE_URL`이 있으면 PostgreSQL에 연결해 아직 적용하지 않은 스키마 마이그레이션을 적용하고, 없으면 메모리 저장소를 씁니다.
3. 저장소에 사건이 하나도 없으면 고정 시드로 만든 더미 사건 240건을 한 번에 넣습니다.
4. Aegis가 사건 조회·등록을 Clio gRPC API에 요청합니다.
5. 사건을 등록할 때 사건 필드, 법원 ID, 해당 법원에 속한 재판부인지, 중복 사건번호인지 확인합니다.

법원 목록은 Clio가 시작할 때 받은 스냅샷입니다. Leviathan의 법원 데이터를 바꾼 뒤에는 Clio를 재시작해야 새 목록이 반영됩니다. PostgreSQL을 쓰면 더미 자료와 등록 사건이 재시작 뒤에도 남고, 더미 자료는 다시 넣지 않습니다. 메모리 저장소를 쓰면 재시작할 때마다 초기화됩니다.

## 사건 검증

- 사건번호는 4자리 연도, 한글 사건부호, 1 이상의 일련번호로 구성됩니다.
- 사건 종류는 민사(`civil`), 행정(`administrative`), 가사(`family`), 민사신청(`civil_application`), 형사(`criminal`) 중 하나여야 합니다.
- 사건명과 당사자 목록이 필요하며, 각 당사자에는 지위와 이름이 있어야 합니다.
- 의뢰인으로 표시된 당사자가 한 명 이상이어야 합니다. 의뢰인이 여러 명이면 모두 같은 당사자 쪽이어야 합니다.
- 계속 법원과 재판부가 필요하고, 법원은 Leviathan 목록에 존재해야 하며 재판부는 그 법원 소속이어야 합니다.
- 같은 법원 안에서 같은 사건번호를 중복 등록할 수 없습니다.

## gRPC API

계약은 [`clio.proto`](../../api/proto/casepulse/clio/v1/clio.proto)에 있습니다.

| RPC | 기능 | 주요 오류 |
|---|---|---|
| `ListCases` | 사건 목록과 전체 건수 조회 | 저장소 장애 `Unavailable` |
| `GetCase` | ID로 사건 한 건 조회 | `NotFound`, 저장소 장애 `Unavailable` |
| `CreateCase` | 검증 후 사건 등록 | 입력 오류 `InvalidArgument`, 중복 `AlreadyExists`, 저장소 장애 `Unavailable` |

서버 어댑터는 [`internal/cliogrpc/server`](../../internal/cliogrpc/server)에 있고, Aegis가 사용하는 클라이언트는 [`internal/cliogrpc`](../../internal/cliogrpc)에 있습니다. Clio는 Leviathan의 클라이언트 어댑터([`internal/leviathangrpc`](../../internal/leviathangrpc))를 통해서만 기관 목록을 가져옵니다.

## 주소와 실행

| 설정 | 기본값 | 용도 |
|---|---|---|
| `-addr` | `:9091` | Clio gRPC listen 주소 |
| `LEVIATHAN_GRPC_ADDR` 또는 `-leviathan-addr` | `localhost:9090` | Leviathan gRPC 주소 |
| `DATABASE_URL` 또는 `-database-url` | (비어 있음) | PostgreSQL 접속 URL. 비어 있으면 메모리 저장소 |
| `CLIO_SEED_DUMMY` 또는 `-seed-dummy` | `true` | 저장소가 비어 있을 때 더미 사건을 넣을지 여부 |
| `-healthcheck` | — | 서버를 켜지 않고 `-addr`의 gRPC health 상태를 확인한 뒤 종료(성공 0, 실패 1) |

로컬에서는 Leviathan을 먼저 실행하고 Clio를 시작합니다.

```sh
go run ./cmd/leviathan
go run ./cmd/clio                     # 메모리 저장소

# 개발용 PostgreSQL 사용
docker compose up -d --wait postgres
DATABASE_URL='postgres://casepulse:casepulse@localhost:5432/casepulse?sslmode=disable' go run ./cmd/clio
```

DB가 아직 준비되지 않았으면 시작 후 30초까지 1초 간격으로 다시 접속합니다. 실행 중 DB를 쓸 수 없으면 gRPC `Unavailable`을 돌려주고, 자세한 오류는 Clio 로그에만 남깁니다.

Compose 네트워크에서는 `clio:9091`과 `leviathan:9090`을 사용합니다. gRPC 포트는 호스트에 공개하지 않습니다.

```sh
docker compose up -d --build clio
docker compose logs -f clio
```

Clio는 개인정보 보호를 위해 요청 로그에 RPC 메서드명만 남기며 사건 요청 본문은 출력하지 않습니다.

Clio는 표준 gRPC health 서비스(`grpc.health.v1.Health`)도 제공합니다. Compose는 이 서비스로 Clio 상태를 확인하고, Aegis는 시작할 때 이 서비스로 Clio 연결을 확인합니다. 주기적인 health 요청은 로그에 남기지 않습니다.

## 테스트

```sh
go test ./internal/clio/... ./internal/cliogrpc/... ./internal/leviathangrpc/... ./cmd/aegis

# PostgreSQL 저장소 통합 테스트(테스트마다 임시 DB를 만들고 지웁니다)
CLIO_TEST_DATABASE_URL='postgres://casepulse:casepulse@localhost:5432/casepulse?sslmode=disable' \
  go test ./internal/clio/postgres
```
