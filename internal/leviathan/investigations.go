package leviathan

// 이 파일은 중대범죄수사청 본청과 지방수사청 목록을 만드는 로직과 데이터를 담습니다.
//
// 원본 자료는 다음과 같습니다.
//   - 명칭·위치·관할구역·약칭: documents/pdfs/prosecution의 「중대범죄수사청과 그 소속기관 직제」 별표 1
//   - 주소: 중대범죄수사청 홈페이지 전국중수청찾기(https://www.sci.go.kr/content.do?key=101041)
//
// 기관 수가 적어 주소도 이 파일의 표에 함께 적습니다. 홈페이지가 지번 주소로 안내하므로 주소는 지번 주소이며,
// 시·도 이름만 법원·공소청 주소와 같은 정식 명칭으로 맞췄습니다. 홈페이지가 우편번호를 안내하지 않아
// PostalCode는 비어 있고, 전화번호는 넣지 않습니다.

// investigationDirectory는 패키지가 처음 로드될 때 한 번 만든 중대범죄수사청 원본 목록입니다.
// 외부에는 InvestigationAgencies가 복사본을 돌려줍니다.
var investigationDirectory = buildInvestigationDirectory()

// InvestigationAgencies는 중대범죄수사청 본청 1곳과 지방수사청 6곳의 목록을 반환합니다.
//
// 순서는 본청 → 서울·수원·대전·대구·부산·광주 지방수사청(별표 1의 순서)입니다.
// 반환값은 Jurisdictions까지 복사한 깊은 복사본이므로 호출한 쪽에서 고쳐도 원본은 바뀌지 않습니다.
func InvestigationAgencies() []InvestigationAgency {
	agencies := make([]InvestigationAgency, len(investigationDirectory))
	copy(agencies, investigationDirectory)
	for i := range agencies {
		agencies[i].Jurisdictions = append([]string(nil), investigationDirectory[i].Jurisdictions...)
	}
	return agencies
}

// buildInvestigationDirectory는 중대범죄수사청 목록을 만듭니다.
// 패키지 초기화 때 investigationDirectory를 만들기 위해 한 번만 호출됩니다.
//
// 기관 ID는 본청이 "investigation-national", 지방수사청이 "investigation-district-<짧은 식별자>"입니다.
// 직제의 이름("중대범죄수사청과 그 소속기관")대로 지방수사청은 모두 본청의 소속기관이므로
// ParentAgencyID로 본청을 가리킵니다. 홈페이지와 직제 어디에도 지방수사청 사이의 상하 관계는 없습니다.
func buildInvestigationDirectory() []InvestigationAgency {
	const nationalID = "investigation-national"

	agencies := []InvestigationAgency{{
		Institution: Institution{
			ID:   nationalID,
			Name: "중대범죄수사청",
			Type: InstitutionTypeInvestigation,
			// 홈페이지 기관 목록의 지번 주소입니다. 지방수사청 주소가 모두 지번이라 형식을 맞췄습니다.
			// 같은 건물의 도로명 주소는 홈페이지 하단에 "서울특별시 중구 충무로9길 38-5"로 적혀 있습니다.
			Address: "서울특별시 중구 수표동 35-10",
		},
		Level:         InvestigationLevelNational,
		Seat:          "서울특별시",
		Jurisdictions: []string{"전국"},
	}}

	// 지방수사청(직제 별표 1). 각 행은 {짧은 식별자, 공식 명칭, 약칭, 위치, 주소, 관할 지역}입니다.
	//   - 약칭: 별표 1 비고의 현판·표지판용 이름(정식 명칭에서 "중대범죄"를 뺀 이름)
	//   - 위치: 별표 1의 "위치" 열. 수원지방수사청은 별표에 시가 아닌 "경기도"로만 적혀 있습니다.
	//   - 관할 지역: Court.Jurisdictions와 같은 규칙으로, 도 이름은 그 도의 첫 시·군에만 붙입니다.
	for _, agency := range []struct {
		id, name, shortName, seat, address string
		jurisdictions                      []string
	}{
		{"seoul", "서울지방중대범죄수사청", "서울지방수사청", "서울특별시",
			"서울특별시 중구 수표동 35-10",
			[]string{"서울특별시", "경기도 의정부시", "포천시", "동두천시", "양주시", "연천군", "고양시",
				"파주시", "남양주시", "구리시", "가평군", "부천시", "김포시", "인천광역시", "강원특별자치도"}},
		// 홈페이지 안내: 2026. 9. ~ 2027. 4.은 임시청사이며, 2027. 5. 이후 성남시 분당구 구미동 192-3으로
		// 옮길 예정입니다. 이전하면 주소를 고쳐야 합니다.
		{"suwon", "수원지방중대범죄수사청", "수원지방수사청", "경기도",
			"경기도 수원시 영통구 도청로 52",
			[]string{"경기도 수원시", "오산시", "화성시", "용인시", "성남시", "하남시", "광주시", "이천시",
				"여주시", "양평군", "평택시", "안성시", "안산시", "광명시", "시흥시", "안양시", "군포시", "의왕시", "과천시"}},
		{"daejeon", "대전지방중대범죄수사청", "대전지방수사청", "세종특별자치시",
			"세종특별자치시 소담동 577",
			[]string{"대전광역시", "세종특별자치시", "충청남도", "충청북도"}},
		{"daegu", "대구지방중대범죄수사청", "대구지방수사청", "대구광역시",
			"대구광역시 남구 대명동 1795-3",
			[]string{"대구광역시", "경상북도"}},
		{"busan", "부산지방중대범죄수사청", "부산지방수사청", "부산광역시",
			"부산광역시 강서구 명지동 3595-2",
			[]string{"부산광역시", "울산광역시", "경상남도"}},
		// 홈페이지 주소는 "광주광역시"로 적혀 있으나, 법원·공소청과 같이 "전남광주통합특별시"로 정규화합니다.
		{"gwangju", "광주지방중대범죄수사청", "광주지방수사청", "전남광주통합특별시",
			"전남광주통합특별시 북구 중흥동 712-3",
			[]string{"전남광주통합특별시", "전북특별자치도", "제주특별자치도"}},
	} {
		agencies = append(agencies, InvestigationAgency{
			Institution: Institution{
				ID:      "investigation-district-" + agency.id,
				Name:    agency.name,
				Type:    InstitutionTypeInvestigation,
				Address: agency.address,
			},
			ShortName:      agency.shortName,
			Level:          InvestigationLevelDistrict,
			ParentAgencyID: nationalID,
			Seat:           agency.seat,
			Jurisdictions:  agency.jurisdictions,
		})
	}

	return agencies
}
