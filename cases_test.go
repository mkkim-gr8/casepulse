package casepulse

// 이 파일은 사건번호 파싱과 사건 정보 검사(cases.go)를 검사합니다.

import (
	"strings"
	"testing"
)

func TestParseCaseNumber(t *testing.T) {
	tests := []struct {
		in   string
		want CaseNumber
	}{
		{"2026가합12345", CaseNumber{2026, "가합", 12345}},
		{" 2025 카단 7 ", CaseNumber{2025, "카단", 7}},
		{"2024노1", CaseNumber{2024, "노", 1}},
	}
	for _, tt := range tests {
		got, err := ParseCaseNumber(tt.in)
		if err != nil {
			t.Errorf("ParseCaseNumber(%q) error: %v", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("ParseCaseNumber(%q) = %+v, want %+v", tt.in, got, tt.want)
		}
	}

	for _, in := range []string{"", "2026", "가합123", "26가합123", "2026가합", "2026ab123", "2026가합0"} {
		if _, err := ParseCaseNumber(in); err == nil {
			t.Errorf("ParseCaseNumber(%q) succeeded, want error", in)
		}
	}

	if s := (CaseNumber{2026, "가합", 12345}).String(); s != "2026가합12345" {
		t.Errorf("String() = %q", s)
	}
}

func TestCaseValidate(t *testing.T) {
	seoulCivil := CourtAssignment{CourtID: "court-district-seoul-central", PanelName: "제91민사단독"}
	num := CaseNumber{2026, "가단", 100}

	valid := []Case{
		{ // 원고·피고 2인
			Number: num, Title: "손해배상(기)", Category: CaseCategoryCivil, Court: seoulCivil,
			Parties: []Party{{"원고", "홍길동"}, {"피고", "주식회사 가나다"}},
		},
		{ // 신청인 1인
			Number: CaseNumber{2026, "카단", 5}, Title: "부동산가압류", Category: CaseCategoryCivilApplication, Court: seoulCivil,
			Parties: []Party{{"채권자", "홍길동"}},
		},
		{ // 3인 이상
			Number: num, Title: "공유물분할", Category: CaseCategoryCivil, Court: seoulCivil,
			Parties: []Party{{"원고", "갑"}, {"피고1", "을"}, {"피고2", "병"}, {"피고3", "정"}},
		},
		{ // 형사: 피고인·검사
			Number: CaseNumber{2026, "고단", 9}, Title: "사기", Category: CaseCategoryCriminal, Court: seoulCivil,
			Parties: []Party{{"피고인", "김철수"}, {"검사", "이영희"}},
		},
	}
	for i, c := range valid {
		if err := c.Validate(); err != nil {
			t.Errorf("valid[%d]: unexpected error: %v", i, err)
		}
	}

	invalid := []struct {
		name    string
		c       Case
		wantMsg string
	}{
		{"no parties", Case{Number: num, Title: "x", Category: CaseCategoryCivil, Court: seoulCivil}, "당사자가 한 명 이상"},
		{"empty role", Case{Number: num, Title: "x", Category: CaseCategoryCivil, Court: seoulCivil, Parties: []Party{{"", "갑"}}}, "지위가 비어"},
		{"empty title", Case{Number: num, Category: CaseCategoryCivil, Court: seoulCivil, Parties: []Party{{"원고", "갑"}}}, "사건명"},
		{"family without court", Case{Number: num, Title: "x", Category: CaseCategoryFamily, Parties: []Party{{"원고", "갑"}}}, "계속 법원이 비어"},
		{"criminal without court", Case{Number: num, Title: "x", Category: CaseCategoryCriminal, Parties: []Party{{"피고인", "갑"}, {"검사", "을"}}}, "계속 법원이 비어"},
		{"criminal without panel", Case{Number: num, Title: "x", Category: CaseCategoryCriminal, Court: CourtAssignment{CourtID: "court-district-seoul-central"}, Parties: []Party{{"피고인", "갑"}}}, "재판부가 비어"},
		{"unknown court", Case{Number: num, Title: "x", Category: CaseCategoryAdministrative, Court: CourtAssignment{"court-nope", "제1부"}, Parties: []Party{{"원고", "갑"}}}, "법원 목록에 없습니다"},
		{"panel of other court", Case{Number: num, Title: "x", Category: CaseCategoryCivil, Court: CourtAssignment{"court-district-incheon", "제91민사단독"}, Parties: []Party{{"원고", "갑"}}}, "소속이 아닙니다"},
		{"bad category", Case{Number: num, Title: "x", Category: "tax", Parties: []Party{{"원고", "갑"}}}, "사건 종류"},
	}
	for _, tt := range invalid {
		err := tt.c.Validate()
		if err == nil || !strings.Contains(err.Error(), tt.wantMsg) {
			t.Errorf("%s: error = %v, want containing %q", tt.name, err, tt.wantMsg)
		}
	}
}
