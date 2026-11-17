package casepulse

import (
	"fmt"
	"strings"
	"testing"
)

func TestBuildCourtDirectoryPrintsTree(t *testing.T) {
	courts := buildCourtDirectory()
	childrenByParent := make(map[string][]Court)
	var roots []Court

	for _, court := range courts {
		if court.ParentCourtID == "" {
			roots = append(roots, court)
			continue
		}
		childrenByParent[court.ParentCourtID] = append(childrenByParent[court.ParentCourtID], court)
	}

	if len(roots) != 1 {
		t.Fatalf("최상위 법원이 %d곳입니다. 1곳이어야 합니다.", len(roots))
	}

	var lines []string
	visited := make(map[string]bool, len(courts))
	var addPanel func(panel CourtPanel, prefix string, isLast bool)
	addPanel = func(panel CourtPanel, prefix string, isLast bool) {
		connector := "├── "
		if isLast {
			connector = "└── "
		}
		lines = append(lines, fmt.Sprintf("%s%s재판부: %s [%s] 소재지: %s",
			prefix, connector, panel.Name, panel.Type, panel.Seat))

		childPrefix := prefix + "│   "
		if isLast {
			childPrefix = prefix + "    "
		}
		for i, child := range panel.Panels {
			addPanel(child, childPrefix, i == len(panel.Panels)-1)
		}
	}

	var addCourt func(court Court, prefix string, isLast, isRoot bool)
	addCourt = func(court Court, prefix string, isLast, isRoot bool) {
		if visited[court.ID] {
			t.Fatalf("법원 %q가 트리에 중복되었거나 순환 참조가 있습니다.", court.ID)
		}
		visited[court.ID] = true

		connector := ""
		if !isRoot {
			connector = "├── "
			if isLast {
				connector = "└── "
			}
		}

		details := fmt.Sprintf("%s / %s / %s", court.Level, court.CourtType, court.UnitType)
		line := fmt.Sprintf("%s%s%s [%s] 소재지: %s", prefix, connector, court.Name, details, court.Seat)
		if len(court.Jurisdictions) > 0 {
			line += " / 관할: " + strings.Join(court.Jurisdictions, ", ")
		}
		lines = append(lines, line)

		children := childrenByParent[court.ID]
		childPrefix := prefix
		if !isRoot {
			if isLast {
				childPrefix += "    "
			} else {
				childPrefix += "│   "
			}
		}
		totalChildren := len(court.Panels) + len(children)
		for i, panel := range court.Panels {
			addPanel(panel, childPrefix, i+len(children) == totalChildren-1)
		}
		for i, child := range children {
			addCourt(child, childPrefix, len(court.Panels)+i == totalChildren-1, false)
		}
	}

	for i, root := range roots {
		addCourt(root, "", i == len(roots)-1, true)
	}

	if len(visited) != len(courts) {
		t.Fatalf("트리에서 %d곳 중 %d곳만 출력했습니다.", len(visited), len(courts))
	}

	t.Logf("전국 법원 구조 (%d곳)\n%s", len(courts), strings.Join(lines, "\n"))
}
