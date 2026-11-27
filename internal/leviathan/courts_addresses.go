package leviathan

// 이 파일은 법원별 우편번호·주소 표만 담습니다.
//
// buildCourtDirectory가 법원을 모두 만든 뒤 addresses.go의 공통 함수로 이 표의 값을 채웁니다.

// courtsWithoutAddress는 법령 별표에는 있지만 대법원 홈페이지 각급법원 안내에 아직 없어 주소를 비워 두는 법원입니다.
// 전주가정법원과 그 지원은 별표 5(2026. 3. 17. 개정)에 신설되었으나 홈페이지에 주소가 없습니다.
// 개원해서 주소가 확인되면 이 표에서 빼고 courtAddresses에 주소를 추가합니다.
var courtsWithoutAddress = map[string]bool{
	"전주가정법원":      true,
	"전주가정법원 군산지원": true,
	"전주가정법원 정읍지원": true,
	"전주가정법원 남원지원": true,
}

// courtAddresses는 법원 이름(Court.Name)별 우편번호와 주소입니다.
//
// 출처는 대법원 홈페이지 각급법원 안내(https://www.scourt.go.kr/region/location/RegionLocationListAction.work)이며,
// 홈페이지 표기를 다음 규칙으로 도로명 주소 형식에 맞춰 정규화했습니다. 공소청 주소도 같은 규칙을 따릅니다.
//   - 시·도는 정식 명칭으로 씁니다(예: "충남" → "충청남도", "서울시" → "서울특별시",
//     "전남광주"·"전라남도"·"광주광역시" → "전남광주통합특별시").
//   - 도로명과 건물번호는 띄어 쓰고(예: "마들로749" → "마들로 749"),
//     "번길"·"길"이 붙은 도로명은 붙여 씁니다(예: "둔산중로 78번길 45" → "둔산중로78번길 45").
//   - 괄호 안의 법정동·지번과 주소 뒤에 붙은 건물 이름은 뺍니다(예: "법원로31(거제동 1500번지)" → "법원로 31").
//   - 지번 주소 뒤 괄호에 도로명 주소가 있으면 도로명 주소를 씁니다.
//   - 군 지역 주소에 읍·면이 빠져 있으면 괄호 안의 읍·면을 채워 넣습니다(예: "양양군 양양로 93" → "양양군 양양읍 양양로 93").
//   - 같은 청사 안의 층처럼 꼭 필요한 상세주소만 쉼표 뒤에 남깁니다(예: "법대로 55, 10층").
//
// 홈페이지 이름이 이 패키지의 이름과 다르면 줄 끝 주석에 홈페이지 이름을 적었습니다.
// 홈페이지는 원외재판부를 "서울고등법원 원외재판부(인천)"으로, 시·군법원을 소속 지원까지 붙여
// "의정부지방법원 고양지원 파주시법원"으로 적습니다.
var courtAddresses = map[string]postalAddress{
	"대법원":              postalAddress{postalCode: "06590", address: "서울특별시 서초구 서초대로 219"},
	"서울고등법원":           postalAddress{postalCode: "06594", address: "서울특별시 서초구 서초중앙로 157"},
	"대전고등법원":           postalAddress{postalCode: "35237", address: "대전광역시 서구 둔산중로78번길 45"},
	"대구고등법원":           postalAddress{postalCode: "42027", address: "대구광역시 수성구 동대구로 364"},
	"부산고등법원":           postalAddress{postalCode: "47510", address: "부산광역시 연제구 법원로 31"},
	"광주고등법원":           postalAddress{postalCode: "61441", address: "전남광주통합특별시 동구 준법로 7-12"},
	"광주고등법원(전주)":       postalAddress{postalCode: "54867", address: "전북특별자치도 전주시 덕진구 가인로 33"}, // 홈페이지 이름: 광주고등법원 원외재판부(전주)
	"광주고등법원(제주)":       postalAddress{postalCode: "63223", address: "제주특별자치도 제주시 남광북5길 3"},    // 홈페이지 이름: 광주고등법원 원외재판부(제주)
	"수원고등법원":           postalAddress{postalCode: "16512", address: "경기도 수원시 영통구 법조로 105"},
	"서울고등법원(인천)":       postalAddress{postalCode: "22220", address: "인천광역시 미추홀구 소성로163번길 17"},  // 홈페이지 이름: 서울고등법원 원외재판부(인천)
	"서울고등법원(춘천)":       postalAddress{postalCode: "24342", address: "강원특별자치도 춘천시 공지로 284"},     // 홈페이지 이름: 서울고등법원 원외재판부(춘천)
	"대전고등법원(청주)":       postalAddress{postalCode: "28624", address: "충청북도 청주시 서원구 산남로62번길 51"}, // 홈페이지 이름: 대전고등법원 원외재판부(청주)
	"부산고등법원(울산)":       postalAddress{postalCode: "44643", address: "울산광역시 남구 법대로 55"},         // 홈페이지 이름: 부산고등법원 원외재판부(울산)
	"부산고등법원(창원)":       postalAddress{postalCode: "51456", address: "경상남도 창원시 성산구 창이대로 681"},   // 홈페이지 이름: 부산고등법원 원외재판부(창원)
	"특허법원":             postalAddress{postalCode: "35239", address: "대전광역시 서구 둔산중로 69"},        // 홈페이지 이름: 특허법원 [대전]
	"서울중앙지방법원":         postalAddress{postalCode: "06594", address: "서울특별시 서초구 서초중앙로 157"},
	"서울동부지방법원":         postalAddress{postalCode: "05856", address: "서울특별시 송파구 법원로 101"},
	"서울남부지방법원":         postalAddress{postalCode: "08088", address: "서울특별시 양천구 신월로 386"},
	"서울북부지방법원":         postalAddress{postalCode: "01322", address: "서울특별시 도봉구 마들로 749"},
	"서울서부지방법원":         postalAddress{postalCode: "04207", address: "서울특별시 마포구 마포대로 174"},
	"의정부지방법원":          postalAddress{postalCode: "11616", address: "경기도 의정부시 녹양로34번길 23"},
	"인천지방법원":           postalAddress{postalCode: "22220", address: "인천광역시 미추홀구 소성로163번길 17"},
	"수원지방법원":           postalAddress{postalCode: "16512", address: "경기도 수원시 영통구 법조로 105"},
	"춘천지방법원":           postalAddress{postalCode: "24342", address: "강원특별자치도 춘천시 공지로 284"},
	"대전지방법원":           postalAddress{postalCode: "35237", address: "대전광역시 서구 둔산중로78번길 45"},
	"청주지방법원":           postalAddress{postalCode: "28624", address: "충청북도 청주시 서원구 산남로62번길 51"},
	"대구지방법원":           postalAddress{postalCode: "42027", address: "대구광역시 수성구 동대구로 364"},
	"부산지방법원":           postalAddress{postalCode: "47510", address: "부산광역시 연제구 법원로 31"},
	"울산지방법원":           postalAddress{postalCode: "44643", address: "울산광역시 남구 법대로 55"},
	"창원지방법원":           postalAddress{postalCode: "51456", address: "경상남도 창원시 성산구 창이대로 681"},
	"광주지방법원":           postalAddress{postalCode: "61441", address: "전남광주통합특별시 동구 준법로 7-12"},
	"전주지방법원":           postalAddress{postalCode: "54867", address: "전북특별자치도 전주시 덕진구 가인로 33"},
	"제주지방법원":           postalAddress{postalCode: "63223", address: "제주특별자치도 제주시 남광북5길 3"},
	"의정부지방법원 고양지원":     postalAddress{postalCode: "10413", address: "경기도 고양시 일산동구 장백로 209"},
	"의정부지방법원 남양주지원":    postalAddress{postalCode: "12249", address: "경기도 남양주시 다산중앙로82번안길 161"},
	"인천지방법원 부천지원":      postalAddress{postalCode: "14602", address: "경기도 부천시 원미구 상일로 129"},
	"수원지방법원 성남지원":      postalAddress{postalCode: "13143", address: "경기도 성남시 수정구 산성대로 451"},
	"수원지방법원 여주지원":      postalAddress{postalCode: "12638", address: "경기도 여주시 현암로 21-12"},
	"수원지방법원 평택지원":      postalAddress{postalCode: "17848", address: "경기도 평택시 평남로 1036"},
	"수원지방법원 안산지원":      postalAddress{postalCode: "15472", address: "경기도 안산시 단원구 광덕서로 75"},
	"수원지방법원 안양지원":      postalAddress{postalCode: "14054", address: "경기도 안양시 동안구 관평로212번길 70"},
	"춘천지방법원 강릉지원":      postalAddress{postalCode: "25463", address: "강원특별자치도 강릉시 동해대로 3288-18"},
	"춘천지방법원 원주지원":      postalAddress{postalCode: "26478", address: "강원특별자치도 원주시 시청로 149"},
	"춘천지방법원 속초지원":      postalAddress{postalCode: "24822", address: "강원특별자치도 속초시 법대로 15"},
	"춘천지방법원 영월지원":      postalAddress{postalCode: "26228", address: "강원특별자치도 영월군 영월읍 영월향교1길 53"},
	"대전지방법원 홍성지원":      postalAddress{postalCode: "32226", address: "충청남도 홍성군 홍성읍 법원로 38"},
	"대전지방법원 공주지원":      postalAddress{postalCode: "32568", address: "충청남도 공주시 한적2길 34-15"},
	"대전지방법원 논산지원":      postalAddress{postalCode: "32930", address: "충청남도 논산시 강경읍 계백로 99"},
	"대전지방법원 서산지원":      postalAddress{postalCode: "31988", address: "충청남도 서산시 공림4로 24"},
	"대전지방법원 천안지원":      postalAddress{postalCode: "31198", address: "충청남도 천안시 동남구 청수14로 77"},
	"청주지방법원 충주지원":      postalAddress{postalCode: "27357", address: "충청북도 충주시 계명대로 103"},
	"청주지방법원 제천지원":      postalAddress{postalCode: "27165", address: "충청북도 제천시 칠성로 53"},
	"청주지방법원 영동지원":      postalAddress{postalCode: "29150", address: "충청북도 영동군 영동읍 영동황간로 99"},
	"대구지방법원 서부지원":      postalAddress{postalCode: "42635", address: "대구광역시 달서구 장산남로 30"},
	"대구지방법원 안동지원":      postalAddress{postalCode: "36752", address: "경상북도 안동시 강남로 304"},
	"대구지방법원 경주지원":      postalAddress{postalCode: "38145", address: "경상북도 경주시 화랑로 89"},
	"대구지방법원 포항지원":      postalAddress{postalCode: "37573", address: "경상북도 포항시 북구 법원로 181"},
	"대구지방법원 김천지원":      postalAddress{postalCode: "39518", address: "경상북도 김천시 물망골길 39"},
	"대구지방법원 상주지원":      postalAddress{postalCode: "37160", address: "경상북도 상주시 북천로 17-9"},
	"대구지방법원 의성지원":      postalAddress{postalCode: "37339", address: "경상북도 의성군 의성읍 군청길 67"},
	"대구지방법원 영덕지원":      postalAddress{postalCode: "36427", address: "경상북도 영덕군 영덕읍 경동로 8337"},
	"부산지방법원 동부지원":      postalAddress{postalCode: "48053", address: "부산광역시 해운대구 재반로112번길 20"},
	"부산지방법원 서부지원":      postalAddress{postalCode: "46726", address: "부산광역시 강서구 명지국제7로 77"},
	"창원지방법원 마산지원":      postalAddress{postalCode: "51265", address: "경상남도 창원시 마산합포구 완월동7길 16"},
	"창원지방법원 진주지원":      postalAddress{postalCode: "52676", address: "경상남도 진주시 진양호로 303"},
	"창원지방법원 통영지원":      postalAddress{postalCode: "53029", address: "경상남도 통영시 용남면 동달안길 67"},
	"창원지방법원 밀양지원":      postalAddress{postalCode: "50424", address: "경상남도 밀양시 법원로 20"},
	"창원지방법원 거창지원":      postalAddress{postalCode: "50129", address: "경상남도 거창군 거창읍 죽전1길 31"},
	"광주지방법원 목포지원":      postalAddress{postalCode: "58671", address: "전남광주통합특별시 목포시 정의로 29"},
	"광주지방법원 장흥지원":      postalAddress{postalCode: "59330", address: "전남광주통합특별시 장흥군 장흥읍 읍성로 121-1"},
	"광주지방법원 순천지원":      postalAddress{postalCode: "57932", address: "전남광주통합특별시 순천시 왕지로 21"},
	"광주지방법원 해남지원":      postalAddress{postalCode: "59031", address: "전남광주통합특별시 해남군 해남읍 중앙1로 330"},
	"전주지방법원 군산지원":      postalAddress{postalCode: "54079", address: "전북특별자치도 군산시 법원로 68"},
	"전주지방법원 정읍지원":      postalAddress{postalCode: "56172", address: "전북특별자치도 정읍시 수성6로 29"},
	"전주지방법원 남원지원":      postalAddress{postalCode: "55761", address: "전북특별자치도 남원시 용성로 59"},
	"서울가정법원":           postalAddress{postalCode: "06749", address: "서울특별시 서초구 강남대로 193"},
	"인천가정법원":           postalAddress{postalCode: "22134", address: "인천광역시 미추홀구 경원대로 881"},
	"수원가정법원":           postalAddress{postalCode: "16703", address: "경기도 수원시 영통구 매영로345번길 67"},
	"대전가정법원":           postalAddress{postalCode: "35239", address: "대전광역시 서구 둔산중로 69"},
	"대구가정법원":           postalAddress{postalCode: "42635", address: "대구광역시 달서구 장산남로 30"},
	"부산가정법원":           postalAddress{postalCode: "47510", address: "부산광역시 연제구 법원로 31"},
	"울산가정법원":           postalAddress{postalCode: "44643", address: "울산광역시 남구 법대로 55, 10층"},
	"광주가정법원":           postalAddress{postalCode: "61946", address: "전남광주통합특별시 서구 상무번영로 85"},
	"인천가정법원 부천지원":      postalAddress{postalCode: "14602", address: "경기도 부천시 원미구 상일로 129"},
	"수원가정법원 성남지원":      postalAddress{postalCode: "13143", address: "경기도 성남시 수정구 산성대로 451"},
	"수원가정법원 여주지원":      postalAddress{postalCode: "12638", address: "경기도 여주시 현암로 21-12"},
	"수원가정법원 평택지원":      postalAddress{postalCode: "17848", address: "경기도 평택시 평남로 1036"},
	"수원가정법원 안산지원":      postalAddress{postalCode: "15472", address: "경기도 안산시 단원구 광덕서로 75"},
	"수원가정법원 안양지원":      postalAddress{postalCode: "14054", address: "경기도 안양시 동안구 관평로212번길 70"},
	"대전가정법원 홍성지원":      postalAddress{postalCode: "32226", address: "충청남도 홍성군 홍성읍 법원로 38"},
	"대전가정법원 공주지원":      postalAddress{postalCode: "32568", address: "충청남도 공주시 한적2길 34-15"},
	"대전가정법원 논산지원":      postalAddress{postalCode: "32930", address: "충청남도 논산시 강경읍 계백로 99"},
	"대전가정법원 서산지원":      postalAddress{postalCode: "31988", address: "충청남도 서산시 공림4로 24"},
	"대전가정법원 천안지원":      postalAddress{postalCode: "31198", address: "충청남도 천안시 동남구 청수14로 77"},
	"대구가정법원 안동지원":      postalAddress{postalCode: "36752", address: "경상북도 안동시 강남로 304"},
	"대구가정법원 경주지원":      postalAddress{postalCode: "38145", address: "경상북도 경주시 화랑로 89"},
	"대구가정법원 포항지원":      postalAddress{postalCode: "37573", address: "경상북도 포항시 북구 법원로 181"},
	"대구가정법원 김천지원":      postalAddress{postalCode: "39518", address: "경상북도 김천시 물망골길 39"},
	"대구가정법원 상주지원":      postalAddress{postalCode: "37160", address: "경상북도 상주시 북천로 17-9"},
	"대구가정법원 의성지원":      postalAddress{postalCode: "37339", address: "경상북도 의성군 의성읍 군청길 67"},
	"대구가정법원 영덕지원":      postalAddress{postalCode: "36427", address: "경상북도 영덕군 영덕읍 경동로 8337"},
	"광주가정법원 목포지원":      postalAddress{postalCode: "58671", address: "전남광주통합특별시 목포시 정의로 29"},
	"광주가정법원 장흥지원":      postalAddress{postalCode: "59330", address: "전남광주통합특별시 장흥군 장흥읍 읍성로 121-1"},
	"광주가정법원 순천지원":      postalAddress{postalCode: "57932", address: "전남광주통합특별시 순천시 왕지로 21"},
	"광주가정법원 해남지원":      postalAddress{postalCode: "59031", address: "전남광주통합특별시 해남군 해남읍 중앙1로 330"},
	"서울행정법원":           postalAddress{postalCode: "06749", address: "서울특별시 서초구 강남대로 193"},
	"서울회생법원":           postalAddress{postalCode: "06265", address: "서울특별시 강남구 강남대로 262"},
	"수원회생법원":           postalAddress{postalCode: "16512", address: "경기도 수원시 영통구 법조로 105"},
	"대전회생법원":           postalAddress{postalCode: "35237", address: "대전광역시 서구 둔산중로78번길 45"},
	"대구회생법원":           postalAddress{postalCode: "42027", address: "대구광역시 수성구 동대구로 364"},
	"부산회생법원":           postalAddress{postalCode: "47510", address: "부산광역시 연제구 법원로 31"},
	"광주회생법원":           postalAddress{postalCode: "61441", address: "전남광주통합특별시 동구 준법로 7-12, 별관 3층"},
	"의정부지방법원 포천시법원":    postalAddress{postalCode: "11149", address: "경기도 포천시 중앙로 27"},
	"의정부지방법원 연천군법원":    postalAddress{postalCode: "11013", address: "경기도 연천군 연천읍 연천로 376"},
	"의정부지방법원 철원군법원":    postalAddress{postalCode: "24040", address: "강원특별자치도 철원군 삼부연로 59"},
	"의정부지방법원 동두천시법원":   postalAddress{postalCode: "11362", address: "경기도 동두천시 지행로 97"},
	"의정부지방법원 파주시법원":    postalAddress{postalCode: "10894", address: "경기도 파주시 금바위로 40"}, // 홈페이지 이름: 의정부지방법원 고양지원 파주시법원
	"의정부지방법원 가평군법원":    postalAddress{postalCode: "12412", address: "경기도 가평군 가화로 180"}, // 홈페이지 이름: 의정부지방법원 남양주지원 가평군법원
	"인천지방법원 강화군법원":     postalAddress{postalCode: "23025", address: "인천광역시 강화군 강화읍 강화대로 300"},
	"인천지방법원 김포시법원":     postalAddress{postalCode: "10108", address: "경기도 김포시 봉화로 16"}, // 홈페이지 이름: 인천지방법원 부천지원 김포시법원
	"수원지방법원 용인시법원":     postalAddress{postalCode: "17020", address: "경기도 용인시 처인구 성산로 7"},
	"수원지방법원 오산시법원":     postalAddress{postalCode: "18122", address: "경기도 오산시 법원로 65"},
	"수원지방법원 광주시법원":     postalAddress{postalCode: "12746", address: "경기도 광주시 행정타운로 49-15"}, // 홈페이지 이름: 수원지방법원 성남지원 광주시법원
	"수원지방법원 양평군법원":     postalAddress{postalCode: "12554", address: "경기도 양평군 양근로 154"},     // 홈페이지 이름: 수원지방법원 여주지원 양평군법원
	"수원지방법원 이천시법원":     postalAddress{postalCode: "17378", address: "경기도 이천시 구만리로 308"},    // 홈페이지 이름: 수원지방법원 여주지원 이천시법원
	"수원지방법원 안성시법원":     postalAddress{postalCode: "17584", address: "경기도 안성시 안성맞춤대로 1188"}, // 홈페이지 이름: 수원지방법원 평택지원 안성시법원
	"수원지방법원 광명시법원":     postalAddress{postalCode: "14235", address: "경기도 광명시 철산로 3-16"},    // 홈페이지 이름: 수원지방법원 안산지원 광명시법원
	"춘천지방법원 인제군법원":     postalAddress{postalCode: "24631", address: "강원특별자치도 인제군 인제읍 인제로187번길 14"},
	"춘천지방법원 홍천군법원":     postalAddress{postalCode: "25136", address: "강원특별자치도 홍천군 홍천읍 홍천로5길 8"},
	"춘천지방법원 양구군법원":     postalAddress{postalCode: "24522", address: "강원특별자치도 양구군 양구읍 관공서로 30"},
	"춘천지방법원 화천군법원":     postalAddress{postalCode: "24117", address: "강원특별자치도 화천군 화천읍 상승로 102"},
	"춘천지방법원 삼척시법원":     postalAddress{postalCode: "25924", address: "강원특별자치도 삼척시 청석로3길 16-17"}, // 홈페이지 이름: 춘천지방법원 강릉지원 삼척시법원
	"춘천지방법원 동해시법원":     postalAddress{postalCode: "25769", address: "강원특별자치도 동해시 해안로 217"},     // 홈페이지 이름: 춘천지방법원 강릉지원 동해시법원
	"춘천지방법원 횡성군법원":     postalAddress{postalCode: "25232", address: "강원특별자치도 횡성군 횡성읍 읍상로 36"},  // 홈페이지 이름: 춘천지방법원 원주지원 횡성군법원
	"춘천지방법원 고성군법원":     postalAddress{postalCode: "24735", address: "강원특별자치도 고성군 간성읍 간성로 84"},  // 홈페이지 이름: 춘천지방법원 속초지원 고성군법원
	"춘천지방법원 양양군법원":     postalAddress{postalCode: "25029", address: "강원특별자치도 양양군 양양읍 양양로 93"},  // 홈페이지 이름: 춘천지방법원 속초지원 양양군법원
	"춘천지방법원 정선군법원":     postalAddress{postalCode: "26130", address: "강원특별자치도 정선군 정선읍 봉양7길 7"},  // 홈페이지 이름: 춘천지방법원 영월지원 정선군법원
	"춘천지방법원 태백시법원":     postalAddress{postalCode: "26017", address: "강원특별자치도 태백시 문화로1길 25"},    // 홈페이지 이름: 춘천지방법원 영월지원 태백시법원
	"춘천지방법원 평창군법원":     postalAddress{postalCode: "25375", address: "강원특별자치도 평창군 평창읍 백오로 153"}, // 홈페이지 이름: 춘천지방법원 영월지원 평창군법원
	"대전지방법원 세종특별자치시법원": postalAddress{postalCode: "30024", address: "세종특별자치시 조치원읍 새내16길 27"},
	"대전지방법원 금산군법원":     postalAddress{postalCode: "32737", address: "충청남도 금산군 금산읍 인삼로 7"},
	"대전지방법원 서천군법원":     postalAddress{postalCode: "33675", address: "충청남도 서천군 장항읍 장항로 245"},   // 홈페이지 이름: 대전지방법원 홍성지원 서천군법원
	"대전지방법원 보령시법원":     postalAddress{postalCode: "33453", address: "충청남도 보령시 중앙로 128"},       // 홈페이지 이름: 대전지방법원 홍성지원 보령시법원
	"대전지방법원 예산군법원":     postalAddress{postalCode: "32428", address: "충청남도 예산군 예산읍 벚꽃로 145"},   // 홈페이지 이름: 대전지방법원 홍성지원 예산군법원
	"대전지방법원 청양군법원":     postalAddress{postalCode: "33330", address: "충청남도 청양군 청양읍 중앙로11길 20"}, // 홈페이지 이름: 대전지방법원 공주지원 청양군법원
	"대전지방법원 부여군법원":     postalAddress{postalCode: "33167", address: "충청남도 부여군 부여읍 사비로 28"},    // 홈페이지 이름: 대전지방법원 논산지원 부여군법원
	"대전지방법원 태안군법원":     postalAddress{postalCode: "32142", address: "충청남도 태안군 태안읍 탑골길 71"},    // 홈페이지 이름: 대전지방법원 서산지원 태안군법원
	"대전지방법원 당진시법원":     postalAddress{postalCode: "31776", address: "충청남도 당진시 청룡길 140"},       // 홈페이지 이름: 대전지방법원 서산지원 당진시법원
	"대전지방법원 아산시법원":     postalAddress{postalCode: "31576", address: "충청남도 아산시 용화로76번길 7"},     // 홈페이지 이름: 대전지방법원 천안지원 아산시법원
	"청주지방법원 보은군법원":     postalAddress{postalCode: "28943", address: "충청북도 보은군 보은읍 보은로 147"},
	"청주지방법원 괴산군법원":     postalAddress{postalCode: "28035", address: "충청북도 괴산군 괴산읍 읍내로5길 20"},
	"청주지방법원 진천군법원":     postalAddress{postalCode: "27840", address: "충청북도 진천군 진천읍 남산3길 6"},
	"청주지방법원 음성군법원":     postalAddress{postalCode: "27691", address: "충청북도 음성군 음성읍 용광로 55"}, // 홈페이지 이름: 청주지방법원 충주지원 음성군법원
	"청주지방법원 단양군법원":     postalAddress{postalCode: "27010", address: "충청북도 단양군 단양읍 중앙1로 9"}, // 홈페이지 이름: 청주지방법원 제천지원 단양군법원
	"청주지방법원 옥천군법원":     postalAddress{postalCode: "29039", address: "충청북도 옥천군 옥천읍 중앙로 96"}, // 홈페이지 이름: 청주지방법원 영동지원 옥천군법원
	"대구지방법원 청도군법원":     postalAddress{postalCode: "38329", address: "경상북도 청도군 화양읍 청화로 31"},
	"대구지방법원 영천시법원":     postalAddress{postalCode: "38845", address: "경상북도 영천시 천문로 353"},
	"대구지방법원 칠곡군법원":     postalAddress{postalCode: "39903", address: "경상북도 칠곡군 왜관읍 전원4길 21"},
	"대구지방법원 경산시법원":     postalAddress{postalCode: "38583", address: "경상북도 경산시 박물관로1길 9"},
	"대구지방법원 성주군법원":     postalAddress{postalCode: "40030", address: "경상북도 성주군 성주읍 성주순환로 256"},  // 홈페이지 이름: 대구지방법원 서부지원 성주군법원
	"대구지방법원 고령군법원":     postalAddress{postalCode: "40135", address: "경상북도 고령군 대가야읍 대가야로 1441"}, // 홈페이지 이름: 대구지방법원 서부지원 고령군법원
	"대구지방법원 영주시법원":     postalAddress{postalCode: "36057", address: "경상북도 영주시 영주로 105"},        // 홈페이지 이름: 대구지방법원 안동지원 영주시법원
	"대구지방법원 봉화군법원":     postalAddress{postalCode: "36236", address: "경상북도 봉화군 봉화읍 봉화로 1085"},   // 홈페이지 이름: 대구지방법원 안동지원 봉화군법원
	"대구지방법원 구미시법원":     postalAddress{postalCode: "39210", address: "경상북도 구미시 봉곡로10길 5-8"},     // 홈페이지 이름: 대구지방법원 김천지원 구미시법원
	"대구지방법원 예천군법원":     postalAddress{postalCode: "36826", address: "경상북도 예천군 예천읍 충효로 81"},     // 홈페이지 이름: 대구지방법원 상주지원 예천군법원
	"대구지방법원 문경시법원":     postalAddress{postalCode: "36975", address: "경상북도 문경시 매봉로 91"},         // 홈페이지 이름: 대구지방법원 상주지원 문경시법원
	"대구지방법원 청송군법원":     postalAddress{postalCode: "37427", address: "경상북도 청송군 청송읍 중앙로 319"},    // 홈페이지 이름: 대구지방법원 의성지원 청송군법원
	"대구지방법원 군위군법원":     postalAddress{postalCode: "43114", address: "대구광역시 군위군 군위읍 도군로 2734"},  // 홈페이지 이름: 대구지방법원 의성지원 군위군법원
	"대구지방법원 울진군법원":     postalAddress{postalCode: "36326", address: "경상북도 울진군 울진읍 월변2길 33"},    // 홈페이지 이름: 대구지방법원 영덕지원 울진군법원
	"대구지방법원 영양군법원":     postalAddress{postalCode: "36541", address: "경상북도 영양군 영양읍 동서대로 35"},    // 홈페이지 이름: 대구지방법원 영덕지원 영양군법원
	"울산지방법원 양산시법원":     postalAddress{postalCode: "50618", address: "경상남도 양산시 북안남5길 12"},
	"창원지방법원 창원남부시법원":   postalAddress{postalCode: "51595", address: "경상남도 창원시 진해구 진해대로 833"},
	"창원지방법원 김해시법원":     postalAddress{postalCode: "50944", address: "경상남도 김해시 우암로 167"},
	"창원지방법원 함안군법원":     postalAddress{postalCode: "52040", address: "경상남도 함안군 가야읍 가야13길 6"}, // 홈페이지 이름: 창원지방법원 마산지원 함안군법원
	"창원지방법원 의령군법원":     postalAddress{postalCode: "52145", address: "경상남도 의령군 의령읍 백산로 12"},  // 홈페이지 이름: 창원지방법원 마산지원 의령군법원
	"창원지방법원 하동군법원":     postalAddress{postalCode: "52333", address: "경상남도 하동군 하동읍 군청로 33"},  // 홈페이지 이름: 창원지방법원 진주지원 하동군법원
	"창원지방법원 사천시법원":     postalAddress{postalCode: "52539", address: "경상남도 사천시 용현면 시청1길 39"}, // 홈페이지 이름: 창원지방법원 진주지원 사천시법원
	"창원지방법원 남해군법원":     postalAddress{postalCode: "52415", address: "경상남도 남해군 남해읍 화전로 99"},  // 홈페이지 이름: 창원지방법원 진주지원 남해군법원
	"창원지방법원 산청군법원":     postalAddress{postalCode: "52219", address: "경상남도 산청군 산청읍 중앙로 27"},  // 홈페이지 이름: 창원지방법원 진주지원 산청군법원
	"창원지방법원 거제시법원":     postalAddress{postalCode: "53252", address: "경상남도 거제시 거제중앙로31길 32"}, // 홈페이지 이름: 창원지방법원 통영지원 거제시법원
	"창원지방법원 고성군법원":     postalAddress{postalCode: "52947", address: "경상남도 고성군 고성읍 동외로 110"}, // 홈페이지 이름: 창원지방법원 통영지원 고성군법원
	"창원지방법원 창녕군법원":     postalAddress{postalCode: "50333", address: "경상남도 창녕군 창녕읍 남창녕로 26"}, // 홈페이지 이름: 창원지방법원 밀양지원 창녕군법원
	"창원지방법원 합천군법원":     postalAddress{postalCode: "50231", address: "경상남도 합천군 합천읍 동서로 105"}, // 홈페이지 이름: 창원지방법원 거창지원 합천군법원
	"창원지방법원 함양군법원":     postalAddress{postalCode: "50032", address: "경상남도 함양군 함양읍 학사루길 8"},  // 홈페이지 이름: 창원지방법원 거창지원 함양군법원
	"광주지방법원 곡성군법원":     postalAddress{postalCode: "57542", address: "전남광주통합특별시 곡성군 곡성읍 곡성로 845"},
	"광주지방법원 영광군법원":     postalAddress{postalCode: "57051", address: "전남광주통합특별시 영광군 영광읍 중앙로 210"},
	"광주지방법원 나주시법원":     postalAddress{postalCode: "58262", address: "전남광주통합특별시 나주시 재신길 41"},
	"광주지방법원 장성군법원":     postalAddress{postalCode: "57220", address: "전남광주통합특별시 장성군 장성읍 영천로 213"},
	"광주지방법원 화순군법원":     postalAddress{postalCode: "58112", address: "전남광주통합특별시 화순군 화순읍 동헌길 21-18"},
	"광주지방법원 담양군법원":     postalAddress{postalCode: "57345", address: "전남광주통합특별시 담양군 담양읍 추성로 1354"},
	"광주지방법원 함평군법원":     postalAddress{postalCode: "57153", address: "전남광주통합특별시 함평군 함평읍 중앙길 36"},  // 홈페이지 이름: 광주지방법원 목포지원 함평군법원
	"광주지방법원 영암군법원":     postalAddress{postalCode: "58414", address: "전남광주통합특별시 영암군 영암읍 서남역로 21"}, // 홈페이지 이름: 광주지방법원 목포지원 영암군법원
	"광주지방법원 무안군법원":     postalAddress{postalCode: "58532", address: "전남광주통합특별시 무안군 무안읍 무안로 598"}, // 홈페이지 이름: 광주지방법원 목포지원 무안군법원
	"광주지방법원 강진군법원":     postalAddress{postalCode: "59225", address: "전남광주통합특별시 강진군 강진읍 중앙로 212"}, // 홈페이지 이름: 광주지방법원 장흥지원 강진군법원
	"광주지방법원 보성군법원":     postalAddress{postalCode: "59455", address: "전남광주통합특별시 보성군 보성읍 송재로 168"}, // 홈페이지 이름: 광주지방법원 순천지원 보성군법원
	"광주지방법원 고흥군법원":     postalAddress{postalCode: "59535", address: "전남광주통합특별시 고흥군 고흥읍 터미널길 7"},  // 홈페이지 이름: 광주지방법원 순천지원 고흥군법원
	"광주지방법원 여수시법원":     postalAddress{postalCode: "59691", address: "전남광주통합특별시 여수시 예울마루로 53"},    // 홈페이지 이름: 광주지방법원 순천지원 여수시법원
	"광주지방법원 구례군법원":     postalAddress{postalCode: "57654", address: "전남광주통합특별시 구례군 구례읍 학교길 91"},  // 홈페이지 이름: 광주지방법원 순천지원 구례군법원
	"광주지방법원 광양시법원":     postalAddress{postalCode: "57785", address: "전남광주통합특별시 광양시 중동로 80"},      // 홈페이지 이름: 광주지방법원 순천지원 광양시법원
	"광주지방법원 완도군법원":     postalAddress{postalCode: "59126", address: "전남광주통합특별시 완도군 완도읍 중앙길 57"},  // 홈페이지 이름: 광주지방법원 해남지원 완도군법원
	"광주지방법원 진도군법원":     postalAddress{postalCode: "58926", address: "전남광주통합특별시 진도군 진도읍 쌍정1길 16"}, // 홈페이지 이름: 광주지방법원 해남지원 진도군법원
	"전주지방법원 진안군법원":     postalAddress{postalCode: "55433", address: "전북특별자치도 진안군 진안읍 우화산길 8"},
	"전주지방법원 김제시법원":     postalAddress{postalCode: "54404", address: "전북특별자치도 김제시 중앙로 239"},
	"전주지방법원 무주군법원":     postalAddress{postalCode: "55519", address: "전북특별자치도 무주군 무주읍 향학로 71"},
	"전주지방법원 임실군법원":     postalAddress{postalCode: "55928", address: "전북특별자치도 임실군 임실읍 중동로 62"},
	"전주지방법원 익산시법원":     postalAddress{postalCode: "54600", address: "전북특별자치도 익산시 주현로 39"},      // 홈페이지 이름: 전주지방법원 군산지원 익산시법원
	"전주지방법원 부안군법원":     postalAddress{postalCode: "56307", address: "전북특별자치도 부안군 부안읍 용계길 17"},  // 홈페이지 이름: 전주지방법원 정읍지원 부안군법원
	"전주지방법원 고창군법원":     postalAddress{postalCode: "56428", address: "전북특별자치도 고창군 고창읍 교촌5길 6"},  // 홈페이지 이름: 전주지방법원 정읍지원 고창군법원
	"전주지방법원 장수군법원":     postalAddress{postalCode: "55630", address: "전북특별자치도 장수군 장수읍 싸리재로 13"}, // 홈페이지 이름: 전주지방법원 남원지원 장수군법원
	"전주지방법원 순창군법원":     postalAddress{postalCode: "56039", address: "전북특별자치도 순창군 순창읍 순창5길 36"}, // 홈페이지 이름: 전주지방법원 남원지원 순창군법원
	"제주지방법원 서귀포시법원":    postalAddress{postalCode: "63584", address: "제주특별자치도 서귀포시 일주동로 8690"},
}
