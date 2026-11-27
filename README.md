# casepulse
golang based micro service architecture for legal office and its letigation case management 

casepulse 는 다음 역할을 수행하는 서비스이다.
1. 법원, 공소청(검찰), 경찰, 등기소 정보를 관리한다.
2. 변호사, 법률사무소, 법무법인 정보를 관리한다.
3. 의뢰인 정보를 관리한다.
4. 소송사건의 정보만을 관리한다. 진행내역, 대응내역, 송달관리, 문서관리는 별도의 서비스를 구축한다.

본 프로젝트 casepulse는 Micro Service로 이용한다. 독립한 하나의 서비스임.

## 소스 구성

| 파일 | 내용 |
|---|---|
| `casepulse.go` | 도메인 자료형(기관 종류·분류 상수, `Institution`, `Court`, `CourtPanel`, `ProsecutionOffice`, `InvestigationAgency`) |
| `addresses.go` | 주소 표 한 줄의 자료형과, 기관 이름으로 주소를 채우는 공통 함수 |
| `courts.go` | 법원 목록을 만드는 로직과 법원 구조 표(고등법원·지방법원·지원·가정법원·회생법원 등) |
| `courts_municipal.go` | 시·군법원의 시/군 구분과 별표 2 소재지 표 |
| `courts_panels.go` | 재판부구성표의 법원별 재판부 이름 표 |
| `courts_addresses.go` | 법원 우편번호·주소 표 |
| `prosecutions.go`, `prosecutions_addresses.go` | 공소청 목록과 주소 표 |
| `investigations.go` | 중대범죄수사청 목록(주소 포함) |
| `cmd/casepulse-web/` | 사건목록 조회·등록 웹 서비스(Go 서버 + shadcn/ui 화면, 더미 데이터) |
| `cases.go` | 소송사건 기본 정보 자료형(`Case`, `CaseNumber`, `Party`, `CourtAssignment`)과 검사 함수 |

모든 목록은 패키지가 로드될 때 한 번 만들어지며, 표 사이의 값이 맞지 않으면(없는 법원 ID 참조, 주소 누락 등) 패닉을 일으켜 잘못된 데이터가 배포되지 않도록 한다. `go test -v -run PrintsTree`로 법원·공소청 트리를 출력해 데이터를 검토할 수 있다.

## 법원 정보

