package casepulse

// 이 파일은 전국 공소청 목록을 만드는 로직과 기관 구조 표를 담습니다.
//
// 원본 자료는 다음과 같습니다.
//   - 명칭·위치: documents/pdfs/prosecution의 「공소청의 위치와 각급 공소청 및 지청의 명칭 및 위치에 관한 규정」 별표
//     (2026. 9. 29. 개정, "공소청 외의 각급 공소청 및 지청의 명칭과 위치")
//   - 소속 관계: 공소청 홈페이지 조직구조(https://www.spo.go.kr/site/spo/06/10602010100002018100812.jsp)
//   - 우편번호·주소: prosecutions_addresses.go(공소청 홈페이지 전국공소청찾기)

// prosecutionDirectory는 패키지가 처음 로드될 때 한 번 만든 공소청 원본 목록입니다.
// 외부에는 ProsecutionOffices가 복사본을 돌려줍니다.
var prosecutionDirectory = buildProsecutionDirectory()

// ProsecutionOffices는 전국 공소청 목록(공소청 1, 광역공소청 6, 지방공소청 18, 지청 42)을 반환합니다.
//
// 순서는 공소청 → 광역공소청 → 지방공소청 → 지청이며, 각 단계 안에서는 별표의 순서를 따릅니다.
// ProsecutionOffice에는 슬라이스 필드가 없으므로 값 복사만으로 내부 원본과 완전히 분리됩니다.
func ProsecutionOffices() []ProsecutionOffice {
	offices := make([]ProsecutionOffice, len(prosecutionDirectory))
	copy(offices, prosecutionDirectory)
	return offices
}

