package clio

// 이 파일은 소송사건 한 건의 기본 정보(사건번호, 사건명, 당사자, 계속 법원·재판부)를 담는 자료형과
// 그 값을 검사하는 함수를 담습니다. 진행내역·송달·문서 등은 이 서비스의 범위가 아니므로 넣지 않습니다.

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// CaseCategory는 사건이 어느 종류의 재판인지 나타냅니다.
type CaseCategory string

const (
	// CaseCategoryCivil은 민사사건입니다(예: 가합, 가단, 가소, 나, 다).
	CaseCategoryCivil CaseCategory = "civil"
	// CaseCategoryAdministrative는 행정사건입니다(예: 구합, 구단, 누, 두).
	CaseCategoryAdministrative CaseCategory = "administrative"
	// CaseCategoryFamily는 가사사건입니다(예: 드합, 드단, 느합, 느단, 르, 므).
	CaseCategoryFamily CaseCategory = "family"
	// CaseCategoryCivilApplication은 민사신청사건입니다(예: 카합, 카단, 카기, 카담).
	CaseCategoryCivilApplication CaseCategory = "civil_application"
	// CaseCategoryCriminal은 형사사건입니다(예: 고합, 고단, 고정, 노, 도).
	CaseCategoryCriminal CaseCategory = "criminal"
)

// valid는 c가 위에 정의한 사건 종류 중 하나인지 확인합니다.
func (c CaseCategory) valid() bool {
	switch c {
	case CaseCategoryCivil, CaseCategoryAdministrative, CaseCategoryFamily,
		CaseCategoryCivilApplication, CaseCategoryCriminal:
		return true
	}
	return false
}

// CaseNumber는 법원 사건번호입니다. 연도 + 사건부호 + 일련번호로 이루어집니다.
// 예: "2026가합12345"는 Year 2026, Code "가합", Serial 12345입니다.
//
// 사건번호는 법원마다 따로 매기므로, 같은 사건번호라도 법원이 다르면 다른 사건입니다.
type CaseNumber struct {
	// Year는 4자리 접수 연도입니다.
	Year int
	// Code는 사건부호입니다(예: "가합", "카단", "고단", "노"). 한글로만 이루어집니다.
	Code string
	// Serial은 그 해 그 사건부호 안에서의 일련번호이며 1 이상입니다.
	Serial int
}

// String은 사건번호를 "2026가합12345" 형식으로 돌려줍니다.
func (n CaseNumber) String() string {
	return fmt.Sprintf("%04d%s%d", n.Year, n.Code, n.Serial)
}

// Validate는 연도·사건부호·일련번호가 형식에 맞는지 검사합니다.
func (n CaseNumber) Validate() error {
	var errs []error
	if n.Year < 1000 || n.Year > 9999 {
		errs = append(errs, fmt.Errorf("사건번호 연도 %d는 4자리여야 합니다", n.Year))
	}
	if n.Code == "" {
		errs = append(errs, errors.New("사건부호가 비어 있습니다"))
	}
	for _, r := range n.Code {
		if !unicode.Is(unicode.Hangul, r) {
			errs = append(errs, fmt.Errorf("사건부호 %q에는 한글만 쓸 수 있습니다", n.Code))
			break
		}
	}
	if n.Serial < 1 {
		errs = append(errs, fmt.Errorf("사건번호 일련번호 %d는 1 이상이어야 합니다", n.Serial))
	}
	return errors.Join(errs...)
}

