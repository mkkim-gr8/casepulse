package clio

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
)

var (
	// ErrDuplicateCase는 같은 법원에 같은 사건번호가 이미 저장되어 있을 때 반환합니다.
	ErrDuplicateCase = errors.New("같은 법원에 같은 사건번호가 이미 등록되어 있습니다")
	// ErrCaseNotFound는 요청한 사건 ID가 저장소에 없을 때 반환합니다.
	ErrCaseNotFound = errors.New("사건을 찾을 수 없습니다")
)

// Record는 사건과 화면에 보여 줄 법원 정보를 함께 묶어 저장합니다.
type Record struct {
	ID    string
	Case  Case
	Court Court
}

// Store는 Clio 프로세스 안에서 사건을 임시로 보관하는 메모리 저장소입니다.
// 현재 DB가 없으므로 프로세스를 재시작하면 여기에 추가한 사건은 사라집니다.
type Store struct {
	mu     sync.RWMutex      // 여러 gRPC 요청이 동시에 접근할 때 자료를 보호합니다.
	courts map[string]Court  // 시작할 때 받은 법원 목록을 ID로 빠르게 찾습니다.
	list   []Record          // 목록 API가 반환할 사건 순서를 보관합니다.
	byID   map[string]Record // 사건 ID를 이용해 한 건을 빠르게 찾습니다.
	keys   map[string]bool   // 법원 ID와 사건번호 조합의 중복 여부를 기억합니다.
	nextID int               // 새 사건에 순서대로 붙일 임시 ID의 번호입니다.
}

// NewStore는 더미 사건과 법원 목록으로 메모리 저장소를 준비합니다.
// 초기 사건이 목록에 없는 법원을 가리키면 데이터 오류를 바로 알 수 있도록 패닉을 일으킵니다.
func NewStore(cases []Case, courts []Court) *Store {
	s := &Store{
		courts: make(map[string]Court, len(courts)),
		byID:   make(map[string]Record, len(cases)),
		keys:   make(map[string]bool, len(cases)),
	}
	for _, court := range courts {
		s.courts[court.ID] = cloneCourt(court)
	}
	for _, c := range cases {
		court, ok := s.courts[c.Court.CourtID]
		if !ok {
			panic(fmt.Sprintf("사건 %s의 법원 ID %q가 법원 목록에 없습니다", c.Number, c.Court.CourtID))
		}
		s.add(c, court)
	}
	sort.SliceStable(s.list, func(i, j int) bool {
		if s.list[i].Case.Number.Year != s.list[j].Case.Number.Year {
			return s.list[i].Case.Number.Year > s.list[j].Case.Number.Year
		}
		return s.list[i].Case.Number.Serial > s.list[j].Case.Number.Serial
	})
	return s
}

// List는 최신 사건부터 복사해 돌려줍니다.
// 복사본을 반환하므로 호출자가 결과를 바꿔도 저장소 안의 사건은 바뀌지 않습니다.
func (s *Store) List(context.Context) []Record {
	s.mu.RLock()
	defer s.mu.RUnlock()
	records := make([]Record, len(s.list))
	for i, record := range s.list {
		records[i] = cloneRecord(record)
	}
	return records
}

// Get은 사건 ID로 사건 한 건을 찾습니다.
func (s *Store) Get(_ context.Context, id string) (Record, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.byID[id]
	if !ok {
		return Record{}, ErrCaseNotFound
	}
	return cloneRecord(record), nil
}

// Create는 새 사건을 검사한 뒤 저장합니다.
// 사건의 법원과 재판부가 Clio 시작 때 받은 Leviathan 목록에 있는지도 확인합니다.
func (s *Store) Create(ctx context.Context, c Case) (Record, error) {
	if err := ctx.Err(); err != nil {
		return Record{}, err
	}
	c.Number.Code = strings.TrimSpace(c.Number.Code)
	c.Title = strings.TrimSpace(c.Title)
	c.Court.CourtID = strings.TrimSpace(c.Court.CourtID)
	c.Court.PanelName = strings.TrimSpace(c.Court.PanelName)
	for i := range c.Parties {
		c.Parties[i].Role = strings.TrimSpace(c.Parties[i].Role)
		c.Parties[i].Name = strings.TrimSpace(c.Parties[i].Name)
	}
	if err := c.ValidateFields(); err != nil {
		return Record{}, err
	}
	if c.Court.CourtID == "" {
		return Record{}, errors.New("계속 법원이 비어 있습니다")
	}

	court, ok := s.courts[c.Court.CourtID]
	if !ok {
		return Record{}, fmt.Errorf("법원 ID %q는 법원 목록에 없습니다", c.Court.CourtID)
	}
	if c.Court.PanelName == "" {
		return Record{}, fmt.Errorf("%s의 재판부가 비어 있습니다", court.Name)
	}
	foundPanel := false
	for _, panel := range court.Panels {
		if panel.Name == c.Court.PanelName {
			foundPanel = true
			break
		}
	}
	if !foundPanel {
		return Record{}, fmt.Errorf("재판부 %q는 %s 소속이 아닙니다", c.Court.PanelName, court.Name)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return Record{}, err
	}
	if s.keys[caseKey(c)] {
		return Record{}, ErrDuplicateCase
	}
	return cloneRecord(s.add(c, court)), nil
}

// add는 검증이 끝난 사건에 ID를 붙여 저장소 앞에 추가합니다.
// 동시 접근을 막기 위해 호출하는 쪽에서 쓰기 잠금을 잡아야 합니다.
func (s *Store) add(c Case, court Court) Record {
	s.nextID++
	record := Record{
		ID:    fmt.Sprintf("case-%04d", s.nextID),
		Case:  cloneCase(c),
		Court: cloneCourt(court),
	}
	s.list = append([]Record{record}, s.list...)
	s.byID[record.ID] = record
	s.keys[caseKey(c)] = true
	return record
}

// caseKey는 같은 법원 안에서 사건번호가 중복되는지 확인할 때 쓰는 고유 키입니다.
func caseKey(c Case) string {
	return c.Court.CourtID + "/" + c.Number.String()
}

// cloneRecord는 사건의 당사자 목록과 법원의 재판부 목록까지 복사합니다.
func cloneRecord(record Record) Record {
	record.Case = cloneCase(record.Case)
	record.Court = cloneCourt(record.Court)
	return record
}

// cloneCase는 내부 슬라이스인 당사자 목록을 복사합니다.
func cloneCase(c Case) Case {
	c.Parties = append([]Party(nil), c.Parties...)
	return c
}

// cloneCourt는 내부 슬라이스인 재판부 목록을 복사합니다.
func cloneCourt(c Court) Court {
	c.Panels = append([]CourtPanel(nil), c.Panels...)
	return c
}