법원 기관은 법원 단계, 관할 분야, 본원·지원·시·군법원 구분, 상위 법원, 소재지와 관할 지역을 별도 상세정보로 표현한다. [`Courts`](courts.go)는 [`documents/pdfs/court`](documents/pdfs/court)에 둔 법원 명칭·소재지 및 관할구역 별표를 바탕으로 215개 법원 목록을 제공한다. `Seat`은 별표의 소재지이며(다만 광주는 전남광주통합특별시 출범에 맞춰 `전남광주통합특별시`로, 옛 광주광역시 관할은 통합특별시의 5개 구로 정규화한다)(시·군법원은 별표 2의 읍·시 단위 소재지, 예: 철원군법원은 `갈말읍`), 주소는 도로명 주소 형식으로 정규화하고 전화번호는 두지 않는다. 또한 우편번호(`PostalCode`)와 주소(`Address`)는 [대법원 홈페이지 각급법원 안내](https://www.scourt.go.kr/region/location/RegionLocationListAction.work)를 따른다. 별표에 신설되었으나 홈페이지에 아직 없는 전주가정법원과 그 지원은 주소를 비워 둔다. 관할구역 별표의 개정 시점은 자료별로 다르므로 최신 법령과 대조해 갱신해야 한다.

[`documents/pdfs/court/재판부구성표.pdf`](documents/pdfs/court/재판부구성표.pdf)에 나온 재판부는 소속 `Court`의 하위 `Panels`에 둔다. 합의부·단독 재판부와 조정·재정·회생·파산 등 표에 기재된 재판 단위를 재판부명으로 기록한다. 원외재판부는 별도의 고등법원 `Court`로 표현하고, 예를 들어 서울고등법원 인천 원외재판부는 `서울고등법원(인천)`이라는 이름의 고등법원으로 등록한다. 원외 고등법원은 본원 고등법원과 동급이므로 둘 다 대법원을 상위 기관으로 가리키며, 인천 재판부들은 `서울고등법원(인천)`의 `Panels`에 직접 둔다. 판사 개인 명단과 괄호·대괄호 안의 사건 분야 안내는 포함하지 않되, 안내가 붙은 재판부(예: `민사제2단독(조정)`)도 괄호만 떼고 반드시 넣는다. 처음 옮겨 적을 때 이런 재판부와 서울중앙지방법원처럼 여러 쪽에 걸친 표의 뒷부분이 빠졌던 것을 2026. 11. PDF와 대조해 보완했다(약 570개 추가, 울산가정법원 재판부는 울산지방법원에서 옮김, 원문 오타 2개 정정). 법원별 재판부 수는 테스트(`expectedPanelCounts`)에 고정해 두었으므로, 구성표 개정 시 PDF와 대조해 데이터와 이 값을 함께 갱신해야 한다.


## 공소청 정보

[`ProsecutionOffices`](prosecutions.go)는 [`documents/pdfs/prosecution`](documents/pdfs/prosecution)에 둔 「공소청의 위치와 각급 공소청 및 지청의 명칭 및 위치에 관한 규정」 별표(2026. 9. 29. 개정)를 바탕으로 공소청과 그 아래 66개 기관(광역공소청 6, 지방공소청 18, 지청 42)을 제공한다. 각 기관은 `ParentOfficeID`로 상위 기관을 가리키며, 공소청 → 광역공소청 → 지방공소청 → 지청 순서로 이어진다. 별표에는 지방공소청이 어느 광역공소청에 속하는지 적혀 있지 않으므로, [공소청 홈페이지의 조직구조](https://www.spo.go.kr/site/spo/06/10602010100002018100812.jsp)에 따라 연결했다. 이 관계는 지방법원과 고등법원의 소속 관계와 같다. 주소와 우편번호는 [공소청 홈페이지의 전국공소청찾기](https://www.spo.go.kr/site/spo/ex/announce/spoSearchList.do)를 따르며, 별표에 없는 공소청 자체의 소재지도 이 주소를 기준으로 기록한다. 대표전화는 모든 기관이 콜센터 1301로만 안내되어 있어 `Phone`에 넣지 않는다. 규정 개정 시 이 데이터도 함께 갱신해야 한다.

## 중대범죄수사청 정보

[`InvestigationAgencies`](investigations.go)는 중대범죄수사청 본청과 6개 지방중대범죄수사청(서울·수원·대전·대구·부산·광주)을 제공한다. 명칭·위치·관할구역은 [`documents/pdfs/prosecution`](documents/pdfs/prosecution)에 둔 「중대범죄수사청과 그 소속기관 직제」 별표 1을, 주소는 [중대범죄수사청 홈페이지](https://www.sci.go.kr/content.do?key=101041)를 따르며, 전화번호는 두지 않는다. 지방수사청은 본청의 소속기관이므로 `ParentAgencyID`로 본청을 가리키며, 현판 등에 쓰는 약칭(예: `서울지방수사청`)은 `ShortName`에 둔다. 수원지방수사청은 2027. 5. 이후 성남시 분당구로 이전할 예정이므로 그때 주소를 갱신해야 한다.

## 사건 정보

[`Case`](cases.go)는 소송사건 한 건의 기본 정보를 담는다.

- **사건번호**(`CaseNumber`): 연도(4자리) + 사건부호(한글, 예: `가합`·`카단`·`고단`) + 일련번호. `ParseCaseNumber("2026가합12345")`로 읽고 `String()`으로 같은 형식으로 쓴다. 사건번호는 법원마다 따로 매기므로 법원과 함께 있어야 사건을 특정할 수 있다.
- **사건명**(`Title`): 사용자가 직접 입력한다.
- **사건 종류**(`Category`): 민사·행정·가사·민사신청·형사.
- **당사자**(`Parties`): `Party{Role, Name, Client}`의 목록으로, 지위와 이름 모두 사용자가 직접 입력한다. 보통 원고·피고(형사는 피고인·검사) 2인이지만 1인이나 3인 이상도 될 수 있으며, 입력 순서를 유지한다.
- **의뢰인**(`Party.Client`): 당사자마다 의뢰인 여부를 표시한다. 같은 지위의 당사자가 여럿이어도 일부만 의뢰인일 수 있으므로(예: 원고 3명 중 원고1·원고3) 지위 단위가 아니라 사람 단위로 표시하며, 사건마다 의뢰인이 한 명 이상 있어야 한다. 또한 지위가 다른 당사자를 함께 대리하면 이해충돌이 생기므로 의뢰인은 모두 같은 지위여야 한다. 같은 지위인지는 `PartySide`로 판단하며 번호·괄호 표기·띄어쓰기는 무시한다(`원고1`·`원고3` → 원고, `피고(반소원고)` → 피고). `Case.Clients()`로 의뢰인만 꺼낼 수 있다.
- **계속 법원·재판부**(`Court`): `Courts` 목록의 법원 ID와 그 법원 `Panels`에 있는 재판부 이름. 형사사건을 포함한 모든 사건에 필수이다. 재판부 ID는 구성표 개정 때 다시 매겨질 수 있으므로 사건에는 재판부 이름을 저장한다.

`Case.Validate()`는 위 규칙을 검사하고 문제를 모두 모아 돌려준다.

## 사건목록 웹 서비스

[`cmd/casepulse-web`](cmd/casepulse-web)은 사건목록을 조회하는 웹 서비스다. Go 서버가 API와 화면을 함께 제공한다.

```sh
# 화면 빌드 (Node.js 20.19+ 필요, 처음 한 번과 화면을 고친 뒤마다)
cd cmd/casepulse-web/web && npm install && npm run build && cd -

go run ./cmd/casepulse-web            # http://localhost:8080
go run ./cmd/casepulse-web -addr :9000
```

화면을 개발할 때는 Go 서버를 켜 둔 채 `cmd/casepulse-web/web`에서 `npm run dev`를 실행한다. Vite 개발 서버(http://localhost:5173)가 고친 내용을 바로 반영하고 `/api` 요청은 Go 서버로 넘긴다.

- **데이터**: 아직 데이터베이스가 없으므로 시작할 때 [`dummy.go`](cmd/casepulse-web/dummy.go)가 고정 시드로 더미 사건 240건을 만든다. 당사자·사건번호는 지어낸 값이고, 법원·재판부는 `Courts` 목록에서 사건 종류에 맞는 것(민사→민사부, 형사→형사부, 가사→가정법원 등)을 고르며, 모든 사건은 `Case.Validate()`를 통과한다.
- **API**: `GET /api/cases`(전체 목록), `GET /api/cases/{id}`(한 건), `POST /api/cases`(사건 등록), `GET /api/courts`(등록 화면용 법원·재판부 목록). 사건은 법원 이름·소재지·주소를 붙인 JSON으로 돌려준다. 등록 요청은 `Case.Validate()`로 검사하고, 문제가 있으면 400과 함께 오류 메시지 목록을, 같은 법원에 같은 사건번호가 있으면 409를 돌려준다. 등록한 사건은 메모리에만 있으므로 서버를 다시 켜면 사라진다.
- **화면**: [`web/`](cmd/casepulse-web/web)의 React(JSX, TypeScript 없음) 앱이며 Vite로 빌드한다. UI는 [shadcn/ui](https://ui.shadcn.com)(Radix 기반, Nova 프리셋)와 Tailwind CSS v4로 만들고, [Vite 설치 안내](https://ui.shadcn.com/docs/installation/vite)를 따르되 TypeScript 대신 JavaScript 방식(`components.json`의 `tsx: false`, `jsconfig.json`의 `@/` 별칭)을 쓴다. 컴포넌트는 `npx shadcn@latest add <이름>`으로 `src/components/ui/`에 추가하며, 화면 테마(밝은 화면·어두운 화면·시스템 설정)는 shadcn 다크 모드 안내의 ThemeProvider 방식을 따른다. 빌드 결과물 `web/dist`는 git에 올리지 않고 Go 실행 파일에 포함(`go:embed`)한다. 로고와 파비콘은 저장소 맨 위 [`casepulse-logo/`](casepulse-logo) 폴더가 원본이며, Vite가 이 폴더를 사이트 루트(`/favicon.ico`, `/casepulse-logo.svg` 등)로 그대로 내보낸다. 메뉴줄 로고는 밝은 화면에서 `casepulse-logo.svg`, 어두운 화면에서 `casepulse-logo-dark.svg`를 쓴다. 빌드하지 않고 서버를 켜면 `/`에 빌드 안내가 나오고 API는 그대로 동작한다.
- **기능**: 사건번호·사건명·당사자·법원·재판부 통합 검색(`/` 키로 이동), 사건 종류·접수 연도·계속 법원 필터, 의뢰인·다른 당사자 구분 표시, 열 정렬, 무한 스크롤(50건씩 이어서 표시), 사건 상세 패널(당사자 전체, 법원 주소), 어두운 화면.
- **사건 등록**: 사건목록 툴바의 `사건 등록` 버튼, 메뉴줄의 `새 사건` 버튼, `N` 키로 등록 탭을 새로 연다. 여러 건을 동시에 작성할 수 있고, 탭을 옮겨도 입력 내용이 남는다. 사건번호 칸에 `2026가합12345`를 붙여 넣으면 연도·사건부호·일련번호로 나뉘어 들어가고, 당사자마다 의뢰인 여부를 체크하며, 법원·재판부는 이름으로 검색해 고른다(재판부 정보가 있는 법원만). `Ctrl+Enter`로 등록하면 탭이 닫히고 사건목록에서 새 사건의 상세 패널이 열린다. 저장하지 않은 탭을 닫거나 브라우저를 닫으려 하면 확인을 받는다.