// ParseCaseNumber는 "2026가합12345" 형식의 문자열을 CaseNumber로 바꿉니다.
// 앞뒤 공백과 사이 공백("2026 가합 12345")은 무시합니다.
func ParseCaseNumber(s string) (CaseNumber, error) {
	compact := strings.Join(strings.Fields(s), "")
	runes := []rune(compact)

	// 앞쪽 숫자 = 연도, 가운데 숫자 아닌 부분 = 사건부호, 뒤쪽 숫자 = 일련번호.
	i := 0
	for i < len(runes) && unicode.IsDigit(runes[i]) {
		i++
	}
	j := len(runes)
	for j > i && unicode.IsDigit(runes[j-1]) {
		j--
	}
	if i != 4 || j == i || j == len(runes) {
		return CaseNumber{}, fmt.Errorf("사건번호 %q는 연도(4자리)+사건부호+일련번호 형식이 아닙니다", s)
	}

	year, err := strconv.Atoi(string(runes[:i]))
	if err != nil {
		return CaseNumber{}, fmt.Errorf("사건번호 %q의 연도를 읽을 수 없습니다: %w", s, err)
	}
	serial, err := strconv.Atoi(string(runes[j:]))
	if err != nil {
		return CaseNumber{}, fmt.Errorf("사건번호 %q의 일련번호를 읽을 수 없습니다: %w", s, err)
	}
	n := CaseNumber{Year: year, Code: string(runes[i:j]), Serial: serial}
	if err := n.Validate(); err != nil {
		return CaseNumber{}, err
	}
	return n, nil
}

// Party는 사건의 당사자 한 명(또는 한 기관)입니다.
//
// 지위와 이름은 모두 사용자가 직접 입력합니다. 지위는 정해진 값 목록이 없으며,
// 민사 등에서는 "원고", "피고", "채권자", "신청인", "원고(반소피고)" 등을,
// 형사사건에서는 "피고인", "검사" 등을 그대로 적습니다.
//
// 의뢰인 여부는 당사자마다 따로 표시합니다. 같은 지위의 당사자가 여럿이어도
// 그중 일부만 의뢰인일 수 있기 때문입니다(예: 원고 3명 중 원고1·원고3만 의뢰인).
type Party struct {
	// Role은 당사자 지위입니다(예: "원고", "피고", "피고인", "검사").
	Role string
	// Name은 당사자 이름입니다(예: "홍길동", "주식회사 가나다", "서울중앙지방공소청 검사 김OO").
	Name string
	// Client는 이 당사자가 우리 사무소의 의뢰인인지 나타냅니다.
	// 한 사건에는 의뢰인인 당사자가 반드시 한 명 이상 있어야 하고, 의뢰인은 모두 같은 지위여야 합니다
	// (지위가 다른 당사자를 함께 대리하면 이해충돌이 생기기 때문입니다. PartySide 참고).
	Client bool
}

// PartySide는 지위를 같은 쪽 당사자끼리 비교할 수 있는 값으로 바꿉니다.
// 괄호 안의 표기, 끝의 번호, 띄어쓰기를 지웁니다.
//
//	"원고1", "원고 2"        → "원고"
//	"피고(반소원고)"         → "피고"
//	"피고인3"               → "피고인"
//
// 의뢰인들이 모두 같은 쪽인지 검사할 때 씁니다(Case.Validate).
func PartySide(role string) string {
	var b strings.Builder
	depth := 0
	for _, r := range role {
		switch {
		case r == '(' || r == '（':
			depth++
		case r == ')' || r == '）':
			if depth > 0 {
				depth--
			}
		case depth == 0 && !unicode.IsSpace(r):
			b.WriteRune(r)
		}
	}
	return strings.TrimRight(b.String(), "0123456789")
}

// CourtAssignment는 Leviathan gRPC 서비스의 법원 ID와 그 법원에 속한 재판부 이름을 가리킵니다.
type CourtAssignment struct {
	// CourtID는 Leviathan 서비스가 발급한 법원 ID입니다(예: "court-district-seoul-central").
	CourtID string
	// PanelName은 해당 법원의 재판부 이름입니다(예: "제91민사단독").
	// 재판부 ID는 재판부구성표가 바뀌면 다시 매겨질 수 있으므로, 사건에는 ID 대신 이름을 저장합니다.
	PanelName string
}

