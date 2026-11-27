# Clio

Clio는 사건의 기본 정보를 소유하고 gRPC API로 제공합니다. 사건번호·사건명·종류·당사자·계속 법원·재판부를 관리하며, 화면이나 기관 원본 자료는 소유하지 않습니다. 사건 진행내역·송달·문서와 데이터베이스 영속성은 아직 구현 범위에 포함되지 않습니다.

## 서비스 흐름

1. 시작할 때 Leviathan의 `ListCourts` RPC를 호출해 법원·재판부 목록을 받습니다.
2. 받은 목록을 메모리에 보관하고, 고정 시드로 더미 사건 240건을 준비합니다.
3. Aegis가 사건 조회·등록을 Clio gRPC API에 요청합니다.
4. 사건을 등록할 때 사건 필드, 법원 ID, 해당 법원에 속한 재판부인지, 중복 사건번호인지 확인합니다.

법원 목록은 Clio가 시작할 때 받은 스냅샷입니다. Leviathan의 법원 데이터를 바꾼 뒤에는 Clio를 재시작해야 새 목록이 반영됩니다. 더미 자료와 수동 등록 자료는 모두 메모리에만 저장되므로 재시작하면 초기화됩니다.

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
| `ListCases` | 사건 목록과 전체 건수 조회 | — |
| `GetCase` | ID로 사건 한 건 조회 | `NotFound` |
| `CreateCase` | 검증 후 사건 등록 | 입력 오류 `InvalidArgument`, 중복 `AlreadyExists` |

서버 어댑터는 [`internal/cliogrpc/server`](../../internal/cliogrpc/server)에 있고, Aegis가 사용하는 클라이언트는 [`internal/cliogrpc`](../../internal/cliogrpc)에 있습니다. Clio는 Leviathan의 클라이언트 어댑터([`internal/leviathangrpc`](../../internal/leviathangrpc))를 통해서만 기관 목록을 가져옵니다.

## 주소와 실행

| 설정 | 기본값 | 용도 |
|---|---|---|
| `-addr` | `:9091` | Clio gRPC listen 주소 |
| `LEVIATHAN_GRPC_ADDR` 또는 `-leviathan-addr` | `localhost:9090` | Leviathan gRPC 주소 |

로컬에서는 Leviathan을 먼저 실행하고 Clio를 시작합니다.

```sh
go run ./cmd/leviathan
go run ./cmd/clio
```

Compose 네트워크에서는 `clio:9091`과 `leviathan:9090`을 사용합니다. gRPC 포트는 호스트에 공개하지 않습니다.

```sh
docker compose up -d --build clio
docker compose logs -f clio
```

Clio는 개인정보 보호를 위해 요청 로그에 RPC 메서드명만 남기며 사건 요청 본문은 출력하지 않습니다.

## 테스트

```sh
go test ./internal/clio ./internal/cliogrpc/... ./internal/leviathangrpc/... ./cmd/aegis
```
