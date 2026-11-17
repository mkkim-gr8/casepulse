package casepulse_test

import (
	"testing"

	"github.com/mkkim-gr8/casepulse"
)

func TestInstitutionPublicAPI(t *testing.T) {
	institution := casepulse.Institution{
		ID:   "police-001",
		Name: "Example Police Station",
		Type: casepulse.InstitutionTypePolice,
	}

	if institution.Type != casepulse.InstitutionTypePolice {
		t.Fatalf("Type = %q, want %q", institution.Type, casepulse.InstitutionTypePolice)
	}
}
