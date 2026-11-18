package casepulse

// 이 파일은 Courts가 돌려주는 법원 목록의 내용을 검사합니다.
// 트리 출력 테스트는 courts_tree_test.go에 있습니다.

import (
	"strings"
	"testing"
)

// TestCourtDirectory는 법원 목록 전체의 구조를 검사합니다.
// 법원 수, ID 중복, 상위 법원 참조, 시·군법원 수, 주요 재판부, 재판부 ID와 소재지, 원외 고등법원 정의를 확인합니다.
func TestCourtDirectory(t *testing.T) {
	courts := Courts()
	if len(courts) != 215 {
		t.Fatalf("Courts() returned %d courts, want 215", len(courts))
	}

	byID := make(map[string]Court, len(courts))
	localCourtCount := 0
	for _, court := range courts {
		if court.ID == "" {
			t.Fatal("court has an empty ID")
		}
		if _, exists := byID[court.ID]; exists {
			t.Fatalf("duplicate court ID %q", court.ID)
		}
		if court.Type != InstitutionTypeCourt {
			t.Errorf("%s has institution type %q, want %q", court.ID, court.Type, InstitutionTypeCourt)
		}
		if court.UnitType == CourtUnitTypeMunicipalCounty {
			localCourtCount++
		}
		byID[court.ID] = court
	}
	if localCourtCount != 99 {
		t.Errorf("found %d municipal/county courts, want 99", localCourtCount)
	}

	for _, court := range courts {
		if court.ParentCourtID != "" {
			if _, exists := byID[court.ParentCourtID]; !exists {
				t.Errorf("%s references missing parent %q", court.ID, court.ParentCourtID)
			}
		}
	}

	expectedPanels := map[string][]string{
		"court-high-seoul": {
			"제1-1민사부",
		},
		"court-high-daejeon": {
			"제1행정부",
		},
		"court-high-busan": {
			"제1행정부",
		},
		"court-high-seoul-incheon": {
			"제1민사부", "제2행정부",
		},
		"court-high-seoul-chuncheon": {
			"제1민사부", "제2형사부",
		},
		"court-high-daejeon-cheongju": {
			"청주제1행정부", "청주 조정전담",
		},
		"court-high-busan-ulsan": {
			"울산제1행정부", "울산제2형사부",
		},
		"court-high-busan-changwon": {
			"창원제1행정부", "창원제4민사부",
		},
		"court-district-seoul-central": {
			"제91민사단독", "제1형사단독",
		},
		"court-family-seoul": {
			"제1가사부", "가사제3단독",
		},
		"court-rehabilitation-seoul": {
			"제1부", "제11부",
		},
	}
	panelIDs := make(map[string]bool)
	panelCount := 0
	var checkPanels func(courtID, parentSeat string, panels []CourtPanel)
	checkPanels = func(courtID, parentSeat string, panels []CourtPanel) {
		names := make(map[string]bool, len(panels))
		for _, panel := range panels {
			if panel.ID == "" || panel.Name == "" || panel.Seat == "" {
				t.Errorf("%s has incomplete panel data: %+v", courtID, panel)
			}
			if panelIDs[panel.ID] {
				t.Errorf("duplicate panel ID %q", panel.ID)
			}
			panelIDs[panel.ID] = true
			if names[panel.Name] {
				t.Errorf("%s has duplicate panel name %q", courtID, panel.Name)
			}
			names[panel.Name] = true
			if panel.Seat != parentSeat {
				t.Errorf("%s panel %q is at %q, want parent seat %q", courtID, panel.Name, panel.Seat, parentSeat)
			}
			panelCount++
		}
	}
	for _, court := range courts {
		checkPanels(court.ID, court.Seat, court.Panels)
		for _, expectedName := range expectedPanels[court.ID] {
			if !containsCourtPanel(court.Panels, expectedName) {
				t.Errorf("%s is missing panel %q", court.ID, expectedName)
			}
		}
	}
	if panelCount != 2883 {
		t.Errorf("found %d panels, want 2883", panelCount)
	}

	remoteHighCourts := map[string]struct {
		name, parentID, seat string
	}{
		"court-high-seoul-incheon":    {"서울고등법원(인천)", "court-supreme", "인천광역시"},
		"court-high-seoul-chuncheon":  {"서울고등법원(춘천)", "court-supreme", "춘천시"},
		"court-high-daejeon-cheongju": {"대전고등법원(청주)", "court-supreme", "청주시"},
		"court-high-busan-ulsan":      {"부산고등법원(울산)", "court-supreme", "울산광역시"},
		"court-high-busan-changwon":   {"부산고등법원(창원)", "court-supreme", "창원시"},
		"court-high-gwangju-jeonju":   {"광주고등법원(전주)", "court-supreme", "전주시"},
		"court-high-gwangju-jeju":     {"광주고등법원(제주)", "court-supreme", "제주시"},
	}
	for id, want := range remoteHighCourts {
		court, exists := byID[id]
		if !exists {
			t.Errorf("remote high court %q is missing", id)
			continue
		}
		if court.Name != want.name || court.Level != CourtLevelHigh ||
			court.CourtType != CourtTypeGeneral || court.UnitType != CourtUnitTypeMain ||
			court.ParentCourtID != want.parentID || court.Seat != want.seat {
			t.Errorf("remote high court %q is incorrectly defined: %+v", id, court)
		}
		if len(court.Panels) == 0 {
			t.Errorf("remote high court %q has no panels", id)
		}
	}
	for _, id := range []string{
		"court-high-seoul",
		"court-high-daejeon",
		"court-high-busan",
	} {
		for _, panel := range byID[id].Panels {
			if strings.Contains(panel.Name, "원외재판부") || strings.Contains(panel.Name, "청주재판부") {
				t.Errorf("%s still contains promoted court panel %q", id, panel.Name)
			}
		}
	}

	bucheon, exists := byID["court-branch-incheon-bucheon"]
	if !exists {
		t.Fatal("Incheon District Court Bucheon Branch is missing")
	}
	if bucheon.Name != "인천지방법원 부천지원" ||
		bucheon.ParentCourtID != "court-district-incheon" ||
		bucheon.UnitType != CourtUnitTypeSupport {
		t.Errorf("Bucheon branch is incorrectly defined: %+v", bucheon)
	}

	courts[0].Jurisdictions[0] = "modified"
	if Courts()[0].Jurisdictions[0] == "modified" {
		t.Fatal("Courts returned a shared jurisdiction slice")
	}
	modifiedPanel := false
	for i := range courts {
		if len(courts[i].Panels) == 0 {
			continue
		}
		courts[i].Panels[0].Name = "modified"
		if Courts()[i].Panels[0].Name == "modified" {
			t.Fatal("Courts returned a shared panel slice")
		}
		modifiedPanel = true
		break
	}
	if !modifiedPanel {
		t.Fatal("court directory has no panels to test")
	}
}

