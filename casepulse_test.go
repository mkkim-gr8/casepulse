package casepulse

// 이 파일은 도메인 자료형(상수와 구조체)의 기본 동작을 검사합니다.

import "testing"

// TestInstitutionTypes는 기관 종류 상수의 문자열 값이 바뀌지 않았는지 확인합니다.
// 이 값은 외부 시스템에 저장·전송될 수 있으므로 바꾸면 호환성이 깨집니다.
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

// TestCourtTypes는 법원 분류 상수(단계, 분야, 본원·지원 구분, 재판부 유형)의 문자열 값을 확인합니다.
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// TestCourtFields는 Court 구조체에 값을 넣고 읽을 수 있는지, Institution 필드가 포함(embed)되는지 확인합니다.
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

// TestInstitutionFields는 Institution 구조체의 각 필드에 값을 넣고 그대로 읽을 수 있는지 확인합니다.
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
