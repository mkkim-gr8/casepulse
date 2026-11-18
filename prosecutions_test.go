package casepulse

import (
	"fmt"
	"strings"
	"testing"
)

// TestProsecutionDirectory는 공소청 목록의 구조를 검사합니다.
// 기관 수, 단계별 수, ID 중복, 상위 기관의 단계, 지청 이름과 소속, 소재지·주소, 복사본 반환을 확인합니다.
func TestProsecutionDirectory(t *testing.T) {
	offices := ProsecutionOffices()
	// 공소청 1곳 + 별표의 66곳(광역 6, 지방 18, 지청 42)
	if len(offices) != 67 {
		t.Fatalf("ProsecutionOffices() returned %d offices, want 67", len(offices))
	}

	byID := make(map[string]ProsecutionOffice, len(offices))
	counts := make(map[string]int)
	for _, office := range offices {
		if office.ID == "" {
			t.Fatal("office has an empty ID")
		}
		if _, exists := byID[office.ID]; exists {
			t.Fatalf("duplicate office ID %q", office.ID)
		}
		if office.Type != InstitutionTypeProsecution {
			t.Errorf("%s has institution type %q, want %q", office.ID, office.Type, InstitutionTypeProsecution)
		}
		byID[office.ID] = office
		counts[string(office.Level)+"/"+string(office.UnitType)]++
	}

	wantCounts := map[string]int{
		"national/main":   1,
		"regional/main":   6,
		"district/main":   18,
		"district/branch": 42,
	}
	for key, want := range wantCounts {
		if counts[key] != want {
			t.Errorf("%s count = %d, want %d", key, counts[key], want)
		}
	}

	wantParentLevel := map[ProsecutionLevel]ProsecutionLevel{
		ProsecutionLevelRegional: ProsecutionLevelNational,
		ProsecutionLevelDistrict: ProsecutionLevelRegional,
	}
	for _, office := range offices {
		if office.Seat == "" {
			t.Errorf("%s has an empty seat", office.ID)
		}
		if len(office.PostalCode) != 5 || office.Address == "" {
			t.Errorf("%s has postal code %q, address %q", office.ID, office.PostalCode, office.Address)
		}
		if office.ParentOfficeID == "" {
			if office.Level != ProsecutionLevelNational {
				t.Errorf("%s has no parent", office.ID)
			}
			continue
		}
		parent, exists := byID[office.ParentOfficeID]
		if !exists {
			t.Errorf("%s references missing parent %q", office.ID, office.ParentOfficeID)
			continue
		}
		switch office.UnitType {
		case ProsecutionUnitTypeBranch:
			// 지청 이름은 소속 지방공소청 이름으로 시작해야 합니다.
			if parent.UnitType != ProsecutionUnitTypeMain || !strings.HasPrefix(office.Name, parent.Name+" ") {
				t.Errorf("branch %s (%s) has unexpected parent %s (%s)", office.ID, office.Name, parent.ID, parent.Name)
			}
		default:
			if parent.Level != wantParentLevel[office.Level] {
				t.Errorf("%s (%s) parent level = %s, want %s", office.ID, office.Level, parent.Level, wantParentLevel[office.Level])
			}
		}
	}

	if got := byID["prosecution-national"]; got.PostalCode != "06590" || got.Address != "서울특별시 서초구 반포대로 157" {
		t.Errorf("공소청 address = (%q) %q", got.PostalCode, got.Address)
	}
	if got := byID["prosecution-branch-chuncheon-yeongwol"].Seat; got != "영월읍" {
		t.Errorf("영월지청 Seat = %q, want %q", got, "영월읍")
	}
	if got := byID["prosecution-regional-gwangju"].Seat; got != "전남광주통합특별시" {
		t.Errorf("광주광역공소청 Seat = %q, want %q", got, "전남광주통합특별시")
	}

	// 반환값을 수정해도 내부 원본이 바뀌지 않아야 합니다.
	offices[0].Name = "changed"
	if ProsecutionOffices()[0].Name == "changed" {
		t.Error("ProsecutionOffices() exposes the internal directory")
	}
}

// TestBuildProsecutionDirectoryPrintsTree는 ParentOfficeID로 공소청 트리를 만들어, 최상위 기관이 하나뿐이고
// 모든 기관이 순환 없이 한 번씩 나타나는지 확인합니다. go test -v로 실행하면 트리를 출력합니다.
func TestBuildProsecutionDirectoryPrintsTree(t *testing.T) {
	offices := buildProsecutionDirectory()
	children := make(map[string][]ProsecutionOffice)
	var roots []ProsecutionOffice
	for _, office := range offices {
		if office.ParentOfficeID == "" {
			roots = append(roots, office)
			continue
		}
		children[office.ParentOfficeID] = append(children[office.ParentOfficeID], office)
	}
	if len(roots) != 1 {
		t.Fatalf("최상위 공소청이 %d곳입니다. 1곳이어야 합니다.", len(roots))
	}

	var lines []string
	visited := make(map[string]bool, len(offices))
	var walk func(office ProsecutionOffice, prefix string, isLast, isRoot bool)
	walk = func(office ProsecutionOffice, prefix string, isLast, isRoot bool) {
		if visited[office.ID] {
			t.Fatalf("공소청 %q가 트리에 중복되었거나 순환 참조가 있습니다.", office.ID)
		}
		visited[office.ID] = true

		connector, childPrefix := "", prefix
		if !isRoot {
			connector, childPrefix = "├── ", prefix+"│   "
			if isLast {
				connector, childPrefix = "└── ", prefix+"    "
			}
		}
		lines = append(lines, fmt.Sprintf("%s%s%s [%s / %s] 소재지: %s",
			prefix, connector, office.Name, office.Level, office.UnitType, office.Seat))
		kids := children[office.ID]
		for i, child := range kids {
			walk(child, childPrefix, i == len(kids)-1, false)
		}
	}
	walk(roots[0], "", true, true)

	if len(visited) != len(offices) {
		t.Fatalf("트리에서 %d곳 중 %d곳만 출력했습니다.", len(offices), len(visited))
	}
	t.Logf("전국 공소청 구조 (%d곳)\n%s", len(offices), strings.Join(lines, "\n"))
}
