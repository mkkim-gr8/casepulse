package clio

// 이 파일은 Clio의 "서비스 계층"입니다. 구조는 다음과 같습니다.
//
//	gRPC 서버(internal/cliogrpc/server) ──▶ Store(이 파일: 검증 + 법원 정보 붙이기) ──▶ Repository(저장)
//	                                                                                ├─ MemoryRepository(memory.go)
//	                                                                                └─ postgres.Repository(postgres/)
//
// Store는 Repository가 메모리인지 DB인지 모른 채 인터페이스로만 사용합니다.

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// 미리 만들어 둔 오류 값(sentinel error)입니다. 호출한 쪽은 errors.Is(err, ErrCaseNotFound)처럼
// 비교해서 어떤 종류의 실패인지 판단하고, gRPC 상태 코드(NotFound 등)로 바꿉니다.
var (
	// ErrDuplicateCase는 같은 법원에 같은 사건번호가 이미 저장되어 있을 때 반환합니다.
	ErrDuplicateCase = errors.New("같은 법원에 같은 사건번호가 이미 등록되어 있습니다")
	// ErrCaseNotFound는 요청한 사건 ID가 저장소에 없을 때 반환합니다.
	ErrCaseNotFound = errors.New("사건을 찾을 수 없습니다")
)

// ValidationError는 사용자 입력이 잘못되어 저장하지 않았음을 나타냅니다.
// 저장소 장애와 구분해 gRPC InvalidArgument로 돌려주기 위해 씁니다.
type ValidationError struct {
	Err error
}

// Error 메서드가 있으면 Go의 error 인터페이스를 만족하므로 일반 오류처럼 돌려줄 수 있습니다.
func (e *ValidationError) Error() string { return e.Err.Error() }

// Unwrap은 감싼 원래 오류를 돌려줍니다. errors.Is/As가 안쪽 오류까지 살펴볼 수 있게 합니다.
func (e *ValidationError) Unwrap() error { return e.Err }

// StoredCase는 저장소가 ID를 붙여 보관한 사건입니다.
type StoredCase struct {
	ID   string
	Case Case
}

// Repository는 사건을 실제로 보관하는 곳(메모리, PostgreSQL)이 갖춰야 할 기능입니다.
//
// interface는 "이런 메서드들을 가진 타입이면 무엇이든 된다"는 약속입니다. MemoryRepository와
// postgres.Repository는 모두 아래 네 메서드를 가지고 있으므로 Repository로 쓸 수 있습니다.
// 모든 메서드가 ctx(context.Context)를 받는 것은 요청이 취소되거나 시간이 초과되면 DB 작업도 멈추기 위해서입니다.
// 입력 검증과 법원 확인은 Store가 맡고, Repository는 저장과 조회만 합니다.
type Repository interface {
	// List는 최근 등록한 사건부터, 같은 때 등록한 사건은 사건번호가 큰 순서로 돌려줍니다.
	List(ctx context.Context) ([]StoredCase, error)
	// Get은 ID로 사건 한 건을 찾고, 없으면 ErrCaseNotFound를 돌려줍니다.
	Get(ctx context.Context, id string) (StoredCase, error)
	// Insert는 검증이 끝난 사건 여러 건을 한 번에 저장합니다. 하나라도 실패하면 아무것도 저장하지 않으며,
	// 같은 법원·사건번호가 이미 있으면 ErrDuplicateCase를 돌려줍니다.
	Insert(ctx context.Context, cases ...Case) ([]StoredCase, error)
	// Count는 저장된 사건 수를 돌려줍니다.
	Count(ctx context.Context) (int, error)
}

// Record는 사건과 화면에 보여 줄 법원 정보를 함께 묶은 값입니다.
type Record struct {
	ID    string
	Case  Case
	Court Court
}

// Store는 사건을 검증해 Repository에 저장하고, 조회 결과에 법원 정보를 붙이는 Clio의 서비스 계층입니다.
// 법원 목록은 시작할 때 Leviathan에서 받은 스냅샷이며 바뀌지 않으므로 잠금 없이 읽습니다.
type Store struct {
	repo   Repository       // 실제 저장소(메모리 또는 PostgreSQL)
	courts map[string]Court // 법원 ID → 법원 정보. ID로 바로 찾기 위해 map을 씁니다.
}

// NewStore는 저장소와 법원 목록으로 서비스 계층을 만듭니다.
func NewStore(repo Repository, courts []Court) *Store {
	s := &Store{repo: repo, courts: make(map[string]Court, len(courts))}
	for _, court := range courts {
		s.courts[court.ID] = cloneCourt(court)
	}
	return s
}

// SeedIfEmpty는 저장소가 비어 있을 때만 cases를 검증해 한 번에 저장하고, 저장한 건수를 돌려줍니다.
// 이미 사건이 있으면 아무것도 하지 않고 0을 돌려줍니다.
func (s *Store) SeedIfEmpty(ctx context.Context, cases []Case) (int, error) {
	count, err := s.repo.Count(ctx)
	if err != nil {
		return 0, err
	}
	if count > 0 || len(cases) == 0 {
		return 0, nil
	}
	// 더미 사건도 사용자가 등록한 사건과 똑같이 검사해서, 잘못된 더미 자료가 DB에 들어가지 않게 합니다.
	for _, c := range cases {
		if _, err := s.validate(c); err != nil {
			return 0, fmt.Errorf("초기 사건 %s: %w", c.Number, err)
		}
	}
	stored, err := s.repo.Insert(ctx, cases...)
	if err != nil {
		return 0, err
	}
	return len(stored), nil
}