// findCourtPanel은 재판부 목록에서 이름이 같은 재판부를 찾아 돌려줍니다. 없으면 nil입니다.
func findCourtPanel(panels []CourtPanel, name string) *CourtPanel {
	for i := range panels {
		if panels[i].Name == name {
			return &panels[i]
		}
	}
	return nil
}

// containsCourtPanel은 재판부 목록에 해당 이름의 재판부가 있는지 알려 줍니다.
func containsCourtPanel(panels []CourtPanel, name string) bool {
	return findCourtPanel(panels, name) != nil
}

// TestCourtAddresses는 법원 주소가 모두 채워졌는지(주소 없는 예외 법원 제외)와
// 별표 2에 맞춘 시·군법원의 소속·이름이 올바른지 확인합니다.
func TestCourtAddresses(t *testing.T) {
	byName := make(map[string]Court)
	for _, court := range Courts() {
		byName[court.Name] = court
		if courtsWithoutAddress[court.Name] {
			if court.Address != "" {
				t.Errorf("%s is listed without address but has %q", court.Name, court.Address)
			}
			continue
		}
		if len(court.PostalCode) != 5 || court.Address == "" {
			t.Errorf("%s has postal code %q, address %q", court.Name, court.PostalCode, court.Address)
		}
	}

	for name, want := range map[string]string{
		"대법원":          "서울특별시 서초구 서초대로 219",
		"서울고등법원(인천)":   "인천광역시 미추홀구 소성로163번길 17",
		"울산가정법원":       "울산광역시 남구 법대로 55, 10층",
		"인천지방법원 김포시법원": "경기도 김포시 봉화로 16",
	} {
		if got := byName[name].Address; got != want {
			t.Errorf("%s Address = %q, want %q", name, got, want)
		}
	}

	// 별표 2에 맞춘 시·군법원 소속과 이름을 확인합니다.
	gimpo := byName["인천지방법원 김포시법원"]
	if gimpo.ParentCourtID != "court-branch-incheon-bucheon" {
		t.Errorf("김포시법원 ParentCourtID = %q, want 부천지원", gimpo.ParentCourtID)
	}
	jincheon := byName["청주지방법원 진천군법원"]
	if jincheon.ParentCourtID != "court-district-cheongju" {
		t.Errorf("진천군법원 ParentCourtID = %q, want 청주지방법원", jincheon.ParentCourtID)
	}
	for _, name := range []string{"인천지방법원 강화군법원", "춘천지방법원 횡성군법원", "대전지방법원 세종특별자치시법원"} {
		if _, exists := byName[name]; !exists {
			t.Errorf("%s is missing", name)
		}
	}
	for _, name := range []string{"인천지방법원 부천시법원", "청주지방법원 증평군법원", "부산가정법원 울산지원"} {
		if _, exists := byName[name]; exists {
			t.Errorf("%s should not exist", name)
		}
	}
}

