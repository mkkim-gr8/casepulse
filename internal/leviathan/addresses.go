package leviathan

import "fmt"

// postalAddress는 주소 표(courts_addresses.go 등)의 한 칸으로, 우편번호와 주소를 함께 담습니다.
type postalAddress struct {
	postalCode string
	address    string
}

// applyPostalAddresses는 기관 이름을 열쇠로 주소 표(table)에서 주소를 찾아 각 기관에 채워 넣습니다.
//
// 데이터 실수를 일찍 잡기 위해 두 가지를 검사하고, 틀리면 panic(프로그램 즉시 중단)을 일으킵니다.
//  1. 기관 목록에 있는데 주소 표에 없는 기관(단, withoutAddress에 "주소 없음"으로 등록한 기관은 예외)
//  2. 주소 표에 있는데 기관 목록에 없는 이름(오타 등) — 일치한 개수가 표의 개수와 다르면 발견됩니다.
//
// kind는 오류 메시지에 쓸 기관 종류 이름입니다(예: "법원").
// institutions는 포인터 목록이라, 여기서 값을 바꾸면 원래 기관 자료에 그대로 반영됩니다.
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
