package casepulse

import (
	"strings"
	"testing"
)

func TestInstitutionTypes(t *testing.T) {
	tests := []struct {
		name string
		got  InstitutionType
		want string
	}{
		{name: "court", got: InstitutionTypeCourt, want: "court"},
		{name: "prosecution", got: InstitutionTypeProsecution, want: "prosecution"},
		{name: "police", got: InstitutionTypePolice, want: "police"},
		{name: "registry", got: InstitutionTypeRegistry, want: "registry"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.got) != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestCourtTypes(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "supreme level", got: string(CourtLevelSupreme), want: "supreme"},
		{name: "high level", got: string(CourtLevelHigh), want: "high"},
		{name: "district level", got: string(CourtLevelDistrict), want: "district"},
		{name: "local level", got: string(CourtLevelLocal), want: "local"},
		{name: "general", got: string(CourtTypeGeneral), want: "general"},
		{name: "family", got: string(CourtTypeFamily), want: "family"},
		{name: "administrative", got: string(CourtTypeAdministrative), want: "administrative"},
		{name: "patent", got: string(CourtTypePatent), want: "patent"},
		{name: "rehabilitation", got: string(CourtTypeRehabilitation), want: "rehabilitation"},
		{name: "main unit", got: string(CourtUnitTypeMain), want: "main"},
		{name: "support unit", got: string(CourtUnitTypeSupport), want: "support"},
		{name: "municipal county unit", got: string(CourtUnitTypeMunicipalCounty), want: "municipal_county"},
		{name: "panel at court seat", got: string(CourtPanelTypeAtSeat), want: "at_seat"},
		{name: "panel out of court seat", got: string(CourtPanelTypeOutOfSeat), want: "out_of_seat"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestCourtFields(t *testing.T) {
	got := Court{
		Institution: Institution{
			ID:   "court-seoul-family",
			Name: "서울가정법원",
			Type: InstitutionTypeCourt,
		},
		Level:         CourtLevelDistrict,
		CourtType:     CourtTypeFamily,
		UnitType:      CourtUnitTypeMain,
		Seat:          "서울특별시",
		Jurisdictions: []string{"서울특별시"},
	}

	if got.Type != InstitutionTypeCourt {
		t.Errorf("Type = %q, want %q", got.Type, InstitutionTypeCourt)
	}
	if got.Level != CourtLevelDistrict {
		t.Errorf("Level = %q, want %q", got.Level, CourtLevelDistrict)
	}
	if got.CourtType != CourtTypeFamily {
		t.Errorf("CourtType = %q, want %q", got.CourtType, CourtTypeFamily)
	}
	if got.UnitType != CourtUnitTypeMain {
		t.Errorf("UnitType = %q, want %q", got.UnitType, CourtUnitTypeMain)
	}
	if got.Seat != "서울특별시" {
		t.Errorf("Seat = %q, want %q", got.Seat, "서울특별시")
	}
	if len(got.Jurisdictions) != 1 || got.Jurisdictions[0] != "서울특별시" {
		t.Errorf("Jurisdictions = %v, want [서울특별시]", got.Jurisdictions)
	}
}

func TestCourtDirectory(t *testing.T) {
	courts := Courts()
	if len(courts) != 221 {
		t.Fatalf("Courts() returned %d courts, want 221", len(courts))
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
	if localCourtCount != 107 {
		t.Errorf("found %d municipal/county courts, want 107", localCourtCount)
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
			if panel.Type != CourtPanelTypeAtSeat && panel.Type != CourtPanelTypeOutOfSeat {
				t.Errorf("%s has panel %q with invalid type %q", courtID, panel.Name, panel.Type)
			}
			if panel.Type == CourtPanelTypeAtSeat && panel.Seat != parentSeat {
				t.Errorf("%s panel %q is at %q, want parent seat %q", courtID, panel.Name, panel.Seat, parentSeat)
			}
			if len(panel.Panels) != 0 {
				t.Errorf("%s panel %q unexpectedly contains child panels; panels should belong directly to their court", courtID, panel.Name)
			}
			panelCount++
			checkPanels(courtID, panel.Seat, panel.Panels)
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

func findCourtPanel(panels []CourtPanel, name string) *CourtPanel {
	for i := range panels {
		if panels[i].Name == name {
			return &panels[i]
		}
		if found := findCourtPanel(panels[i].Panels, name); found != nil {
			return found
		}
	}
	return nil
}

func containsCourtPanel(panels []CourtPanel, name string) bool {
	return findCourtPanel(panels, name) != nil
}

func TestInstitutionFields(t *testing.T) {
	got := Institution{
		ID:      "court-001",
		Name:    "Seoul Central District Court",
		Type:    InstitutionTypeCourt,
		Address: "서울특별시 서초구",
		Phone:   "02-0000-0000",
		Email:   "contact@example.com",
	}

	if got.ID != "court-001" {
		t.Errorf("ID = %q, want %q", got.ID, "court-001")
	}
	if got.Name != "Seoul Central District Court" {
		t.Errorf("Name = %q, want %q", got.Name, "Seoul Central District Court")
	}
	if got.Type != InstitutionTypeCourt {
		t.Errorf("Type = %q, want %q", got.Type, InstitutionTypeCourt)
	}
	if got.Address != "서울특별시 서초구" {
		t.Errorf("Address = %q, want %q", got.Address, "서울특별시 서초구")
	}
	if got.Phone != "02-0000-0000" {
		t.Errorf("Phone = %q, want %q", got.Phone, "02-0000-0000")
	}
	if got.Email != "contact@example.com" {
		t.Errorf("Email = %q, want %q", got.Email, "contact@example.com")
	}
}