// buildProsecutionDirectory는 표를 합쳐 전국 공소청 목록을 만듭니다.
// 패키지 초기화 때 prosecutionDirectory를 만들기 위해 한 번만 호출됩니다.
//
// 기관 ID는 "prosecution-<단계>-<짧은 식별자>" 형식입니다(예: "prosecution-regional-seoul",
// "prosecution-district-incheon", "prosecution-branch-incheon-bucheon"). 짧은 식별자는 같은 지역의
// 법원 ID와 맞춰 두어 법원과 공소청을 쉽게 짝지을 수 있게 했습니다.
// 각 기관은 ParentOfficeID로 바로 위 기관을 가리킵니다.
func buildProsecutionDirectory() []ProsecutionOffice {
	var offices []ProsecutionOffice
	// add는 기관 한 곳을 offices 끝에 추가하는 내부 도우미입니다. 주소는 마지막 단계에서 채웁니다.
	add := func(id, name string, level ProsecutionLevel, unit ProsecutionUnitType, parentID, seat string) {
		offices = append(offices, ProsecutionOffice{
			Institution:    Institution{ID: id, Name: name, Type: InstitutionTypeProsecution},
			Level:          level,
			UnitType:       unit,
			ParentOfficeID: parentID,
			Seat:           seat,
		})
	}

	// 공소청(본청)은 최상위 기관이라 상위 기관이 없습니다. 별표는 "공소청 외의" 기관만 다루므로
	// 공소청 자체의 소재지는 홈페이지 주소(서울특별시 서초구 반포대로 157)의 시·도를 씁니다.
	add("prosecution-national", "공소청", ProsecutionLevelNational, ProsecutionUnitTypeMain, "", "서울특별시")

	// 광역공소청. 각 행은 {짧은 식별자, 공식 명칭, 위치}이며 모두 공소청 소속입니다.
	// 광역공소청은 고등법원에 대응하며, 광주광역공소청의 위치는 별표대로 "전남광주통합특별시"입니다.
	for _, office := range []struct{ id, name, seat string }{
		{"seoul", "서울광역공소청", "서울특별시"},
		{"daejeon", "대전광역공소청", "대전광역시"},
		{"daegu", "대구광역공소청", "대구광역시"},
		{"busan", "부산광역공소청", "부산광역시"},
		{"gwangju", "광주광역공소청", "전남광주통합특별시"},
		{"suwon", "수원광역공소청", "수원시"},
	} {
		add("prosecution-regional-"+office.id, office.name, ProsecutionLevelRegional, ProsecutionUnitTypeMain, "prosecution-national", office.seat)
	}

	// 지방공소청. 각 행은 {짧은 식별자, 공식 명칭, 소속 광역공소청의 짧은 식별자, 위치}입니다.
	// 별표에는 어느 광역공소청에 속하는지 나와 있지 않아 공소청 홈페이지 조직구조에 따라 연결했습니다.
	// 이 관계는 지방법원이 고등법원에 소속되는 관계(courts.go)와 같습니다.
	for _, office := range []struct{ id, name, regional, seat string }{
		{"seoul-central", "서울중앙지방공소청", "seoul", "서울특별시"},
		{"seoul-east", "서울동부지방공소청", "seoul", "서울특별시"},
		{"seoul-south", "서울남부지방공소청", "seoul", "서울특별시"},
		{"seoul-north", "서울북부지방공소청", "seoul", "서울특별시"},
		{"seoul-west", "서울서부지방공소청", "seoul", "서울특별시"},
		{"uijeongbu", "의정부지방공소청", "seoul", "의정부시"},
		{"incheon", "인천지방공소청", "seoul", "인천광역시"},
		{"suwon", "수원지방공소청", "suwon", "수원시"},
		{"chuncheon", "춘천지방공소청", "seoul", "춘천시"},
		{"daejeon", "대전지방공소청", "daejeon", "대전광역시"},
		{"cheongju", "청주지방공소청", "daejeon", "청주시"},
		{"daegu", "대구지방공소청", "daegu", "대구광역시"},
		{"busan", "부산지방공소청", "busan", "부산광역시"},
		{"ulsan", "울산지방공소청", "busan", "울산광역시"},
		{"changwon", "창원지방공소청", "busan", "창원시"},
		{"gwangju", "광주지방공소청", "gwangju", "전남광주통합특별시"},
		{"jeonju", "전주지방공소청", "gwangju", "전주시"},
		{"jeju", "제주지방공소청", "gwangju", "제주시"},
	} {
		add("prosecution-district-"+office.id, office.name, ProsecutionLevelDistrict, ProsecutionUnitTypeMain, "prosecution-regional-"+office.regional, office.seat)
	}

	// 지청. 각 행은 {짧은 식별자, 공식 명칭, 소속 지방공소청의 짧은 식별자, 위치}입니다.
	// 지청 이름은 항상 "<소속 지방공소청> <지청 이름>" 형식이며, 위치는 별표대로 시·읍 단위입니다(예: "영월읍").
	for _, office := range []struct{ id, name, district, seat string }{
		{"uijeongbu-goyang", "의정부지방공소청 고양지청", "uijeongbu", "고양시"},
		{"uijeongbu-namyangju", "의정부지방공소청 남양주지청", "uijeongbu", "남양주시"},
		{"incheon-bucheon", "인천지방공소청 부천지청", "incheon", "부천시"},
		{"suwon-seongnam", "수원지방공소청 성남지청", "suwon", "성남시"},
		{"suwon-yeoju", "수원지방공소청 여주지청", "suwon", "여주시"},
		{"suwon-pyeongtaek", "수원지방공소청 평택지청", "suwon", "평택시"},
		{"suwon-ansan", "수원지방공소청 안산지청", "suwon", "안산시"},
		{"suwon-anyang", "수원지방공소청 안양지청", "suwon", "안양시"},
		{"chuncheon-gangneung", "춘천지방공소청 강릉지청", "chuncheon", "강릉시"},
		{"chuncheon-wonju", "춘천지방공소청 원주지청", "chuncheon", "원주시"},
		{"chuncheon-sokcho", "춘천지방공소청 속초지청", "chuncheon", "속초시"},
		{"chuncheon-yeongwol", "춘천지방공소청 영월지청", "chuncheon", "영월읍"},
		{"daejeon-hongseong", "대전지방공소청 홍성지청", "daejeon", "홍성읍"},
		{"daejeon-gongju", "대전지방공소청 공주지청", "daejeon", "공주시"},
		{"daejeon-nonsan", "대전지방공소청 논산지청", "daejeon", "논산시"},
		{"daejeon-seosan", "대전지방공소청 서산지청", "daejeon", "서산시"},
		{"daejeon-cheonan", "대전지방공소청 천안지청", "daejeon", "천안시"},
		{"cheongju-chungju", "청주지방공소청 충주지청", "cheongju", "충주시"},
		{"cheongju-jecheon", "청주지방공소청 제천지청", "cheongju", "제천시"},
		{"cheongju-yeongdong", "청주지방공소청 영동지청", "cheongju", "영동읍"},
		{"daegu-seobu", "대구지방공소청 서부지청", "daegu", "대구광역시"},
		{"daegu-andong", "대구지방공소청 안동지청", "daegu", "안동시"},
		{"daegu-gyeongju", "대구지방공소청 경주지청", "daegu", "경주시"},
		{"daegu-pohang", "대구지방공소청 포항지청", "daegu", "포항시"},
		{"daegu-gimcheon", "대구지방공소청 김천지청", "daegu", "김천시"},
		{"daegu-sangju", "대구지방공소청 상주지청", "daegu", "상주시"},
		{"daegu-uiseong", "대구지방공소청 의성지청", "daegu", "의성읍"},
		{"daegu-yeongdeok", "대구지방공소청 영덕지청", "daegu", "영덕읍"},
		{"busan-dongbu", "부산지방공소청 동부지청", "busan", "부산광역시"},
		{"busan-seobu", "부산지방공소청 서부지청", "busan", "부산광역시"},
		{"changwon-masan", "창원지방공소청 마산지청", "changwon", "창원시"},
		{"changwon-jinju", "창원지방공소청 진주지청", "changwon", "진주시"},
		{"changwon-tongyeong", "창원지방공소청 통영지청", "changwon", "통영시"},
		{"changwon-miryang", "창원지방공소청 밀양지청", "changwon", "밀양시"},
		{"changwon-geochang", "창원지방공소청 거창지청", "changwon", "거창읍"},
		{"gwangju-mokpo", "광주지방공소청 목포지청", "gwangju", "목포시"},
		{"gwangju-jangheung", "광주지방공소청 장흥지청", "gwangju", "장흥읍"},
		{"gwangju-suncheon", "광주지방공소청 순천지청", "gwangju", "순천시"},
		{"gwangju-haenam", "광주지방공소청 해남지청", "gwangju", "해남읍"},
		{"jeonju-gunsan", "전주지방공소청 군산지청", "jeonju", "군산시"},
		{"jeonju-jeongeup", "전주지방공소청 정읍지청", "jeonju", "정읍시"},
		{"jeonju-namwon", "전주지방공소청 남원지청", "jeonju", "남원시"},
	} {
		add("prosecution-branch-"+office.id, office.name, ProsecutionLevelDistrict, ProsecutionUnitTypeBranch, "prosecution-district-"+office.district, office.seat)
	}

	// ── 우편번호와 주소 ──
	// prosecutions_addresses.go의 표에서 기관 이름으로 주소를 찾아 채웁니다.
	// 모든 기관의 주소가 있어야 하므로 예외 목록은 nil입니다.
	institutions := make([]*Institution, len(offices))
	for i := range offices {
		institutions[i] = &offices[i].Institution
	}
	applyPostalAddresses("공소청", institutions, prosecutionAddresses, nil)

	return offices
}
