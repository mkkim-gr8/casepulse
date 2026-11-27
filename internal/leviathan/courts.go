package leviathan

import (
	"fmt"
	"strings"
)

// 이 파일은 전국 법원 목록을 만드는 로직과, 법원의 종류·소속 관계를 정하는 구조 표를 담습니다.
//
// 법원 목록은 여러 파일의 표를 합쳐 만듭니다.
//   - courts.go: 법원의 이름·단계·소속·소재지·관할 지역(이 파일)
//   - courts_municipal.go: 시·군법원 이름의 시/군 구분과 소재지
//   - courts_panels.go: 법원별 재판부 이름
//   - courts_addresses.go: 법원별 우편번호·주소
//
// 원본 자료는 documents/pdfs/court의 「각급 법원의 설치와 관할구역에 관한 법률」 별표,
// 재판부구성표, 그리고 대법원 홈페이지 각급법원 안내입니다.

// courtDirectory는 패키지가 처음 로드될 때 buildCourtDirectory로 한 번 만든 법원 원본 목록입니다.
// 외부에는 직접 내보내지 않고 Courts가 복사본을 돌려줍니다.
var courtDirectory = buildCourtDirectory()

// Courts는 전국 법원 목록(대법원부터 시·군법원까지)을 반환합니다.
//
// 순서는 대법원 → 고등법원 → 원외재판부 → 특허법원 → 지방법원 → 지원 → 가정법원 → 가정법원 지원
// → 행정법원 → 회생법원 → 시·군법원입니다. 특정 법원은 ID나 Name으로 찾으면 됩니다.
//
// 반환값은 내부 원본의 깊은 복사본입니다. 호출한 쪽에서 목록의 순서를 바꾸거나
// Jurisdictions·Panels 같은 내부 슬라이스를 고쳐도 패키지가 보관한 원본은 바뀌지 않습니다.
func Courts() []Court {
	// copy는 구조체 값만 복사하므로 Jurisdictions와 Panels는 여전히 원본 배열을 가리킵니다.
	// 그래서 아래 반복문에서 두 슬라이스를 새로 만들어 원본과의 연결을 끊습니다.
	courts := make([]Court, len(courtDirectory))
	copy(courts, courtDirectory)
	for i := range courts {
		courts[i].Jurisdictions = append([]string(nil), courtDirectory[i].Jurisdictions...)
		// CourtPanel에는 슬라이스 필드가 없으므로 Panels는 한 단계 복사만으로 원본과 분리됩니다.
		courts[i].Panels = append([]CourtPanel(nil), courtDirectory[i].Panels...)
	}
	return courts
}

