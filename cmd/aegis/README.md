# Aegis

Aegis는 React 화면과 Gin 기반 HTTP API를 제공하는 진입 서비스입니다. 사건 데이터를 직접 저장하거나 검증하지 않고, 사건 작업은 Clio에, 법원 선택 목록은 Leviathan에 gRPC로 요청합니다.

## 책임과 요청 흐름

- 브라우저는 Aegis의 HTTP API를 호출합니다.
- 사건 목록·단건 조회·등록 요청은 Clio로 전달합니다.
- 등록 화면용 법원·재판부 선택지는 Leviathan의 응답으로 만듭니다.
- 사건 응답에 포함된 법원명·재판부·소재지·우편번호·주소는 Clio가 구성해 반환합니다.
- Aegis는 Clio와 Leviathan의 도메인 패키지를 가져오지 않습니다. 생성된 Protobuf 계약과 gRPC 클라이언트만 사용합니다.

시작할 때 Aegis는 Leviathan에서 법원 목록을 받고 Clio의 사건 목록을 조회해 두 gRPC 연결을 확인합니다. 연결에 실패하거나 법원 목록이 비어 있으면 시작하지 않습니다.

## HTTP API

| 메서드와 경로 | 동작 | 주요 응답 |
|---|---|---|
| `GET /api/cases` | 사건 목록과 전체 건수(`total`) 조회 | `200`, Clio 연결 실패 시 `503` |
| `GET /api/cases/{id}` | 사건 한 건 조회 | `200`, 없으면 `404`, Clio 연결 실패 시 `503` |
| `POST /api/cases` | 사건 등록 | 성공 시 `201`, 입력 오류 `400`, 중복 `409`, Clio 연결 실패 `503` |
| `GET /api/courts` | 재판부가 있는 법원 목록 조회 | `200` |

`POST /api/cases`는 다음 필드를 받습니다. 입력 검증과 사건 중복 확인은 Clio가 수행합니다.

```json
{
  "number": {"year": 2026, "code": "가합", "serial": 12345},
  "title": "손해배상",
  "category": "civil",
  "parties": [{"role": "원고", "name": "홍길동", "client": true}],
  "court": {"id": "court-district-seoul-central", "panel": "제91민사단독"}
}
```

HTTP JSON과 gRPC 메시지 사이의 변환 및 경로 등록은 [`api.go`](api.go)에 있습니다.

## 주소와 포트

| 설정 | 기본값 | 용도 |
|---|---|---|
| `-addr` | `:8080` | HTTP listen 주소 |
| `LEVIATHAN_GRPC_ADDR` 또는 `-leviathan-addr` | `localhost:9090` | Leviathan gRPC 주소 |
| `CLIO_GRPC_ADDR` 또는 `-clio-addr` | `localhost:9091` | Clio gRPC 주소 |

## 로컬 실행

저장소 루트에서 프런트엔드를 먼저 빌드합니다. Node.js와 npm이 필요합니다.

```sh
cd web
npm ci
npm run build
cd ..
```

이후 각 명령을 별도 터미널에서 실행합니다.

```sh
go run ./cmd/leviathan
go run ./cmd/clio
go run ./cmd/aegis
```

Aegis 화면은 <http://localhost:8080>에서 열 수 있습니다. 프런트엔드 개발 중에는 `web/`에서 `npm run dev`를 실행합니다. Vite가 화면 변경을 반영하고 `/api` 요청을 Aegis로 전달합니다. 빌드된 화면은 Go 바이너리에 포함되며, 프런트엔드를 빌드하지 않은 경우 API는 동작하고 `/`에는 빌드 안내가 표시됩니다.

## Docker Compose

```sh
docker compose up -d --build --wait
docker compose logs -f aegis
```

Compose는 세 서비스를 함께 실행하지만 호스트에는 Aegis의 `8080`만 공개합니다. Aegis는 내부 주소 `clio:9091`, `leviathan:9090`을 사용합니다.

## 테스트

```sh
go test ./cmd/aegis
```
