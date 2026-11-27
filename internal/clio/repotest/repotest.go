// Package repotest는 clio.Repository 구현(메모리, PostgreSQL)이 같은 방식으로 동작하는지 확인하는
// 공통 테스트를 제공합니다.
package repotest

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/mkkim-gr8/casepulse/internal/clio"
)

// Run은 newRepo가 돌려주는 빈 저장소마다 공통 테스트를 실행합니다.
func Run(t *testing.T, newRepo func(t *testing.T) clio.Repository) {
	t.Run("InsertGetList", func(t *testing.T) { testInsertGetList(t, newRepo(t)) })
	t.Run("Duplicate", func(t *testing.T) { testDuplicate(t, newRepo(t)) })
	t.Run("BatchIsAtomic", func(t *testing.T) { testBatchIsAtomic(t, newRepo(t)) })
	t.Run("ListOrder", func(t *testing.T) { testListOrder(t, newRepo(t)) })
	t.Run("NotFound", func(t *testing.T) { testNotFound(t, newRepo(t)) })
}

// sampleCase는 법원 court-a의 serial번 민사사건을 만듭니다.
func sampleCase(serial int) clio.Case {
	return clio.Case{
		Number:   clio.CaseNumber{Year: 2026, Code: "가합", Serial: serial},
		Title:    "손해배상",
		Category: clio.CaseCategoryCivil,
		Parties: []clio.Party{
			{Role: "원고1", Name: "홍길동", Client: true},
			{Role: "원고2", Name: "주식회사 가나다"},
			{Role: "피고", Name: "김철수"},
		},
		Court: clio.CourtAssignment{CourtID: "court-a", PanelName: "민사부"},
	}
}

func testInsertGetList(t *testing.T, repo clio.Repository) {
	ctx := context.Background()
	input := sampleCase(1)
	stored, err := repo.Insert(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	if len(stored) != 1 || stored[0].ID == "" || !reflect.DeepEqual(stored[0].Case, input) {
		t.Fatalf("Insert() = %+v", stored)
	}

	got, err := repo.Get(ctx, stored[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, stored[0]) {
		t.Fatalf("Get() = %+v, want %+v", got, stored[0])
	}

	list, err := repo.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || !reflect.DeepEqual(list[0], stored[0]) {
		t.Fatalf("List() = %+v", list)
	}
	if count, err := repo.Count(ctx); err != nil || count != 1 {
		t.Fatalf("Count() = %d, %v", count, err)
	}

	// 돌려받은 값을 바꿔도 저장된 사건은 바뀌지 않아야 합니다.
	got.Case.Parties[0].Name = "변경"
	again, err := repo.Get(ctx, stored[0].ID)
	if err != nil || again.Case.Parties[0].Name != "홍길동" {
		t.Fatalf("stored case changed through returned value: %+v, %v", again, err)
	}
}

func testDuplicate(t *testing.T, repo clio.Repository) {
	ctx := context.Background()
	if _, err := repo.Insert(ctx, sampleCase(1)); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Insert(ctx, sampleCase(1)); !errors.Is(err, clio.ErrDuplicateCase) {
		t.Fatalf("duplicate Insert() error = %v, want ErrDuplicateCase", err)
	}
	// 법원이 다르면 같은 사건번호라도 다른 사건입니다.
	other := sampleCase(1)
	other.Court.CourtID = "court-b"
	if _, err := repo.Insert(ctx, other); err != nil {
		t.Fatalf("same number in another court: %v", err)
	}
}

func testBatchIsAtomic(t *testing.T, repo clio.Repository) {
	ctx := context.Background()
	if _, err := repo.Insert(ctx, sampleCase(1), sampleCase(2), sampleCase(1)); !errors.Is(err, clio.ErrDuplicateCase) {
		t.Fatalf("batch with duplicate error = %v, want ErrDuplicateCase", err)
	}
	if count, err := repo.Count(ctx); err != nil || count != 0 {
		t.Fatalf("Count() after failed batch = %d, %v; want 0", count, err)
	}
}

func testListOrder(t *testing.T, repo clio.Repository) {
	ctx := context.Background()
	// 한 번에 넣은 사건은 사건번호가 큰 순서, 나중에 넣은 사건은 그보다 앞에 옵니다.
	if _, err := repo.Insert(ctx, sampleCase(3), sampleCase(7), sampleCase(5)); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Insert(ctx, sampleCase(1)); err != nil {
		t.Fatal(err)
	}
	list, err := repo.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var serials []int
	for _, item := range list {
		serials = append(serials, item.Case.Number.Serial)
	}
	if want := []int{1, 7, 5, 3}; !reflect.DeepEqual(serials, want) {
		t.Fatalf("List() serials = %v, want %v", serials, want)
	}
}

func testNotFound(t *testing.T, repo clio.Repository) {
	ctx := context.Background()
	for _, id := range []string{"missing", "00000000-0000-0000-0000-000000000000"} {
		if _, err := repo.Get(ctx, id); !errors.Is(err, clio.ErrCaseNotFound) {
			t.Errorf("Get(%q) error = %v, want ErrCaseNotFound", id, err)
		}
	}
}