// buildCourtDirectory는 이 파일과 보조 파일의 표를 합쳐 전국 법원 목록을 만듭니다.
// 패키지 초기화 때 courtDirectory를 만들기 위해 한 번만 호출됩니다.
//
// 처리 순서는 다음과 같습니다.
//  1. 대법원, 고등법원, 원외재판부, 특허법원을 추가합니다.
//  2. 지방법원과 그 지원을 추가합니다.
//  3. 가정법원과 그 지원, 행정법원, 회생법원을 추가합니다.
//  4. 시·군법원을 추가합니다(이름과 소재지는 courts_municipal.go의 표로 정합니다).
//  5. 재판부구성표(courts_panels.go)의 재판부를 소속 법원의 Panels에 붙입니다.
//  6. 주소 표(courts_addresses.go)의 우편번호와 주소를 채웁니다.
//
// 법원 ID는 "court-<종류>-<짧은 식별자>" 형식입니다(예: "court-district-incheon",
// "court-branch-incheon-bucheon"). 하위 법원은 ParentCourtID로 바로 위 법원의 ID를 가리킵니다.
// 예를 들어 인천지방법원 부천지원은 인천지방법원을, 인천지방법원 김포시법원은 부천지원을 가리킵니다.
//
// 원외재판부는 재판부가 아니라 독립된 고등법원 항목(예: "서울고등법원(인천)")으로 만들어,
// 원외재판부 소속 재판부를 그 항목의 Panels에 둡니다. 원외 고등법원은 본원 고등법원과
// 동급으로 보고 모두 대법원을 상위 법원으로 둡니다.
//
// 표의 값이 서로 맞지 않으면(없는 법원 ID 참조, 주소 누락 등) 패닉을 일으켜
// 잘못된 데이터가 조용히 배포되지 않도록 합니다.
func buildCourtDirectory() []Court {
	var courts []Court
	// add는 법원 한 곳을 courts 끝에 추가하는 내부 도우미입니다.
	//   - id: 법원 ID, parentID: 상위 법원 ID(대법원은 "")
	//   - seat: 별표에 적힌 소재지
	//   - jurisdictions: 관할 지역(가변 인자이므로 add(..., "A시", "B군")처럼 여러 개를 넘깁니다)
	// 우편번호·주소·재판부는 여기서 넣지 않고 5·6단계에서 채웁니다.
	add := func(id, name string, level CourtLevel, courtType CourtType, unit CourtUnitType, parentID, seat string, jurisdictions ...string) {
		courts = append(courts, Court{
			Institution: Institution{ID: id, Name: name, Type: InstitutionTypeCourt},
			Level:       level, CourtType: courtType, UnitType: unit,
			ParentCourtID: parentID, Seat: seat, Jurisdictions: jurisdictions,
		})
	}

	// ── 1단계: 대법원, 고등법원, 원외재판부, 특허법원 ──

	// 대법원은 최상위 법원이므로 상위 법원 ID가 없고, 전국을 관할하므로 관할 지역을 "전국"으로 적습니다.
	add("court-supreme", "대법원", CourtLevelSupreme, CourtTypeGeneral, CourtUnitTypeMain, "", "서울특별시", "전국")

	// 고등법원(별표 1). 각 행은 {짧은 식별자, 공식 명칭, 소재지}이며 모두 대법원 소속입니다.
	// 고등법원의 관할 지역은 소속 지방법원들의 관할을 합친 것이므로 따로 적지 않습니다.
	for _, court := range []struct {
		id, name, seat string
	}{
		{"seoul", "서울고등법원", "서울특별시"},
		{"daejeon", "대전고등법원", "대전광역시"},
		{"daegu", "대구고등법원", "대구광역시"},
		{"busan", "부산고등법원", "부산광역시"},
		{"gwangju", "광주고등법원", "전남광주통합특별시"},
		{"suwon", "수원고등법원", "수원시"},
	} {
		add("court-high-"+court.id, court.name, CourtLevelHigh, CourtTypeGeneral, CourtUnitTypeMain, "court-supreme", court.seat)
	}

	// 원외재판부(재판부구성표, 대법원 홈페이지 각급법원 안내). 각 행은 {짧은 식별자, 이름, 상위 법원 ID, 소재지}입니다.
	// 이름은 "<본원 고등법원>(<소재 도시>)" 형식이며, 짧은 식별자는 "<본원>-<소재 도시>" 형식입니다.
	for _, court := range []struct {
		id, name, parentID, seat string
	}{
		{"seoul-incheon", "서울고등법원(인천)", "court-supreme", "인천광역시"},
		{"seoul-chuncheon", "서울고등법원(춘천)", "court-supreme", "춘천시"},
		{"daejeon-cheongju", "대전고등법원(청주)", "court-supreme", "청주시"},
		{"busan-ulsan", "부산고등법원(울산)", "court-supreme", "울산광역시"},
		{"busan-changwon", "부산고등법원(창원)", "court-supreme", "창원시"},
		{"gwangju-jeonju", "광주고등법원(전주)", "court-supreme", "전주시"},
		{"gwangju-jeju", "광주고등법원(제주)", "court-supreme", "제주시"},
	} {
		add("court-high-"+court.id, court.name, CourtLevelHigh, CourtTypeGeneral, CourtUnitTypeMain, court.parentID, court.seat)
	}

	// 특허법원(별표 1·4)은 고등법원 단계의 전문 법원으로, 전국의 특허 사건을 맡습니다.
	add("court-patent", "특허법원", CourtLevelHigh, CourtTypePatent, CourtUnitTypeMain, "court-supreme", "대전광역시", "전국")

	// ── 2단계: 지방법원과 지원 ──

	// district는 지방법원 표의 한 행입니다. high는 소속 고등법원의 짧은 식별자입니다(예: "seoul").
	type district struct {
		id, name, high, seat string
		jurisdictions        []string
	}

	// 지방법원(별표 1·3). 관할 지역은 Court.Jurisdictions의 규칙대로, 시·도 이름을 첫 항목에만 붙입니다.
	// 예: 서울중앙지방법원의 {"서울특별시 종로구", "중구", ...}는 서울특별시 종로구·중구 등입니다.
	//
	// 별표는 광주를 "광주광역시"로, 강원을 "강원도"로 적고 있으나, 현재 행정구역 명칭에 맞춰
	// "전남광주통합특별시", "강원특별자치도"로 정규화합니다. 옛 광주광역시 관할은
	// 통합특별시의 동구·서구·남구·북구·광산구로 풀어 적어, 옛 전라남도 지역과 섞이지 않게 합니다.
	districts := []district{
		{"seoul-central", "서울중앙지방법원", "seoul", "서울특별시", []string{"서울특별시 종로구", "중구", "강남구", "서초구", "관악구", "동작구"}},
		{"seoul-east", "서울동부지방법원", "seoul", "서울특별시", []string{"서울특별시 성동구", "광진구", "강동구", "송파구"}},
		{"seoul-south", "서울남부지방법원", "seoul", "서울특별시", []string{"서울특별시 영등포구", "강서구", "양천구", "구로구", "금천구"}},
		{"seoul-north", "서울북부지방법원", "seoul", "서울특별시", []string{"서울특별시 동대문구", "중랑구", "성북구", "도봉구", "강북구", "노원구"}},
		{"seoul-west", "서울서부지방법원", "seoul", "서울특별시", []string{"서울특별시 서대문구", "마포구", "은평구", "용산구"}},
		{"uijeongbu", "의정부지방법원", "seoul", "의정부시", []string{"의정부시", "동두천시", "양주시", "연천군", "포천시", "강원특별자치도 철원군"}},
		{"incheon", "인천지방법원", "seoul", "인천광역시", []string{"인천광역시"}},
		{"suwon", "수원지방법원", "suwon", "수원시", []string{"수원시", "오산시", "용인시", "화성시"}},
		{"chuncheon", "춘천지방법원", "seoul", "춘천시", []string{"춘천시", "화천군", "양구군", "인제군", "홍천군"}},
		{"daejeon", "대전지방법원", "daejeon", "대전광역시", []string{"대전광역시", "세종특별자치시", "금산군"}},
		{"cheongju", "청주지방법원", "daejeon", "청주시", []string{"청주시", "진천군", "보은군", "괴산군", "증평군"}},
		{"daegu", "대구지방법원", "daegu", "대구광역시", []string{"대구광역시 중구", "동구", "남구", "북구", "수성구", "영천시", "경산시", "칠곡군", "청도군"}},
		{"busan", "부산지방법원", "busan", "부산광역시", []string{"부산광역시 중구", "동구", "영도구", "부산진구", "동래구", "연제구", "금정구"}},
		{"ulsan", "울산지방법원", "busan", "울산광역시", []string{"울산광역시", "양산시"}},
		{"changwon", "창원지방법원", "busan", "창원시", []string{"창원시 의창구", "성산구", "진해구", "김해시"}},
		{"gwangju", "광주지방법원", "gwangju", "전남광주통합특별시", []string{"전남광주통합특별시 동구", "서구", "남구", "북구", "광산구", "나주시", "화순군", "장성군", "담양군", "곡성군", "영광군"}},
		{"jeonju", "전주지방법원", "gwangju", "전주시", []string{"전주시", "김제시", "완주군", "임실군", "진안군", "무주군"}},
		{"jeju", "제주지방법원", "gwangju", "제주시", []string{"제주시", "서귀포시"}},
	}

	// districtIDs는 짧은 식별자(예: "incheon")를 법원 ID(예: "court-district-incheon")로 바꾸는 표입니다.
	// 지원과 시·군법원을 등록할 때 상위 지방법원의 ID를 찾는 데 씁니다.
	districtIDs := make(map[string]string, len(districts))
	// districtNames는 짧은 식별자를 지방법원 공식 이름으로 바꾸는 표입니다.
	// 시·군법원 이름 앞에 소속 지방법원 이름을 붙일 때 씁니다(예: "인천지방법원 김포시법원").
	districtNames := make(map[string]string, len(districts))
	for _, court := range districts {
		id := "court-district-" + court.id
		districtIDs[court.id] = id
		districtNames[court.id] = court.name
		add(id, court.name, CourtLevelDistrict, CourtTypeGeneral, CourtUnitTypeMain, "court-high-"+court.high, court.seat, court.jurisdictions...)
	}

	// branch는 지원 표의 한 행입니다. district는 소속 지방법원의 짧은 식별자입니다(예: "incheon").
	type branch struct {
		id, name, district, seat string
		jurisdictions            []string
	}

	// 지방법원 지원(별표 1·3). 짧은 식별자는 "<지방법원>-<지원 소재지>" 형식입니다(예: "incheon-bucheon").
	branches := []branch{
		{"uijeongbu-goyang", "의정부지방법원 고양지원", "uijeongbu", "고양시", []string{"고양시", "파주시"}},
		{"uijeongbu-namyangju", "의정부지방법원 남양주지원", "uijeongbu", "남양주시", []string{"남양주시", "구리시", "가평군"}},
		{"incheon-bucheon", "인천지방법원 부천지원", "incheon", "부천시", []string{"부천시"}},
		{"suwon-seongnam", "수원지방법원 성남지원", "suwon", "성남시", []string{"성남시", "하남시", "광주시"}},
		{"suwon-yeoju", "수원지방법원 여주지원", "suwon", "여주시", []string{"이천시", "여주시", "양평군"}},
		{"suwon-pyeongtaek", "수원지방법원 평택지원", "suwon", "평택시", []string{"평택시", "안성시"}},
		{"suwon-ansan", "수원지방법원 안산지원", "suwon", "안산시", []string{"안산시", "광명시", "시흥시"}},
		{"suwon-anyang", "수원지방법원 안양지원", "suwon", "안양시", []string{"안양시", "과천시", "의왕시", "군포시"}},
		{"chuncheon-gangneung", "춘천지방법원 강릉지원", "chuncheon", "강릉시", []string{"강릉시", "동해시", "삼척시", "속초시", "양양군", "고성군"}},
		{"chuncheon-wonju", "춘천지방법원 원주지원", "chuncheon", "원주시", []string{"원주시", "횡성군"}},
		{"chuncheon-sokcho", "춘천지방법원 속초지원", "chuncheon", "속초시", []string{"속초시", "양양군", "고성군"}},
		{"chuncheon-yeongwol", "춘천지방법원 영월지원", "chuncheon", "영월읍", []string{"태백시", "영월군", "정선군", "평창군"}},
		{"daejeon-hongseong", "대전지방법원 홍성지원", "daejeon", "홍성읍", []string{"보령시", "홍성군", "예산군", "서천군"}},
		{"daejeon-gongju", "대전지방법원 공주지원", "daejeon", "공주시", []string{"공주시", "청양군"}},
		{"daejeon-nonsan", "대전지방법원 논산지원", "daejeon", "논산시", []string{"논산시", "계룡시", "부여군"}},
		{"daejeon-seosan", "대전지방법원 서산지원", "daejeon", "서산시", []string{"서산시", "당진시", "태안군"}},
		{"daejeon-cheonan", "대전지방법원 천안지원", "daejeon", "천안시", []string{"천안시", "아산시"}},
		{"cheongju-chungju", "청주지방법원 충주지원", "cheongju", "충주시", []string{"충주시", "음성군"}},
		{"cheongju-jecheon", "청주지방법원 제천지원", "cheongju", "제천시", []string{"제천시", "단양군"}},
		{"cheongju-yeongdong", "청주지방법원 영동지원", "cheongju", "영동읍", []string{"영동군", "옥천군"}},
		{"daegu-seobu", "대구지방법원 서부지원", "daegu", "대구광역시", []string{"대구광역시 서구", "달서구", "달성군", "성주군", "고령군"}},
		{"daegu-andong", "대구지방법원 안동지원", "daegu", "안동시", []string{"안동시", "영주시", "봉화군"}},
		{"daegu-gyeongju", "대구지방법원 경주지원", "daegu", "경주시", []string{"경주시"}},
		{"daegu-pohang", "대구지방법원 포항지원", "daegu", "포항시", []string{"포항시", "울릉군"}},
		{"daegu-gimcheon", "대구지방법원 김천지원", "daegu", "김천시", []string{"김천시", "구미시"}},
		{"daegu-sangju", "대구지방법원 상주지원", "daegu", "상주시", []string{"상주시", "문경시", "예천군"}},
		{"daegu-uiseong", "대구지방법원 의성지원", "daegu", "의성읍", []string{"의성군", "군위군", "청송군"}},
		{"daegu-yeongdeok", "대구지방법원 영덕지원", "daegu", "영덕읍", []string{"영덕군", "영양군", "울진군"}},
		{"busan-dongbu", "부산지방법원 동부지원", "busan", "부산광역시", []string{"부산광역시 해운대구", "남구", "수영구", "기장군"}},
		{"busan-seobu", "부산지방법원 서부지원", "busan", "부산광역시", []string{"부산광역시 서구", "북구", "사상구", "사하구", "강서구"}},
		{"changwon-masan", "창원지방법원 마산지원", "changwon", "창원시", []string{"창원시 마산합포구", "마산회원구", "함안군", "의령군"}},
		{"changwon-jinju", "창원지방법원 진주지원", "changwon", "진주시", []string{"진주시", "사천시", "남해군", "하동군", "산청군"}},
		{"changwon-tongyeong", "창원지방법원 통영지원", "changwon", "통영시", []string{"통영시", "거제시", "고성군"}},
		{"changwon-miryang", "창원지방법원 밀양지원", "changwon", "밀양시", []string{"밀양시", "창녕군"}},
		{"changwon-geochang", "창원지방법원 거창지원", "changwon", "거창읍", []string{"거창군", "함양군", "합천군"}},
		{"gwangju-mokpo", "광주지방법원 목포지원", "gwangju", "목포시", []string{"목포시", "무안군", "신안군", "함평군", "영암군"}},
		{"gwangju-jangheung", "광주지방법원 장흥지원", "gwangju", "장흥읍", []string{"장흥군", "강진군"}},
		{"gwangju-suncheon", "광주지방법원 순천지원", "gwangju", "순천시", []string{"순천시", "여수시", "광양시", "구례군", "고흥군", "보성군"}},
		{"gwangju-haenam", "광주지방법원 해남지원", "gwangju", "해남읍", []string{"해남군", "완도군", "진도군"}},
		{"jeonju-gunsan", "전주지방법원 군산지원", "jeonju", "군산시", []string{"군산시", "익산시"}},
		{"jeonju-jeongeup", "전주지방법원 정읍지원", "jeonju", "정읍시", []string{"정읍시", "부안군", "고창군"}},
		{"jeonju-namwon", "전주지방법원 남원지원", "jeonju", "남원시", []string{"남원시", "장수군", "순창군"}},
	}

	// branchIDs는 지원의 짧은 식별자를 법원 ID로 바꾸는 표입니다.
	// 지원에 소속된 시·군법원을 등록할 때 상위 법원 ID로 씁니다.
	branchIDs := make(map[string]string, len(branches))
	for _, court := range branches {
		id := "court-branch-" + court.id
		branchIDs[court.id] = id
		add(id, court.name, CourtLevelDistrict, CourtTypeGeneral, CourtUnitTypeSupport, districtIDs[court.district], court.seat, court.jurisdictions...)
	}

	// ── 3단계: 가정법원과 그 지원, 행정법원, 회생법원 ──

	// 가정법원(별표 1·5). 각 행의 뜻은 지방법원 표와 같고, high는 소속 고등법원의 짧은 식별자입니다.
	// 전주가정법원은 별표 5(2026. 3. 17. 개정)에 신설된 법원으로, 대법원 홈페이지에는 아직 없습니다.
	familyCourts := []struct {
		id, name, high, seat string
		jurisdictions        []string
	}{
		{"seoul", "서울가정법원", "seoul", "서울특별시", []string{"서울특별시"}},
		{"incheon", "인천가정법원", "seoul", "인천광역시", []string{"인천광역시"}},
		{"suwon", "수원가정법원", "suwon", "수원시", []string{"수원시", "오산시", "용인시", "화성시"}},
		{"daejeon", "대전가정법원", "daejeon", "대전광역시", []string{"대전광역시", "세종특별자치시", "금산군"}},
		{"daegu", "대구가정법원", "daegu", "대구광역시", []string{"대구광역시", "영천시", "경산시", "칠곡군", "성주군", "고령군", "청도군"}},
		{"busan", "부산가정법원", "busan", "부산광역시", []string{"부산광역시"}},
		{"ulsan", "울산가정법원", "busan", "울산광역시", []string{"울산광역시", "양산시"}},
		{"gwangju", "광주가정법원", "gwangju", "전남광주통합특별시", []string{"전남광주통합특별시 동구", "서구", "남구", "북구", "광산구", "나주시", "화순군", "장성군", "담양군", "곡성군", "영광군"}},
		{"jeonju", "전주가정법원", "gwangju", "전주시", []string{"전주시", "김제시", "완주군", "임실군", "진안군", "무주군"}},
	}

	// familyIDs는 가정법원의 짧은 식별자를 법원 ID로 바꾸는 표로, 가정법원 지원의 상위 법원 ID에 씁니다.
	familyIDs := make(map[string]string, len(familyCourts))
	for _, court := range familyCourts {
		id := "court-family-" + court.id
		familyIDs[court.id] = id
		add(id, court.name, CourtLevelDistrict, CourtTypeFamily, CourtUnitTypeMain, "court-high-"+court.high, court.seat, court.jurisdictions...)
	}

	// 가정법원 지원(별표 1·5). 같은 건물의 지방법원 지원과 관할이 같더라도 별도의 법원이므로 따로 등록합니다.
	// parent는 소속 가정법원의 짧은 식별자입니다. 소재지는 별표 1을 따릅니다(예: 홍성지원은 "홍성읍").
	familyBranches := []struct {
		id, name, parent, seat string
		jurisdictions          []string
	}{
		{"incheon-bucheon", "인천가정법원 부천지원", "incheon", "부천시", []string{"부천시"}},
		{"suwon-seongnam", "수원가정법원 성남지원", "suwon", "성남시", []string{"성남시", "하남시", "광주시"}},
		{"suwon-yeoju", "수원가정법원 여주지원", "suwon", "여주시", []string{"이천시", "여주시", "양평군"}},
		{"suwon-pyeongtaek", "수원가정법원 평택지원", "suwon", "평택시", []string{"평택시", "안성시"}},
		{"suwon-ansan", "수원가정법원 안산지원", "suwon", "안산시", []string{"안산시", "광명시", "시흥시"}},
		{"suwon-anyang", "수원가정법원 안양지원", "suwon", "안양시", []string{"안양시", "과천시", "의왕시", "군포시"}},
		{"daejeon-hongseong", "대전가정법원 홍성지원", "daejeon", "홍성읍", []string{"보령시", "홍성군", "예산군", "서천군"}},
		{"daejeon-gongju", "대전가정법원 공주지원", "daejeon", "공주시", []string{"공주시", "청양군"}},
		{"daejeon-nonsan", "대전가정법원 논산지원", "daejeon", "논산시", []string{"논산시", "계룡시", "부여군"}},
		{"daejeon-seosan", "대전가정법원 서산지원", "daejeon", "서산시", []string{"서산시", "당진시", "태안군"}},
		{"daejeon-cheonan", "대전가정법원 천안지원", "daejeon", "천안시", []string{"천안시", "아산시"}},
		{"daegu-andong", "대구가정법원 안동지원", "daegu", "안동시", []string{"안동시", "영주시", "봉화군"}},
		{"daegu-gyeongju", "대구가정법원 경주지원", "daegu", "경주시", []string{"경주시"}},
		{"daegu-pohang", "대구가정법원 포항지원", "daegu", "포항시", []string{"포항시", "울릉군"}},
		{"daegu-gimcheon", "대구가정법원 김천지원", "daegu", "김천시", []string{"김천시", "구미시"}},
		{"daegu-sangju", "대구가정법원 상주지원", "daegu", "상주시", []string{"상주시", "문경시", "예천군"}},
		{"daegu-uiseong", "대구가정법원 의성지원", "daegu", "의성읍", []string{"의성군", "군위군", "청송군"}},
		{"daegu-yeongdeok", "대구가정법원 영덕지원", "daegu", "영덕읍", []string{"영덕군", "영양군", "울진군"}},
		{"gwangju-mokpo", "광주가정법원 목포지원", "gwangju", "목포시", []string{"목포시", "무안군", "신안군", "함평군", "영암군"}},
		{"gwangju-jangheung", "광주가정법원 장흥지원", "gwangju", "장흥읍", []string{"장흥군", "강진군"}},
		{"gwangju-suncheon", "광주가정법원 순천지원", "gwangju", "순천시", []string{"순천시", "여수시", "광양시", "구례군", "고흥군", "보성군"}},
		{"gwangju-haenam", "광주가정법원 해남지원", "gwangju", "해남읍", []string{"해남군", "완도군", "진도군"}},
		{"jeonju-gunsan", "전주가정법원 군산지원", "jeonju", "군산시", []string{"군산시", "익산시"}},
		{"jeonju-jeongeup", "전주가정법원 정읍지원", "jeonju", "정읍시", []string{"정읍시", "부안군", "고창군"}},
		{"jeonju-namwon", "전주가정법원 남원지원", "jeonju", "남원시", []string{"남원시", "장수군", "순창군"}},
	}
	for _, court := range familyBranches {
		add("court-family-branch-"+court.id, court.name, CourtLevelDistrict, CourtTypeFamily, CourtUnitTypeSupport, familyIDs[court.parent], court.seat, court.jurisdictions...)
	}

	// 행정법원(별표 1·6). 서울행정법원만 있으며, 그 밖의 지역은 지방법원 본원이 행정사건을 맡습니다.
	add("court-administrative-seoul", "서울행정법원", CourtLevelDistrict, CourtTypeAdministrative, CourtUnitTypeMain, "court-high-seoul", "서울특별시", "서울특별시")

	// 회생법원(별표 1·10). 각 행은 {짧은 식별자, 공식 명칭, 소속 고등법원의 짧은 식별자, 소재지}입니다.
	// 관할 지역은 별표 10의 구역이 길어 아직 적지 않았습니다.
	for _, court := range []struct{ id, name, parent, seat string }{
		{"seoul", "서울회생법원", "seoul", "서울특별시"},
		{"suwon", "수원회생법원", "suwon", "수원시"},
		{"daejeon", "대전회생법원", "daejeon", "대전광역시"},
		{"daegu", "대구회생법원", "daegu", "대구광역시"},
		{"busan", "부산회생법원", "busan", "부산광역시"},
		{"gwangju", "광주회생법원", "gwangju", "전남광주통합특별시"},
	} {
		add("court-rehabilitation-"+court.id, court.name, CourtLevelDistrict, CourtTypeRehabilitation, CourtUnitTypeMain, "court-high-"+court.parent, court.seat)
	}

	// ── 4단계: 시·군법원 ──

	// addMunicipal은 같은 상위 법원에 속한 시·군법원들을 한꺼번에 추가합니다.
	//   - districtID: 시·군법원이 속한 지방법원의 짧은 식별자(이름과 ID의 앞부분에 씁니다)
	//   - parentID: 실제 상위 법원 ID(지방법원 본원 또는 지원)
	//   - areas: 별표 2 "시·군법원" 열의 지역 이름들(예: "포천", "연천")
	//
	// 각 지역에 대해 다음 값을 만듭니다.
	//   - 이름: "<지방법원 이름> <지역><시|군>법원"(예: "의정부지방법원 연천군법원").
	//     시와 군의 구분은 courts_municipal.go의 municipalCourtCounties로 정합니다.
	//   - 소재지: 별표 2 "소재지" 열(courts_municipal.go의 municipalCourtSeats, 예: 연천군법원은 "연천읍")
	//   - 관할 지역: 이름의 시·군 하나. 별표 7과 다른 예외(오산, 창원남부)는 아래에서 덮어씁니다.
	//   - ID: "<지방법원 ID>-municipal-<3자리 일련번호>". 일련번호는 지방법원별로 아래 호출 순서대로 매깁니다.
	municipalSequence := make(map[string]int)
	addMunicipal := func(districtID, parentID string, areas ...string) {
		for _, area := range areas {
			municipalSequence[districtID]++

			// 지역이 군이면 "군", 그렇지 않으면 "시"를 붙입니다.
			kind := "시"
			if municipalCourtCounties[area] {
				kind = "군"
			}

			// 대부분의 시·군법원은 이름에 나온 시·군 하나를 관할합니다.
			jurisdictions := []string{area + kind}

			// 별표 7의 예외: 오산시법원은 화성시까지 맡고, 창원남부시법원은 창원시 진해구만 맡습니다.
			if area == "오산" {
				jurisdictions = []string{"오산시", "화성시"}
			} else if area == "창원남부" {
				jurisdictions = []string{"창원시 진해구"}
			}

			// 소재지는 별표 2의 "소재지" 열을 씁니다. 같은 지역 이름(예: 고성)이 여러 지방법원에 있어
			// "<지방법원 식별자>/<지역>"을 키로 찾습니다. 표에 없으면 데이터 누락이므로 패닉을 일으킵니다.
			seat, exists := municipalCourtSeats[districtID+"/"+area]
			if !exists {
				panic(fmt.Sprintf("시·군법원 소재지를 찾을 수 없습니다: %s/%s", districtID, area))
			}

			// %03d는 일련번호를 세 자리로 맞춥니다(예: 1 → "001").
			add(fmt.Sprintf("%s-municipal-%03d", districtIDs[districtID], municipalSequence[districtID]),
				districtNames[districtID]+" "+area+kind+"법원",
				CourtLevelLocal, CourtTypeGeneral, CourtUnitTypeMunicipalCounty,
				parentID, seat, jurisdictions...)
		}
	}

	// 아래 호출은 별표 2의 "지방법원 / 지원 / 시·군법원" 열을 그대로 옮긴 것입니다.
	// 지원 열이 비어 있으면 지방법원 본원(districtIDs), 지원 열이 있으면 그 지원(branchIDs)을 상위 법원으로 넘깁니다.
	// 예: 김포시법원은 별표 2에서 인천지방법원 부천지원 소속이므로 branchIDs["incheon-bucheon"]을 넘깁니다.
	addMunicipal("uijeongbu", districtIDs["uijeongbu"], "포천", "연천", "철원", "동두천")
	addMunicipal("uijeongbu", branchIDs["uijeongbu-goyang"], "파주")
	addMunicipal("uijeongbu", branchIDs["uijeongbu-namyangju"], "가평")
	addMunicipal("incheon", districtIDs["incheon"], "강화")
	addMunicipal("incheon", branchIDs["incheon-bucheon"], "김포")
	addMunicipal("suwon", districtIDs["suwon"], "용인", "오산")
	addMunicipal("suwon", branchIDs["suwon-seongnam"], "광주")
	addMunicipal("suwon", branchIDs["suwon-yeoju"], "양평", "이천")
	addMunicipal("suwon", branchIDs["suwon-pyeongtaek"], "안성")
	addMunicipal("suwon", branchIDs["suwon-ansan"], "광명")
	addMunicipal("chuncheon", districtIDs["chuncheon"], "인제", "홍천", "양구", "화천")
	addMunicipal("chuncheon", branchIDs["chuncheon-gangneung"], "삼척", "동해")
	addMunicipal("chuncheon", branchIDs["chuncheon-wonju"], "횡성")
	addMunicipal("chuncheon", branchIDs["chuncheon-sokcho"], "고성", "양양")
	addMunicipal("chuncheon", branchIDs["chuncheon-yeongwol"], "정선", "태백", "평창")
	addMunicipal("daejeon", districtIDs["daejeon"], "세종특별자치", "금산")
	addMunicipal("daejeon", branchIDs["daejeon-hongseong"], "서천", "보령", "예산")
	addMunicipal("daejeon", branchIDs["daejeon-gongju"], "청양")
	addMunicipal("daejeon", branchIDs["daejeon-nonsan"], "부여")
	addMunicipal("daejeon", branchIDs["daejeon-seosan"], "태안", "당진")
	addMunicipal("daejeon", branchIDs["daejeon-cheonan"], "아산")
	addMunicipal("cheongju", districtIDs["cheongju"], "보은", "괴산", "진천")
	addMunicipal("cheongju", branchIDs["cheongju-chungju"], "음성")
	addMunicipal("cheongju", branchIDs["cheongju-jecheon"], "단양")
	addMunicipal("cheongju", branchIDs["cheongju-yeongdong"], "옥천")
	addMunicipal("daegu", districtIDs["daegu"], "청도", "영천", "칠곡", "경산")
	addMunicipal("daegu", branchIDs["daegu-seobu"], "성주", "고령")
	addMunicipal("daegu", branchIDs["daegu-andong"], "영주", "봉화")
	addMunicipal("daegu", branchIDs["daegu-gimcheon"], "구미")
	addMunicipal("daegu", branchIDs["daegu-sangju"], "예천", "문경")
	addMunicipal("daegu", branchIDs["daegu-uiseong"], "청송", "군위")
	addMunicipal("daegu", branchIDs["daegu-yeongdeok"], "울진", "영양")
	addMunicipal("ulsan", districtIDs["ulsan"], "양산")
	addMunicipal("changwon", districtIDs["changwon"], "창원남부", "김해")
	addMunicipal("changwon", branchIDs["changwon-masan"], "함안", "의령")
	addMunicipal("changwon", branchIDs["changwon-jinju"], "하동", "사천", "남해", "산청")
	addMunicipal("changwon", branchIDs["changwon-tongyeong"], "거제", "고성")
	addMunicipal("changwon", branchIDs["changwon-miryang"], "창녕")
	addMunicipal("changwon", branchIDs["changwon-geochang"], "합천", "함양")
	addMunicipal("gwangju", districtIDs["gwangju"], "곡성", "영광", "나주", "장성", "화순", "담양")
	addMunicipal("gwangju", branchIDs["gwangju-mokpo"], "함평", "영암", "무안")
	addMunicipal("gwangju", branchIDs["gwangju-jangheung"], "강진")
	addMunicipal("gwangju", branchIDs["gwangju-suncheon"], "보성", "고흥", "여수", "구례", "광양")
	addMunicipal("gwangju", branchIDs["gwangju-haenam"], "완도", "진도")
	addMunicipal("jeonju", districtIDs["jeonju"], "진안", "김제", "무주", "임실")
	addMunicipal("jeonju", branchIDs["jeonju-gunsan"], "익산")
	addMunicipal("jeonju", branchIDs["jeonju-jeongeup"], "부안", "고창")
	addMunicipal("jeonju", branchIDs["jeonju-namwon"], "장수", "순창")
	addMunicipal("jeju", districtIDs["jeju"], "서귀포")

	// ── 5단계: 재판부 ──

	// courts_panels.go의 표에서 법원 ID별 재판부 이름을 읽어 Panels에 붙입니다.
	// 표의 법원 ID가 목록에 없으면 오타이거나 법원을 등록하지 않은 것이므로 패닉을 일으킵니다.
	// 표는 map이라 법원을 도는 순서는 매번 다르지만, 한 법원 안의 재판부는 항상 표에 적힌 순서대로 붙으므로
	// 재판부 ID의 일련번호는 실행할 때마다 같습니다.
	courtIndexes := make(map[string]int, len(courts))
	for i := range courts {
		courtIndexes[courts[i].ID] = i
	}
	for courtID, panelNames := range courtPanelNamesByCourt {
		courtIndex, exists := courtIndexes[courtID]
		if !exists {
			panic(fmt.Sprintf("재판부의 소속 법원을 찾을 수 없습니다: %s", courtID))
		}
		court := &courts[courtIndex]
		// 표의 값은 줄마다 재판부 이름 하나를 적은 여러 줄 문자열입니다. 앞뒤 공백과 빈 줄은 버립니다.
		for _, line := range strings.Split(panelNames, "\n") {
			name := strings.TrimSpace(line)
			if name == "" {
				continue
			}
			// 원외재판부는 별도 법원으로 등록했으므로, 표의 재판부는 모두 소속 법원 소재지의 재판부입니다.
			court.Panels = append(court.Panels, CourtPanel{
				ID:   fmt.Sprintf("%s-panel-%03d", court.ID, len(court.Panels)+1),
				Name: name,
				Seat: court.Seat,
			})
		}
	}

	// ── 6단계: 우편번호와 주소 ──

	// courts_addresses.go의 표에서 법원 이름으로 주소를 찾아 채웁니다.
	// 대법원 홈페이지에 아직 없는 법원(courtsWithoutAddress)만 주소를 비워 둘 수 있습니다.
	institutions := make([]*Institution, len(courts))
	for i := range courts {
		institutions[i] = &courts[i].Institution
	}
	applyPostalAddresses("법원", institutions, courtAddresses, courtsWithoutAddress)

	return courts
}
