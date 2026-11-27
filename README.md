# CasePulse

법률사무소의 사건 기본 정보를 관리하고, 법원·공소청·중대범죄수사청 정보를 제공하는 Go 기반 서비스입니다.

## 서비스 구성과 책임

| 서비스 | 책임과 데이터 소유권 | 기본 주소 |
|---|---|---|
| **Aegis** | React 화면과 HTTP API. 사건 요청은 Clio에, 법원 선택 목록은 Leviathan에 gRPC로 전달합니다. | HTTP `:8080` |
| **Clio** | 사건 자료형·검증·저장소(PostgreSQL, DB 주소가 없으면 메모리)를 소유하고 사건 gRPC API를 제공합니다. 시작할 때 법원 목록을 Leviathan에서 받아 검증에 사용합니다. | gRPC `:9091` |
| **Leviathan** | 법원·공소청·중대범죄수사청의 정적 기관 자료와 주소를 소유하고 조회 gRPC API를 제공합니다. | gRPC `:9090` |

Aegis는 Clio나 Leviathan의 도메인 패키지를 직접 사용하지 않고 gRPC 계약으로만 통신합니다. 실행 파일마다 서버 어댑터와 클라이언트 어댑터도 분리되어 있습니다. Docker Compose에서는 Aegis의 HTTP 포트 `8080`만 호스트에 공개하고 gRPC 포트는 내부 네트워크에서만 사용합니다. 각 컨테이너는 실행 파일의 `-healthcheck` 옵션으로 상태를 보고하며, Clio는 Leviathan이, Aegis는 두 서비스가 healthy가 된 뒤에 시작합니다.

현재 구현은 사건 조회·등록과 기관 디렉터리 조회까지입니다. 변호사·법률사무소·법무법인·의뢰인 관리, 사건 진행내역·송달·문서 관리 기능은 아직 구현되지 않았습니다.

## 빠른 실행

Docker와 Docker Compose가 설치된 상태에서 저장소 루트에서 실행합니다.

```sh
docker compose up -d --build --wait
```

화면은 <http://localhost:8080>에서 열 수 있습니다. 상태와 로그는 `docker compose ps`, `docker compose logs -f aegis`로 확인하고, 종료할 때는 `docker compose down`을 실행합니다.

## 소스 구성

| 파일 | 내용 |
|---|---|
| `internal/leviathan/` | 법원·공소청·중대범죄수사청 자료형, 목록, 주소 및 재판부 데이터 |
| `internal/clio/` | 사건 자료형·검증, 저장소 계약과 메모리 저장소, 초기 더미 데이터 |
| `internal/clio/postgres/` | 사건 PostgreSQL 저장소와 스키마 마이그레이션(`migrations/`, goose) |
| `internal/clio/repotest/` | 메모리·PostgreSQL 저장소가 같은 방식으로 동작하는지 확인하는 공통 테스트 |
| `cmd/aegis/` | 사건목록 조회·등록 HTTP API와 웹 서버 |
| `cmd/clio/` | 사건 gRPC 서비스 |
| `cmd/leviathan/` | 기관 디렉터리를 제공하는 gRPC 서비스 |
| `api/proto/casepulse/{clio,leviathan}/`, `gen/go/` | 서비스별 gRPC Protobuf 계약과 생성된 Go 코드 |
| `internal/cliogrpc/`, `internal/cliogrpc/server/` | Clio gRPC 클라이언트와 서버 어댑터 |
| `internal/leviathangrpc/`, `internal/leviathangrpc/server/` | Leviathan gRPC 클라이언트와 서버 어댑터 |
| `internal/healthcheck/` | gRPC health 서비스 등록과 컨테이너 healthcheck용 점검 함수 |
| `internal/leviathan/addresses.go` | 서비스 내 기관 주소 표를 적용하는 공통 도우미 |
| `web/` | 사건목록 웹 화면(React + shadcn/ui, Vite) 및 Go 정적 파일 임베드 패키지 |
| `compose.yaml` | 웹·Clio·Leviathan 세 컨테이너와 개발용 PostgreSQL 구성, healthcheck |
| `.env.example` | 개발용 PostgreSQL 접속 정보 예시 |

서비스별 실행·구현 문서는 [Clio](cmd/clio/README.md), [Leviathan](cmd/leviathan/README.md), [Aegis](cmd/aegis/README.md)를 참고합니다.

Leviathan의 기관 목록은 패키지가 로드될 때 한 번 만들어지며, 표 사이의 값이 맞지 않으면(없는 상위 ID 참조, 주소 누락 등) 패닉을 일으켜 잘못된 데이터가 배포되지 않도록 합니다. `go test -v ./internal/leviathan -run PrintsTree`로 법원·공소청 트리를 출력해 데이터를 검토할 수 있습니다.

