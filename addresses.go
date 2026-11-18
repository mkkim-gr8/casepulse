package casepulse

import "fmt"

// postalAddress는 주소 표 한 줄에 해당하는 우편번호와 주소입니다.
//
// 법원(courtAddresses)과 공소청(prosecutionAddresses)의 주소 표가 같은 형식을 쓰도록
// 공통 자료형으로 둡니다. 두 값은 그대로 Institution.PostalCode와 Institution.Address에 들어갑니다.
type postalAddress struct {
	// postalCode는 5자리 우편번호입니다(예: "06590").
	postalCode string
	// address는 정규화한 도로명 주소이며 우편번호를 포함하지 않습니다.
	// 정규화 규칙은 courts_addresses.go의 courtAddresses 주석을 참고합니다.
	address string
}

// applyPostalAddresses는 기관 이름으로 주소 표를 찾아 각 기관의 우편번호와 주소를 채웁니다.
//
// kind는 오류 메시지에 쓰는 기관 종류 이름(예: "법원")입니다.
// institutions는 값을 채울 기관들의 포인터 목록이며, 표의 키와 Institution.Name이 정확히 같아야 합니다.
// withoutAddress에 들어 있는 이름은 표에 없어도 되는 기관으로, 주소를 비워 둡니다(nil이면 예외 없음).
//
// 다음 두 경우에는 패키지 초기화 시점에 패닉을 일으킵니다.
// 데이터 표를 고치다 생긴 실수를 조용히 넘기지 않고, 프로그램을 시작하는 순간 바로 드러내기 위해서입니다.
//   - 기관 목록에 있는 기관의 주소가 표에 없고, withoutAddress에도 없는 경우
//     (새 기관을 추가하고 주소를 빠뜨렸거나, 기관 이름을 바꾸고 표를 고치지 않은 경우)
//   - 표에 있는 이름 중 기관 목록과 맞지 않는 것이 있는 경우
//     (표의 이름에 오타가 있거나, 폐지·개편된 기관의 주소가 남아 있는 경우)
func applyPostalAddresses(kind string, institutions []*Institution, table map[string]postalAddress, withoutAddress map[string]bool) {
	matched := 0
	for _, institution := range institutions {
		entry, exists := table[institution.Name]
		if !exists {
			if withoutAddress[institution.Name] {
				continue
			}
			panic(fmt.Sprintf("%s 주소를 찾을 수 없습니다: %s", kind, institution.Name))
		}
		institution.PostalCode = entry.postalCode
		institution.Address = entry.address
		matched++
	}
	if matched != len(table) {
		panic(fmt.Sprintf("%s 주소 %d건 중 %d건만 기관 목록과 일치합니다", kind, len(table), matched))
	}
}
