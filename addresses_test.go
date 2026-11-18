package casepulse

// 이 파일은 여러 기관 목록에 공통으로 적용되는 주소 정규화 규칙을 검사합니다.

import (
	"strings"
	"testing"
)

// TestAddressesAreNormalized는 법원·공소청·중수청의 모든 주소가 정규화 규칙을 따르는지 확인합니다.
// 정식 시·도 이름으로 시작해야 하고, 괄호와 이중 공백이 없어야 합니다.
func TestAddressesAreNormalized(t *testing.T) {
	var addresses []string
	for _, c := range Courts() {
		addresses = append(addresses, c.Address)
	}
	for _, p := range ProsecutionOffices() {
		addresses = append(addresses, p.Address)
	}
	for _, i := range InvestigationAgencies() {
		addresses = append(addresses, i.Address)
	}
	provinces := []string{"서울특별시", "인천광역시", "대전광역시", "대구광역시", "부산광역시", "울산광역시",
		"세종특별자치시", "경기도", "강원특별자치도", "충청남도", "충청북도", "경상북도", "경상남도",
		"전남광주통합특별시", "전북특별자치도", "제주특별자치도"}
	for _, address := range addresses {
		if address == "" {
			continue
		}
		first := strings.Fields(address)[0]
		known := false
		for _, p := range provinces {
			known = known || first == p
		}
		if !known {
			t.Errorf("address %q does not start with a full province name", address)
		}
		if strings.ContainsAny(address, "()") || strings.Contains(address, "  ") {
			t.Errorf("address %q is not normalized", address)
		}
	}
}
