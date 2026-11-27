package leviathan

import "fmt"

type postalAddress struct {
	postalCode string
	address    string
}

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
