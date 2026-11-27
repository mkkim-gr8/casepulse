package clio

import (
	"testing"

	"github.com/mkkim-gr8/casepulse/internal/leviathan"
)

func TestBuildDummyCases(t *testing.T) {
	source := leviathan.Courts()
	courts := make([]Court, len(source))
	for i, court := range source {
		panels := make([]CourtPanel, len(court.Panels))
		for j, panel := range court.Panels {
			panels[j] = CourtPanel{ID: panel.ID, Name: panel.Name}
		}
		courts[i] = Court{
			ID: court.ID, Name: court.Name, Seat: court.Seat,
			PostalCode: court.PostalCode, Address: court.Address,
			Level: string(court.Level), CourtType: string(court.CourtType), UnitType: string(court.UnitType),
			Panels: panels,
		}
	}

	cases := BuildDummyCases(courts)
	if len(cases) != dummyCaseCount {
		t.Fatalf("got %d cases, want %d", len(cases), dummyCaseCount)
	}
	counts := make(map[CaseCategory]int)
	partialSide := 0
	for _, c := range cases {
		if err := c.Validate(); err != nil {
			t.Fatalf("%s: %v", c.Number, err)
		}
		counts[c.Category]++
		if len(c.Clients()) < len(c.Parties) {
			partialSide++
		}
	}
	if partialSide == 0 {
		t.Error("all parties are marked as clients")
	}
	for _, category := range []CaseCategory{
		CaseCategoryCivil, CaseCategoryCivilApplication, CaseCategoryFamily,
		CaseCategoryAdministrative, CaseCategoryCriminal,
	} {
		if counts[category] == 0 {
			t.Errorf("no dummy cases of category %s", category)
		}
	}
}
