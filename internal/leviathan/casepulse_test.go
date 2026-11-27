package leviathan

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

// TestInstitutionFields는 Institution 구조체의 각 필드에 값을 넣고 그대로 읽을 수 있는지 확인합니다.
func TestInstitutionFields(t *testing.T) {
	got := Institution{
		ID:      "prosecution-001",
		Name:    "Seoul Regional Prosecution Office",
		Type:    InstitutionTypeProsecution,
		Address: "서울특별시 서초구",
		Phone:   "02-0000-0000",
		Email:   "contact@example.com",
	}

	if got.ID != "prosecution-001" {
		t.Errorf("ID = %q, want %q", got.ID, "prosecution-001")
	}
	if got.Name != "Seoul Regional Prosecution Office" {
		t.Errorf("Name = %q, want %q", got.Name, "Seoul Regional Prosecution Office")
	}
	if got.Type != InstitutionTypeProsecution {
		t.Errorf("Type = %q, want %q", got.Type, InstitutionTypeProsecution)
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