// TestGwangjuRemotePanels는 전주·제주 원외재판부의 재판부가 광주고등법원 본원이 아니라
// 각 원외 고등법원 항목에 들어 있는지 확인합니다.
func TestGwangjuRemotePanels(t *testing.T) {
	byID := make(map[string]Court)
	for _, court := range Courts() {
		byID[court.ID] = court
	}
	for id, prefix := range map[string]string{
		"court-high-gwangju-jeonju": "전주제",
		"court-high-gwangju-jeju":   "제주제",
	} {
		court := byID[id]
		if len(court.Panels) == 0 {
			t.Errorf("%s has no panels", id)
		}
		for _, panel := range court.Panels {
			if !strings.HasPrefix(panel.Name, prefix) {
				t.Errorf("%s has unexpected panel %q", id, panel.Name)
			}
		}
	}
	for _, panel := range byID["court-high-gwangju"].Panels {
		if strings.HasPrefix(panel.Name, "전주제") || strings.HasPrefix(panel.Name, "제주제") {
			t.Errorf("광주고등법원 still has remote panel %q", panel.Name)
		}
	}
}

// TestMunicipalCourtSeatsFollowByeolpyo2는 시·군법원 소재지가 별표 2의 소재지 열과 같은지 확인합니다.
// 시·군법원 수와 소재지 표의 건수가 같아야 표의 모든 항목이 쓰인 것입니다.
func TestMunicipalCourtSeatsFollowByeolpyo2(t *testing.T) {
	byName := make(map[string]Court)
	count := 0
	for _, court := range Courts() {
		byName[court.Name] = court
		if court.UnitType == CourtUnitTypeMunicipalCounty {
			count++
		}
	}
	if count != len(municipalCourtSeats) {
		t.Errorf("municipal courts = %d, seat table = %d", count, len(municipalCourtSeats))
	}
	for name, want := range map[string]string{
		"의정부지방법원 철원군법원":    "갈말읍",
		"대구지방법원 칠곡군법원":     "왜관읍",
		"창원지방법원 거제시법원":     "신현읍",
		"창원지방법원 함안군법원":     "가야읍",
		"창원지방법원 고성군법원":     "고성읍",
		"춘천지방법원 고성군법원":     "간성읍",
		"대전지방법원 서천군법원":     "장항읍",
		"창원지방법원 창원남부시법원":   "창원시 진해구",
		"대전지방법원 세종특별자치시법원": "세종특별자치시",
		"인천지방법원 김포시법원":     "김포시",
	} {
		if got := byName[name].Seat; got != want {
			t.Errorf("%s Seat = %q, want %q", name, got, want)
		}
	}
}

// TestGwangjuIsNormalized는 법원의 소재지·주소·관할 지역에 옛 명칭("광주광역시", "전라남도")이 남아 있지 않은지 확인합니다.
func TestGwangjuIsNormalized(t *testing.T) {
	for _, court := range Courts() {
		if strings.Contains(court.Seat, "광주광역시") || strings.Contains(court.Address, "광주광역시") {
			t.Errorf("%s still uses 광주광역시: seat %q, address %q", court.Name, court.Seat, court.Address)
		}
		for _, area := range court.Jurisdictions {
			if strings.Contains(area, "광주광역시") || strings.Contains(area, "전라남도") {
				t.Errorf("%s jurisdiction %q is not normalized", court.Name, area)
			}
		}
	}
	for _, id := range []string{"court-high-gwangju", "court-district-gwangju", "court-family-gwangju", "court-rehabilitation-gwangju"} {
		for _, court := range Courts() {
			if court.ID == id && court.Seat != "전남광주통합특별시" {
				t.Errorf("%s Seat = %q, want 전남광주통합특별시", id, court.Seat)
			}
		}
	}
}

// TestFamilyBranchSeatsFollowByeolpyo1은 가정법원 지원의 소재지가 별표 1처럼 읍 단위로 적혀 있는지 확인합니다.
// 같은 지역의 지방법원 지원과 소재지가 같아야 합니다(예: 대전지방법원 홍성지원과 대전가정법원 홍성지원은 모두 "홍성읍").
func TestFamilyBranchSeatsFollowByeolpyo1(t *testing.T) {
	byName := make(map[string]Court)
	for _, court := range Courts() {
		byName[court.Name] = court
	}
	for name, want := range map[string]string{
		"대전가정법원 홍성지원": "홍성읍",
		"대구가정법원 의성지원": "의성읍",
		"대구가정법원 영덕지원": "영덕읍",
		"광주가정법원 장흥지원": "장흥읍",
		"광주가정법원 해남지원": "해남읍",
	} {
		if got := byName[name].Seat; got != want {
			t.Errorf("%s Seat = %q, want %q", name, got, want)
		}
	}
}