// List는 저장된 사건에 법원 정보를 붙여 최신 사건부터 돌려줍니다.
func (s *Store) List(ctx context.Context) ([]Record, error) {
	stored, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	records := make([]Record, len(stored))
	for i, item := range stored {
		records[i] = s.record(item)
	}
	return records, nil
}

// Get은 사건 ID로 사건 한 건을 찾습니다.
func (s *Store) Get(ctx context.Context, id string) (Record, error) {
	stored, err := s.repo.Get(ctx, id)
	if err != nil {
		return Record{}, err
	}
	return s.record(stored), nil
}

// Create는 새 사건을 검사한 뒤 저장합니다.
// 사건의 법원과 재판부가 Clio 시작 때 받은 Leviathan 목록에 있는지도 확인합니다.
func (s *Store) Create(ctx context.Context, c Case) (Record, error) {
	// 클라이언트가 이미 요청을 취소했다면 더 진행하지 않습니다.
	if err := ctx.Err(); err != nil {
		return Record{}, err
	}
	c = normalize(c) // 1) 앞뒤 공백 정리
	// 2) 검사: 필드 형식, 법원 존재, 재판부 소속
	if _, err := s.validate(c); err != nil {
		return Record{}, err
	}
	// 3) 저장: 중복 사건번호면 ErrDuplicateCase가 돌아옵니다.
	stored, err := s.repo.Insert(ctx, c)
	if err != nil {
		return Record{}, err
	}
	// 4) 저장 결과(ID 포함)에 법원 정보를 붙여 돌려줍니다.
	return s.record(stored[0]), nil
}

// validate는 사건 필드와 법원·재판부 소속을 확인하고 해당 법원을 돌려줍니다.
// 문제가 있으면 ValidationError로 감싸 돌려줍니다.
func (s *Store) validate(c Case) (Court, error) {
	court, err := s.checkCase(c)
	if err != nil {
		return Court{}, &ValidationError{Err: err}
	}
	return court, nil
}

// checkCase는 validate의 실제 검사입니다.
func (s *Store) checkCase(c Case) (Court, error) {
	if err := c.ValidateFields(); err != nil {
		return Court{}, err
	}
	if c.Court.CourtID == "" {
		return Court{}, errors.New("계속 법원이 비어 있습니다")
	}
	// map에서 값을 꺼낼 때 두 번째 값 ok는 "그 키가 있었는지"를 알려 줍니다.
	court, ok := s.courts[c.Court.CourtID]
	if !ok {
		return Court{}, fmt.Errorf("법원 ID %q는 법원 목록에 없습니다", c.Court.CourtID)
	}
	if c.Court.PanelName == "" {
		return Court{}, fmt.Errorf("%s의 재판부가 비어 있습니다", court.Name)
	}
	for _, panel := range court.Panels {
		if panel.Name == c.Court.PanelName {
			return court, nil
		}
	}
	return Court{}, fmt.Errorf("재판부 %q는 %s 소속이 아닙니다", c.Court.PanelName, court.Name)
}

// record는 저장된 사건에 법원 정보를 붙입니다. 저장 뒤 법원 목록에서 빠진 법원이면 ID만 채웁니다.
func (s *Store) record(item StoredCase) Record {
	court, ok := s.courts[item.Case.Court.CourtID]
	if !ok {
		court = Court{ID: item.Case.Court.CourtID}
	}
	return Record{ID: item.ID, Case: cloneCase(item.Case), Court: cloneCourt(court)}
}

// normalize는 사용자가 입력한 문자열 앞뒤 공백을 지웁니다.
func normalize(c Case) Case {
	c = cloneCase(c)
	c.Number.Code = strings.TrimSpace(c.Number.Code)
	c.Title = strings.TrimSpace(c.Title)
	c.Court.CourtID = strings.TrimSpace(c.Court.CourtID)
	c.Court.PanelName = strings.TrimSpace(c.Court.PanelName)
	for i := range c.Parties {
		c.Parties[i].Role = strings.TrimSpace(c.Parties[i].Role)
		c.Parties[i].Name = strings.TrimSpace(c.Parties[i].Name)
	}
	return c
}

// caseKey는 같은 법원 안에서 사건번호가 중복되는지 확인할 때 쓰는 고유 키입니다.
func caseKey(c Case) string {
	return c.Court.CourtID + "/" + c.Number.String()
}

// cloneCase는 내부 슬라이스인 당사자 목록을 복사합니다.
// 구조체를 대입하면 값이 복사되지만, 슬라이스는 같은 배열을 가리키는 "참조"만 복사됩니다.
// 그래서 복사본에서 Parties를 고치면 원본까지 바뀌는 문제를 막으려고 슬라이스를 새로 만듭니다.
func cloneCase(c Case) Case {
	c.Parties = append([]Party(nil), c.Parties...)
	return c
}

// cloneCourt는 내부 슬라이스인 재판부 목록을 복사합니다.
func cloneCourt(c Court) Court {
	c.Panels = append([]CourtPanel(nil), c.Panels...)
	return c
}
