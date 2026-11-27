package clio_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mkkim-gr8/casepulse/internal/clio"
	"github.com/mkkim-gr8/casepulse/internal/clio/repotest"
)

func TestMemoryRepository(t *testing.T) {
	repotest.Run(t, func(*testing.T) clio.Repository { return clio.NewMemoryRepository() })
}

var testCourt = clio.Court{
	ID: "court-a", Name: "테스트법원", Seat: "서울",
	Panels: []clio.CourtPanel{{ID: "panel-a", Name: "민사부"}},
}

func validCase() clio.Case {
	return clio.Case{
		Number:   clio.CaseNumber{Year: 2026, Code: "가합", Serial: 1},
		Title:    "손해배상",
		Category: clio.CaseCategoryCivil,
		Parties:  []clio.Party{{Role: "원고", Name: "홍길동", Client: true}, {Role: "피고", Name: "김철수"}},
		Court:    clio.CourtAssignment{CourtID: "court-a", PanelName: "민사부"},
	}
}

func TestStoreCreateAndLookup(t *testing.T) {
	ctx := context.Background()
	store := clio.NewStore(clio.NewMemoryRepository(), []clio.Court{testCourt})
	input := validCase()
	input.Title = "  손해배상  "

	created, err := store.Create(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == "" || created.Court.Name != testCourt.Name || created.Case.Title != "손해배상" {
		t.Fatalf("created record = %+v", created)
	}
	got, err := store.Get(ctx, created.ID)
	if err != nil || got.Case.Number != input.Number || got.Court.Name != testCourt.Name {
		t.Fatalf("Get() = %+v, %v", got, err)
	}
	records, err := store.List(ctx)
	if err != nil || len(records) != 1 {
		t.Fatalf("List() = %d records, %v; want 1", len(records), err)
	}
	if _, err := store.Create(ctx, input); !errors.Is(err, clio.ErrDuplicateCase) {
		t.Fatalf("duplicate Create() error = %v, want ErrDuplicateCase", err)
	}
	if _, err := store.Get(ctx, "missing"); !errors.Is(err, clio.ErrCaseNotFound) {
		t.Fatalf("missing Get() error = %v, want ErrCaseNotFound", err)
	}
}

func TestStoreCreateValidatesCourtAndPanel(t *testing.T) {
	ctx := context.Background()
	store := clio.NewStore(clio.NewMemoryRepository(), []clio.Court{testCourt})
	var validation *clio.ValidationError

	input := validCase()
	input.Court = clio.CourtAssignment{CourtID: "missing", PanelName: "민사부"}
	if _, err := store.Create(ctx, input); !errors.As(err, &validation) || !strings.Contains(err.Error(), "법원 목록에 없습니다") {
		t.Fatalf("unknown court error = %v", err)
	}
	input.Court = clio.CourtAssignment{CourtID: "court-a", PanelName: "형사부"}
	if _, err := store.Create(ctx, input); !errors.As(err, &validation) || !strings.Contains(err.Error(), "소속이 아닙니다") {
		t.Fatalf("unknown panel error = %v", err)
	}
	input = validCase()
	input.Parties = nil
	if _, err := store.Create(ctx, input); !errors.As(err, &validation) {
		t.Fatalf("missing parties error = %v, want ValidationError", err)
	}
}

func TestStoreSeedIfEmpty(t *testing.T) {
	ctx := context.Background()
	store := clio.NewStore(clio.NewMemoryRepository(), []clio.Court{testCourt})
	first, second := validCase(), validCase()
	second.Number.Serial = 2

	seeded, err := store.SeedIfEmpty(ctx, []clio.Case{first, second})
	if err != nil || seeded != 2 {
		t.Fatalf("first SeedIfEmpty() = %d, %v; want 2", seeded, err)
	}
	third := validCase()
	third.Number.Serial = 3
	seeded, err = store.SeedIfEmpty(ctx, []clio.Case{third})
	if err != nil || seeded != 0 {
		t.Fatalf("second SeedIfEmpty() = %d, %v; want 0", seeded, err)
	}

	bad := validCase()
	bad.Court.CourtID = "missing"
	empty := clio.NewStore(clio.NewMemoryRepository(), []clio.Court{testCourt})
	if _, err := empty.SeedIfEmpty(ctx, []clio.Case{validCase(), bad}); err == nil {
		t.Fatal("SeedIfEmpty() with invalid case succeeded")
	}
	if records, _ := empty.List(ctx); len(records) != 0 {
		t.Fatalf("invalid seed stored %d cases, want 0", len(records))
	}
}