## 개발용 PostgreSQL

사건 데이터를 저장하는 개발용 PostgreSQL 19를 Compose의 `postgres` 서비스로 제공합니다. Compose의 Clio는 `DATABASE_URL`로 이 DB에 연결하며, 시작할 때 스키마 마이그레이션을 적용하고 사건이 하나도 없으면 더미 사건 240건을 넣습니다.

```sh
docker compose up -d --wait postgres   # DB만 실행
docker compose exec postgres psql -U casepulse -d casepulse
```

| 항목 | 기본값 |
|---|---|
| 호스트·포트 | `127.0.0.1:5432` (이 컴퓨터에서만 접속 가능) |
| 데이터베이스·사용자·비밀번호 | `casepulse` / `casepulse` / `casepulse` |
| 접속 URL | `postgres://casepulse:casepulse@localhost:5432/casepulse?sslmode=disable` |
| Compose 내부 주소 | `postgres:5432` |

기본값은 개발용입니다. 바꾸려면 `.env.example`을 `.env`로 복사해 수정합니다. `.env`는 Git에 올리지 않습니다. 데이터는 이름 있는 볼륨 `postgres-data`에 저장되므로 `docker compose down`으로 컨테이너를 지워도 남습니다. DB를 처음 상태로 되돌리려면 `docker compose down -v`로 볼륨까지 지웁니다. 사용자와 비밀번호는 볼륨을 처음 만들 때만 적용되므로, `.env`의 값을 바꾼 뒤에도 볼륨을 지워야 반영됩니다. 비밀번호는 Clio의 접속 URL에 그대로 들어가므로 `@`, `/`, `:` 같은 문자는 쓰지 않습니다.

스키마는 [`internal/clio/postgres/migrations`](internal/clio/postgres/migrations)의 SQL 파일로 관리하며 실행 파일에 포함됩니다. 스키마를 바꿀 때는 기존 파일을 고치지 않고 다음 번호의 파일(예: `00002_...sql`)을 추가합니다. 사건 ID는 PostgreSQL의 `uuidv7()`로 만듭니다.

## 법원 정보

