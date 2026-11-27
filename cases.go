package casepulse

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
type Party struct {
	// Role은 당사자 지위입니다(예: "원고", "피고", "피고인", "검사").
	Role string
	// Name은 당사자 이름입니다(예: "홍길동", "주식회사 가나다", "서울중앙지방공소청 검사 김OO").
	Name string
}

// CourtAssignment는 사건이 계속 중인 법원과 재판부입니다.
type CourtAssignment struct {
	// CourtID는 Courts 목록의 법원 ID입니다(예: "court-district-seoul-central").
	CourtID string
	// PanelName은 그 법원 Panels 중 하나의 재판부 이름입니다(예: "제91민사단독").
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
	// Parties는 당사자 목록이며 1명 이상입니다. 사용자가 입력한 순서를 그대로 유지합니다.
	// 보통 원고·피고(형사는 피고인·검사) 두 명이지만, 1명이거나 3명 이상일 수도 있습니다.
	Parties []Party
	// Court는 사건이 계속 중인 법원과 재판부이며, 모든 종류의 사건에 필수입니다.
	Court CourtAssignment
}

// Validate는 사건 정보가 빠짐없이 올바르게 들어 있는지 검사하고, 문제를 모두 모아 돌려줍니다.
//
//   - 사건번호 형식, 사건명, 사건 종류
//   - 당사자 1명 이상, 각 당사자의 지위와 이름
//   - 계속 법원·재판부: Courts 목록에 있는 법원인지, 재판부가 그 법원 소속인지
func (c Case) Validate() error {
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
	}
	for i, p := range c.Parties {
		if strings.TrimSpace(p.Role) == "" {
			errs = append(errs, fmt.Errorf("%d번째 당사자의 지위가 비어 있습니다", i+1))
		}
		if strings.TrimSpace(p.Name) == "" {
			errs = append(errs, fmt.Errorf("%d번째 당사자의 이름이 비어 있습니다", i+1))
		}
	}

	if err := c.Court.validate(); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// validate는 법원 ID가 Courts 목록에 있고 재판부가 그 법원에 소속되어 있는지 확인합니다.
func (a CourtAssignment) validate() error {
	if strings.TrimSpace(a.CourtID) == "" {
		return errors.New("계속 법원이 비어 있습니다")
	}
	court, ok := courtByID(a.CourtID)
	if !ok {
		return fmt.Errorf("법원 ID %q는 법원 목록에 없습니다", a.CourtID)
	}
	if strings.TrimSpace(a.PanelName) == "" {
		return fmt.Errorf("%s의 재판부가 비어 있습니다", court.Name)
	}
	for _, panel := range court.Panels {
		if panel.Name == a.PanelName {
			return nil
		}
	}
	return fmt.Errorf("재판부 %q는 %s 소속이 아닙니다", a.PanelName, court.Name)
}

// courtByID는 내부 법원 원본 목록에서 ID로 법원을 찾습니다. 읽기 전용으로만 써야 합니다.
func courtByID(id string) (*Court, bool) {
	for i := range courtDirectory {
		if courtDirectory[i].ID == id {
			return &courtDirectory[i], true
		}
	}
	return nil, false
}
