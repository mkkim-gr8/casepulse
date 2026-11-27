package leviathan_test

import (
	"testing"

	"github.com/mkkim-gr8/casepulse/internal/leviathan"
)

// TestInstitutionPublicAPI는 패키지 밖(casepulse_test 패키지)에서 공개 자료형과 상수를 쓸 수 있는지 확인합니다.
func TestInstitutionPublicAPI(t *testing.T) {
	institution := leviathan.Institution{
		ID:   "police-001",
		Name: "Example Police Station",
		Type: leviathan.InstitutionTypePolice,
	}

	if institution.Type != leviathan.InstitutionTypePolice {
		t.Fatalf("Type = %q, want %q", institution.Type, leviathan.InstitutionTypePolice)
	}
}