// Case는 소송사건 한 건의 기본 정보입니다.
type Case struct {
	// Number는 사건번호입니다.
	Number CaseNumber
	// Title은 사건명이며 사용자가 직접 입력합니다(예: "손해배상(기)", "사기").
	Title string
	// Category는 사건 종류입니다.
	Category CaseCategory
	// Parties는 당사자 목록이며 1명 이상이고, 그중 의뢰인(Client)이 1명 이상 있어야 합니다.
	// 사용자가 입력한 순서를 그대로 유지합니다.
	// 보통 원고·피고(형사는 피고인·검사) 두 명이지만, 1명이거나 3명 이상일 수도 있습니다.
	Parties []Party
	// Court는 사건이 계속 중인 법원과 재판부이며, 모든 종류의 사건에 필수입니다.
	Court CourtAssignment
}

// Clients는 의뢰인으로 표시된 당사자를 입력 순서대로 돌려줍니다.
func (c Case) Clients() []Party {
	var clients []Party
	for _, p := range c.Parties {
		if p.Client {
			clients = append(clients, p)
		}
	}
	return clients
}

// clientSides는 의뢰인들의 지위(PartySide)를 처음 나온 순서대로 중복 없이 돌려줍니다.
func clientSides(clients []Party) []string {
	var sides []string
	seen := make(map[string]bool)
	for _, p := range clients {
		side := PartySide(p.Role)
		if side != "" && !seen[side] {
			seen[side] = true
			sides = append(sides, side)
		}
	}
	return sides
}

// Validate는 사건 정보가 빠짐없이 올바르게 들어 있는지 검사하고, 문제를 모두 모아 돌려줍니다.
//
//   - 사건번호 형식, 사건명, 사건 종류
//   - 당사자 1명 이상, 각 당사자의 지위와 이름, 의뢰인으로 표시된 당사자 1명 이상
//   - 의뢰인이 모두 같은 지위인지(이해충돌 방지)
//   - 법원 ID와 재판부 이름이 입력되었는지
//
// 법원 존재 여부와 재판부 소속 여부는 Clio 저장소가 시작할 때 Leviathan에서 받은 목록으로 확인합니다.
func (c Case) Validate() error {
	var errs []error
	if err := c.ValidateFields(); err != nil {
		errs = append(errs, err)
	}
	if err := c.Court.validate(); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// ValidateFields는 사건 자체의 필드를 검사합니다. 법원과 재판부의 존재 여부는 검사하지 않으므로,
// Clio 서비스 계층은 이 검사가 끝난 뒤 법원 목록에서 실제 법원과 재판부를 확인합니다.
func (c Case) ValidateFields() error {
	var errs []error
	if err := c.Number.Validate(); err != nil {
		errs = append(errs, err)
	}
	if strings.TrimSpace(c.Title) == "" {
		errs = append(errs, errors.New("사건명이 비어 있습니다"))
	}
	if !c.Category.valid() {
		errs = append(errs, fmt.Errorf("알 수 없는 사건 종류 %q입니다", c.Category))
	}

	if len(c.Parties) == 0 {
		errs = append(errs, errors.New("당사자가 한 명 이상 있어야 합니다"))
	} else if clients := c.Clients(); len(clients) == 0 {
		errs = append(errs, errors.New("의뢰인으로 표시된 당사자가 한 명 이상 있어야 합니다"))
	} else if sides := clientSides(clients); len(sides) > 1 {
		errs = append(errs, fmt.Errorf("지위가 다른 당사자(%s)를 함께 의뢰인으로 할 수 없습니다(이해충돌)", strings.Join(sides, ", ")))
	}
	for i, p := range c.Parties {
		if strings.TrimSpace(p.Role) == "" {
			errs = append(errs, fmt.Errorf("%d번째 당사자의 지위가 비어 있습니다", i+1))
		}
		if strings.TrimSpace(p.Name) == "" {
			errs = append(errs, fmt.Errorf("%d번째 당사자의 이름이 비어 있습니다", i+1))
		}
	}

	return errors.Join(errs...)
}

// validate는 법원 목록을 찾아보기 전에 법원 ID와 재판부 이름이 입력되었는지 확인합니다.
func (a CourtAssignment) validate() error {
	if strings.TrimSpace(a.CourtID) == "" {
		return errors.New("계속 법원이 비어 있습니다")
	}
	if strings.TrimSpace(a.PanelName) == "" {
		return errors.New("재판부가 비어 있습니다")
	}
	return nil
}