법원 기관은 법원 단계, 관할 분야, 본원·지원·시·군법원 구분, 상위 법원, 소재지와 관할 지역을 별도 상세정보로 표현합니다. Leviathan의 [`internal/leviathan`](internal/leviathan)은 공식 법원 명칭·소재지 및 관할구역 별표를 바탕으로 215개 법원 목록을 제공합니다. 원본 PDF는 저장소에 포함하지 않으며, 최신 자료는 [대법원 홈페이지 각급법원 안내](https://www.scourt.go.kr/region/location/RegionLocationListAction.work)와 관련 법령에서 확인합니다. Clio와 Aegis는 법원 도메인 타입이나 자료를 직접 참조하지 않고, [`api/proto/casepulse/leviathan/v1/leviathan.proto`](api/proto/casepulse/leviathan/v1/leviathan.proto)의 gRPC 계약을 사용합니다. `Seat`은 별표의 소재지이며(다만 광주는 전남광주통합특별시 출범에 맞춰 `전남광주통합특별시`로, 옛 광주광역시 관할은 통합특별시의 5개 구로 정규화합니다)(시·군법원은 별표 2의 읍·시 단위 소재지, 예: 철원군법원은 `갈말읍`), 주소는 도로명 주소 형식으로 정규화하고 전화번호는 두지 않습니다. 우편번호(`PostalCode`)와 주소(`Address`)도 대법원 홈페이지 표기를 따릅니다. 별표에 신설되었으나 홈페이지에 아직 없는 전주가정법원과 그 지원은 주소를 비워 둡니다. 관할구역 별표의 개정 시점은 자료별로 다르므로 최신 법령과 대조해 갱신해야 합니다.

공식 재판부 구성표의 재판부는 소속 `Court`의 하위 `Panels`에 둡니다. 합의부·단독 재판부와 조정·재정·회생·파산 등 표에 기재된 재판 단위를 재판부명으로 기록합니다. 원외재판부는 별도의 고등법원 `Court`로 표현하고, 예를 들어 서울고등법원 인천 원외재판부는 `서울고등법원(인천)`이라는 이름의 고등법원으로 등록합니다. 원외 고등법원은 본원 고등법원과 동급이므로 둘 다 대법원을 상위 기관으로 가리키며, 인천 재판부들은 `서울고등법원(인천)`의 `Panels`에 직접 둡니다. 판사 개인 명단과 괄호·대괄호 안의 사건 분야 안내는 포함하지 않되, 안내가 붙은 재판부(예: `민사제2단독(조정)`)도 괄호만 떼고 반드시 넣습니다. 자료를 옮긴 뒤 2026년 11월 기준 공식 구성표와 대조해 누락된 재판부와 여러 쪽에 걸친 표의 뒷부분을 보완했다(약 570개 추가, 울산가정법원 재판부는 울산지방법원에서 이동, 원문 오타 2개 정정). 법원별 재판부 수는 테스트(`expectedPanelCounts`)에 고정해 두었으므로, 구성표 개정 시 최신 공식 자료와 대조해 데이터와 이 값을 함께 갱신해야 합니다.


## 공소청 정보

[`ProsecutionOffices`](internal/leviathan/prosecutions.go)는 「공소청의 위치와 각급 공소청 및 지청의 명칭 및 위치에 관한 규정」 별표(2026. 9. 29. 개정)를 바탕으로 공소청과 그 아래 66개 기관(광역공소청 6, 지방공소청 18, 지청 42)을 제공합니다. 원본 PDF는 저장소에 포함하지 않습니다. 각 기관은 `ParentOfficeID`로 상위 기관을 가리키며, 공소청 → 광역공소청 → 지방공소청 → 지청 순서로 이어집니다. 별표에는 지방공소청이 어느 광역공소청에 속하는지 적혀 있지 않으므로, [공소청 홈페이지의 조직구조](https://www.spo.go.kr/site/spo/06/10602010100002018100812.jsp)에 따라 연결했습니다. 이 관계는 지방법원과 고등법원의 소속 관계와 같습니다. 주소와 우편번호는 [공소청 홈페이지의 전국공소청찾기](https://www.spo.go.kr/site/spo/ex/announce/spoSearchList.do)를 따르며, 별표에 없는 공소청 자체의 소재지도 이 주소를 기준으로 기록합니다. 대표전화는 모든 기관이 콜센터 1301로만 안내되어 있어 `Phone`에 넣지 않습니다. 규정 개정 시 이 데이터도 함께 갱신해야 합니다.

## 중대범죄수사청 정보

[`InvestigationAgencies`](internal/leviathan/investigations.go)는 중대범죄수사청 본청과 6개 지방중대범죄수사청(서울·수원·대전·대구·부산·광주)을 제공합니다. 명칭·위치·관할구역은 「중대범죄수사청과 그 소속기관 직제」 별표 1을, 주소는 [중대범죄수사청 홈페이지](https://www.sci.go.kr/content.do?key=101041)를 따릅니다. 원본 PDF는 저장소에 포함하지 않으며 전화번호도 두지 않습니다. 지방수사청은 본청의 소속기관이므로 `ParentAgencyID`로 본청을 가리키며, 현판 등에 쓰는 약칭(예: `서울지방수사청`)은 `ShortName`에 둡니다. 수원지방수사청은 2027. 5. 이후 성남시 분당구로 이전할 예정이므로 그때 주소를 갱신해야 합니다.

## 사건 정보

[`Case`](internal/clio/cases.go)는 소송사건 한 건의 기본 정보를 담습니다.

- **사건번호**(`CaseNumber`): 연도(4자리) + 사건부호(한글, 예: `가합`·`카단`·`고단`) + 일련번호. `ParseCaseNumber("2026가합12345")`로 읽고 `String()`으로 같은 형식으로 씁니다. 사건번호는 법원마다 따로 매기므로 법원과 함께 있어야 사건을 특정할 수 있습니다.
- **사건명**(`Title`): 사용자가 직접 입력합니다.
- **사건 종류**(`Category`): 민사·행정·가사·민사신청·형사.
- **당사자**(`Parties`): `Party{Role, Name, Client}`의 목록으로, 지위와 이름 모두 사용자가 직접 입력합니다. 보통 원고·피고(형사는 피고인·검사) 2인이지만 1인이나 3인 이상도 될 수 있으며, 입력 순서를 유지합니다.
- **의뢰인**(`Party.Client`): 당사자마다 의뢰인 여부를 표시합니다. 같은 지위의 당사자가 여럿이어도 일부만 의뢰인일 수 있으므로(예: 원고 3명 중 원고1·원고3) 지위 단위가 아니라 사람 단위로 표시하며, 사건마다 의뢰인이 한 명 이상 있어야 합니다. 또한 지위가 다른 당사자를 함께 대리하면 이해충돌이 생기므로 의뢰인은 모두 같은 지위여야 합니다. 같은 지위인지는 `PartySide`로 판단하며 번호·괄호 표기·띄어쓰기는 무시한다(`원고1`·`원고3` → 원고, `피고(반소원고)` → 피고). `Case.Clients()`로 의뢰인만 꺼낼 수 있습니다.
- **계속 법원·재판부**(`CourtAssignment`): Leviathan 서비스가 발급한 법원 ID와 그 법원에 속한 재판부 이름. 형사사건을 포함한 모든 사건에 필수입니다. 재판부 ID는 구성표 개정 때 다시 매겨질 수 있으므로 사건에는 재판부 이름을 저장합니다. `Case.Validate()`는 두 값의 필수 여부를 검사하고, 존재·소속 검증은 Clio가 Leviathan의 법원 디렉터리로 수행합니다.

Clio는 시작할 때 Leviathan에서 법원 디렉터리를 gRPC로 받아 사건의 법원·재판부 소속을 확인합니다. `DATABASE_URL`이 있으면 사건을 PostgreSQL에 저장하므로 Clio를 재시작해도 남습니다. 비어 있으면 메모리 저장소를 쓰며, 이때는 재시작하면 등록 사건이 사라집니다.

## Aegis 웹 서비스

[`cmd/aegis`](cmd/aegis)는 사건목록을 조회하는 웹 서비스입니다. Gin 기반 Go 서버가 API와 화면을 함께 제공합니다.

```sh
# 화면 빌드 (Node.js 22 권장, 저장소 루트에서)
cd web && npm ci && npm run build && cd -

# 터미널 1: 기관 디렉터리
go run ./cmd/leviathan
# 터미널 2: 사건 서비스 (Leviathan에 연결)
go run ./cmd/clio
# 터미널 3: 웹 API (Clio와 Leviathan에 연결)
go run ./cmd/aegis                    # http://localhost:8080
go run ./cmd/aegis -addr :9000
```

웹 서비스는 시작할 때 Leviathan에서 법원 목록을 받고 Clio에 연결한 뒤, 사건 조회·등록 요청을 Clio gRPC로 전달합니다. 법원 선택 항목은 Leviathan 응답으로 구성합니다. Clio는 사건 목록·단건 조회·등록 RPC를 제공하고 등록 시 사건 필드와 법원·재판부 소속을 검사합니다. Leviathan은 기관 목록·ID 조회를 제공하며 주소도 응답에 포함합니다. 두 서비스는 요청 로그를 표준 출력에 남기며, Clio는 개인정보 보호를 위해 메서드명만 기록하고 사건 요청 내용을 남기지 않습니다. 기본 주소는 각각 `localhost:9090`(Leviathan), `localhost:9091`(Clio)이며 `LEVIATHAN_GRPC_ADDR`·`-leviathan-addr`, `CLIO_GRPC_ADDR`·`-clio-addr`로 바꿀 수 있습니다.

gRPC 계약을 수정하면 저장소 루트에서 Protobuf Go 코드를 다시 생성합니다.

```sh
protoc --go_out=. --go_opt=module=github.com/mkkim-gr8/casepulse \
  --go-grpc_out=. --go-grpc_opt=module=github.com/mkkim-gr8/casepulse \
  api/proto/casepulse/leviathan/v1/leviathan.proto \
  api/proto/casepulse/clio/v1/clio.proto
```

Dockerfile은 각 실행 프로그램 폴더에 있고, 저장소 루트에서 Compose로 세 서비스를 함께 실행합니다. 웹 프런트엔드 빌드 결과는 Go 바이너리에 임베드되므로 최종 이미지에 Node.js나 소스 파일은 포함되지 않습니다.

```sh
docker compose up -d --build --wait
```

사건 데이터와 검증은 Clio가, 법원·공소청·중대범죄수사청과 주소 데이터는 Leviathan이 소유합니다. 웹 API는 두 서비스의 gRPC 계약을 통해 데이터를 조회·등록합니다.

화면을 개발할 때는 Go 서버를 켜 둔 채 저장소 최상위 `web/`에서 `npm run dev`를 실행합니다. Vite 개발 서버(http://localhost:5173)가 고친 내용을 바로 반영하고 `/api` 요청은 Go 서버로 넘깁니다.

- **데이터**: 저장소가 비어 있으면 [`internal/clio/dummy.go`](internal/clio/dummy.go)가 Leviathan에서 받은 법원 목록으로 고정 시드 더미 사건 240건을 만들어 넣습니다(`CLIO_SEED_DUMMY=false`로 끌 수 있습니다). 당사자·사건번호는 지어낸 값이며, 법원·재판부는 사건 종류에 맞는 항목에서 고릅니다.
- **API**: `GET /api/cases`(전체 목록), `GET /api/cases/{id}`(한 건), `POST /api/cases`(사건 등록), `GET /api/courts`(등록 화면용 법원·재판부 목록), `GET /healthz`(상태 확인). 웹 API는 사건 조회·등록을 Clio gRPC로 전달하고, 법원 목록은 Leviathan에서 받아 응답합니다. 사건은 법원 이름·소재지·주소를 붙인 JSON으로 돌려줍니다. 검증 오류는 400과 함께 오류 메시지 목록을, 같은 법원에 같은 사건번호가 있으면 409를 돌려줍니다. Clio의 저장소를 쓸 수 없으면(DB 중단 등) 503을 돌려줍니다.
- **화면**: 저장소 최상위 [`web/`](web)의 React(JSX, TypeScript 없음) 앱이며 Vite로 빌드합니다. UI는 [shadcn/ui](https://ui.shadcn.com)(Radix 기반, Nova 프리셋)와 Tailwind CSS v4로 만들고, [Vite 설치 안내](https://ui.shadcn.com/docs/installation/vite)를 따르되 TypeScript 대신 JavaScript 방식(`components.json`의 `tsx: false`, `jsconfig.json`의 `@/` 별칭)을 씁니다. 컴포넌트는 `npx shadcn@latest add <이름>`으로 `src/components/ui/`에 추가하며, 화면 테마(밝은 화면·어두운 화면·시스템 설정)는 shadcn 다크 모드 안내의 ThemeProvider 방식을 따릅니다. 빌드 결과물 `web/dist`는 git에 올리지 않고 Go 실행 파일에 포함합니다. `web`의 Go 패키지가 이 디렉터리를 `embed`하고, `cmd/aegis`가 이를 서버 화면으로 제공합니다. 로고와 파비콘은 저장소 맨 위 [`casepulse-logo/`](casepulse-logo) 폴더가 원본이며, Vite가 이 폴더를 사이트 루트(`/favicon.ico`, `/casepulse-logo.svg` 등)로 그대로 내보냅니다. 메뉴줄 로고는 밝은 화면에서 `casepulse-logo.svg`, 어두운 화면에서 `casepulse-logo-dark.svg`를 씁니다. 빌드하지 않고 서버를 켜면 `/`에 빌드 안내가 나오고 API는 그대로 동작합니다.
- **기능**: 사건번호·사건명·당사자·법원·재판부 통합 검색(`/` 키로 이동), 사건 종류·접수 연도·계속 법원 필터, 의뢰인·다른 당사자 구분 표시, 열 정렬, 무한 스크롤(50건씩 이어서 표시), 사건 보기(한 번 클릭은 오른쪽 상세 패널로 미리보기, 더블클릭·⌘/Ctrl+클릭·패널의 `탭으로 열기` 버튼은 사건 탭으로 전체 화면 보기, 같은 사건은 탭 하나만 열림), 어두운 화면.
- **신건등록**: 메뉴줄의 `신건등록` 버튼, 사건목록 왼쪽 독바의 신건등록 아이콘, `N` 키로 등록 탭을 새로 연다(독바에는 필터 아이콘도 있습니다). 여러 건을 동시에 작성할 수 있고, 탭을 옮겨도 입력 내용이 남습니다. 사건번호 칸에 `2026가합12345`를 붙여 넣으면 연도·사건부호·일련번호로 나뉘어 들어가고, 당사자마다 의뢰인 여부를 체크하며, 법원·재판부는 이름으로 검색해 고른다(재판부 정보가 있는 법원만). `Ctrl+Enter`로 등록하면 탭이 닫히고 사건목록에서 새 사건의 상세 패널이 열립니다. 저장하지 않은 탭을 닫거나 브라우저를 닫으려 하면 확인을 받습니다.

## 개발 확인

저장소 루트에서 Go 테스트와 정적 분석을 실행합니다.

```sh
go test ./...
go vet ./...
docker compose config -q
```

PostgreSQL 저장소 통합 테스트는 `CLIO_TEST_DATABASE_URL`이 있을 때만 실행합니다. 테스트마다 임시 DB를 만들고 지우므로 DB를 만들 수 있는 사용자가 필요합니다.

```sh
docker compose up -d --wait postgres
CLIO_TEST_DATABASE_URL='postgres://casepulse:casepulse@localhost:5432/casepulse?sslmode=disable' \
  go test ./internal/clio/postgres
```

화면을 빌드하려면 `web/`에서 `npm ci`를 한 번 실행한 뒤 `npm run build`를 실행합니다.
