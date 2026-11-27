package clio

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestStoreCreateAndLookup(t *testing.T) {
	court := Court{
		ID: "court-a", Name: "테스트법원", Seat: "서울",
		Panels: []CourtPanel{{ID: "panel-a", Name: "민사부"}},
	}
	store := NewStore(nil, []Court{court})
	input := Case{
		Number:   CaseNumber{Year: 2026, Code: "가합", Serial: 1},
		Title:    "손해배상",
		Category: CaseCategoryCivil,
		Parties:  []Party{{Role: "원고", Name: "홍길동", Client: true}, {Role: "피고", Name: "김철수"}},
		Court:    CourtAssignment{CourtID: "court-a", PanelName: "민사부"},
	}

	created, err := store.Create(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != "case-0001" || created.Court.Name != court.Name {
		t.Fatalf("created record = %+v", created)
	}
	got, err := store.Get(context.Background(), created.ID)
	if err != nil || got.Case.Number != input.Number {
		t.Fatalf("Get() = %+v, %v", got, err)
	}
	records := store.List(context.Background())
	if len(records) != 1 {
		t.Fatalf("List() length = %d, want 1", len(records))
	}
	if _, err := store.Create(context.Background(), input); !errors.Is(err, ErrDuplicateCase) {
		t.Fatalf("duplicate Create() error = %v, want ErrDuplicateCase", err)
	}
	if _, err := store.Get(context.Background(), "missing"); !errors.Is(err, ErrCaseNotFound) {
		t.Fatalf("missing Get() error = %v, want ErrCaseNotFound", err)
	}
}

func TestStoreCreateValidatesCourtAndPanel(t *testing.T) {
	court := Court{ID: "court-a", Name: "테스트법원", Panels: []CourtPanel{{Name: "민사부"}}}
	store := NewStore(nil, []Court{court})
	input := Case{
		Number: CaseNumber{Year: 2026, Code: "가합", Serial: 1}, Title: "손해배상",
		Category: CaseCategoryCivil, Parties: []Party{{Role: "원고", Name: "홍길동", Client: true}},
		Court: CourtAssignment{CourtID: "missing", PanelName: "민사부"},
	}
	if _, err := store.Create(context.Background(), input); err == nil || !strings.Contains(err.Error(), "법원 목록에 없습니다") {
		t.Fatalf("unknown court error = %v", err)
	}
	input.Court = CourtAssignment{CourtID: "court-a", PanelName: "형사부"}
	if _, err := store.Create(context.Background(), input); err == nil || !strings.Contains(err.Error(), "소속이 아닙니다") {
		t.Fatalf("unknown panel error = %v", err)
	}
}
