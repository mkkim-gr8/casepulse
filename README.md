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

모든 목록은 패키지가 로드될 때 한 번 만들어지며, 표 사이의 값이 맞지 않으면(없는 법원 ID 참조, 주소 누락 등) 패닉을 일으켜 잘못된 데이터가 배포되지 않도록 한다. `go test -v -run PrintsTree`로 법원·공소청 트리를 출력해 데이터를 검토할 수 있다.

## 법원 정보

법원 기관은 법원 단계, 관할 분야, 본원·지원·시·군법원 구분, 상위 법원, 소재지와 관할 지역을 별도 상세정보로 표현한다. [`Courts`](courts.go)는 [`documents/pdfs/court`](documents/pdfs/court)에 둔 법원 명칭·소재지 및 관할구역 별표를 바탕으로 215개 법원 목록을 제공한다. `Seat`은 별표의 소재지이며(다만 광주는 전남광주통합특별시 출범에 맞춰 `전남광주통합특별시`로, 옛 광주광역시 관할은 통합특별시의 5개 구로 정규화한다)(시·군법원은 별표 2의 읍·시 단위 소재지, 예: 철원군법원은 `갈말읍`), 주소는 도로명 주소 형식으로 정규화하고 전화번호는 두지 않는다. 또한 우편번호(`PostalCode`)와 주소(`Address`)는 [대법원 홈페이지 각급법원 안내](https://www.scourt.go.kr/region/location/RegionLocationListAction.work)를 따른다. 별표에 신설되었으나 홈페이지에 아직 없는 전주가정법원과 그 지원은 주소를 비워 둔다. 관할구역 별표의 개정 시점은 자료별로 다르므로 최신 법령과 대조해 갱신해야 한다.

[`documents/pdfs/court/재판부구성표.pdf`](documents/pdfs/court/재판부구성표.pdf)에 나온 재판부는 소속 `Court`의 하위 `Panels`에 둔다. 합의부·단독 재판부와 조정·재정·회생·파산 등 표에 기재된 재판 단위를 재판부명으로 기록한다. 원외재판부는 별도의 고등법원 `Court`로 표현하고, 예를 들어 서울고등법원 인천 원외재판부는 `서울고등법원(인천)`이라는 이름의 고등법원으로 등록한다. 원외 고등법원은 본원 고등법원과 동급이므로 둘 다 대법원을 상위 기관으로 가리키며, 인천 재판부들은 `서울고등법원(인천)`의 `Panels`에 직접 둔다. 판사 개인 명단과 대괄호 안의 사건 분야 안내는 포함하지 않는다. 구성표 개정 시 이 데이터도 함께 갱신해야 한다.


## 공소청 정보

[`ProsecutionOffices`](prosecutions.go)는 [`documents/pdfs/prosecution`](documents/pdfs/prosecution)에 둔 「공소청의 위치와 각급 공소청 및 지청의 명칭 및 위치에 관한 규정」 별표(2026. 9. 29. 개정)를 바탕으로 공소청과 그 아래 66개 기관(광역공소청 6, 지방공소청 18, 지청 42)을 제공한다. 각 기관은 `ParentOfficeID`로 상위 기관을 가리키며, 공소청 → 광역공소청 → 지방공소청 → 지청 순서로 이어진다. 별표에는 지방공소청이 어느 광역공소청에 속하는지 적혀 있지 않으므로, [공소청 홈페이지의 조직구조](https://www.spo.go.kr/site/spo/06/10602010100002018100812.jsp)에 따라 연결했다. 이 관계는 지방법원과 고등법원의 소속 관계와 같다. 주소와 우편번호는 [공소청 홈페이지의 전국공소청찾기](https://www.spo.go.kr/site/spo/ex/announce/spoSearchList.do)를 따르며, 별표에 없는 공소청 자체의 소재지도 이 주소를 기준으로 기록한다. 대표전화는 모든 기관이 콜센터 1301로만 안내되어 있어 `Phone`에 넣지 않는다. 규정 개정 시 이 데이터도 함께 갱신해야 한다.

## 중대범죄수사청 정보

[`InvestigationAgencies`](investigations.go)는 중대범죄수사청 본청과 6개 지방중대범죄수사청(서울·수원·대전·대구·부산·광주)을 제공한다. 명칭·위치·관할구역은 [`documents/pdfs/prosecution`](documents/pdfs/prosecution)에 둔 「중대범죄수사청과 그 소속기관 직제」 별표 1을, 주소는 [중대범죄수사청 홈페이지](https://www.sci.go.kr/content.do?key=101041)를 따르며, 전화번호는 두지 않는다. 지방수사청은 본청의 소속기관이므로 `ParentAgencyID`로 본청을 가리키며, 현판 등에 쓰는 약칭(예: `서울지방수사청`)은 `ShortName`에 둔다. 수원지방수사청은 2027. 5. 이후 성남시 분당구로 이전할 예정이므로 그때 주소를 갱신해야 한다.
