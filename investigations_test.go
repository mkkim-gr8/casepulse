package casepulse

import (
	"strings"
	"testing"
)

// TestInvestigationDirectory는 중대범죄수사청 목록의 구조를 검사합니다.
// 기관 수, ID 중복, 기관 종류, 필수 값, 상위 기관, 약칭 규칙, 전화번호 없음, 복사본 반환을 확인합니다.
func TestInvestigationDirectory(t *testing.T) {
	agencies := InvestigationAgencies()
	// 본청 1곳 + 별표 1의 지방수사청 6곳
	if len(agencies) != 7 {
		t.Fatalf("InvestigationAgencies() returned %d agencies, want 7", len(agencies))
	}

	byID := make(map[string]InvestigationAgency, len(agencies))
	for _, agency := range agencies {
		if _, exists := byID[agency.ID]; exists {
			t.Fatalf("duplicate agency ID %q", agency.ID)
		}
		byID[agency.ID] = agency
	}

	for _, agency := range agencies {
		if agency.Phone != "" {
			t.Errorf("%s has phone %q, want none", agency.ID, agency.Phone)
		}
		if agency.Type != InstitutionTypeInvestigation {
			t.Errorf("%s has institution type %q, want %q", agency.ID, agency.Type, InstitutionTypeInvestigation)
		}
		if agency.Seat == "" || agency.Address == "" || len(agency.Jurisdictions) == 0 {
			t.Errorf("%s has missing seat, address or jurisdictions: %+v", agency.ID, agency)
		}
		switch agency.Level {
		case InvestigationLevelNational:
			if agency.ParentAgencyID != "" {
				t.Errorf("national agency %s has parent %q", agency.ID, agency.ParentAgencyID)
			}
		case InvestigationLevelDistrict:
			parent, exists := byID[agency.ParentAgencyID]
			if !exists || parent.Level != InvestigationLevelNational {
				t.Errorf("%s has unexpected parent %q", agency.ID, agency.ParentAgencyID)
			}
			// 약칭은 정식 명칭에서 "중대범죄"를 뺀 이름입니다.
			if want := strings.Replace(agency.Name, "중대범죄", "", 1); agency.ShortName != want {
				t.Errorf("%s ShortName = %q, want %q", agency.ID, agency.ShortName, want)
			}
		default:
			t.Errorf("%s has unknown level %q", agency.ID, agency.Level)
		}
	}

	daejeon := byID["investigation-district-daejeon"]
	if daejeon.Seat != "세종특별자치시" {
		t.Errorf("대전지방중대범죄수사청 Seat = %q, want %q", daejeon.Seat, "세종특별자치시")
	}
	if got := len(byID["investigation-district-suwon"].Jurisdictions); got != 19 {
		t.Errorf("수원지방중대범죄수사청 has %d jurisdictions, want 19", got)
	}

	// 반환값을 수정해도 내부 원본이 바뀌지 않아야 합니다.
	agencies[1].Jurisdictions[0] = "changed"
	if InvestigationAgencies()[1].Jurisdictions[0] == "changed" {
		t.Error("InvestigationAgencies() exposes the internal jurisdictions")
	}
}
