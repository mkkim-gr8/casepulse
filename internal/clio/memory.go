package clio

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// MemoryRepository는 프로세스 안에서 사건을 보관하는 저장소입니다.
// DB 없이 실행하거나 테스트할 때 쓰며, 프로세스를 재시작하면 저장한 사건은 사라집니다.
type MemoryRepository struct {
	mu     sync.RWMutex          // 여러 gRPC 요청이 동시에 접근할 때 자료를 보호합니다.
	list   []StoredCase          // 목록 API가 반환할 사건 순서를 보관합니다.
	byID   map[string]StoredCase // 사건 ID를 이용해 한 건을 빠르게 찾습니다.
	keys   map[string]bool       // 법원 ID와 사건번호 조합의 중복 여부를 기억합니다.
	nextID int                   // 새 사건에 순서대로 붙일 임시 ID의 번호입니다.
}

// NewMemoryRepository는 빈 메모리 저장소를 만듭니다.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{byID: make(map[string]StoredCase), keys: make(map[string]bool)}
}

// List는 최신 사건부터 복사해 돌려줍니다.
func (m *MemoryRepository) List(context.Context) ([]StoredCase, error) {
	// RLock은 "읽기 잠금"입니다. 여러 요청이 동시에 읽는 것은 허용하고, 쓰는 동안에는 기다리게 합니다.
	m.mu.RLock()
	defer m.mu.RUnlock()
	items := make([]StoredCase, len(m.list))
	for i, item := range m.list {
		items[i] = StoredCase{ID: item.ID, Case: cloneCase(item.Case)}
	}
	return items, nil
}

// Get은 ID로 사건 한 건을 찾습니다.
func (m *MemoryRepository) Get(_ context.Context, id string) (StoredCase, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	item, ok := m.byID[id]
	if !ok {
		return StoredCase{}, ErrCaseNotFound
	}
	return StoredCase{ID: item.ID, Case: cloneCase(item.Case)}, nil
}

// Insert는 사건들을 한 번에 저장합니다. 중복이 하나라도 있으면 아무것도 저장하지 않습니다.
// 한 번에 넣은 사건들은 사건번호가 큰 순서로 목록 맨 앞에 놓입니다.
func (m *MemoryRepository) Insert(ctx context.Context, cases ...Case) ([]StoredCase, error) {
	// Lock은 "쓰기 잠금"입니다. 저장하는 동안에는 다른 읽기·쓰기가 모두 기다립니다.
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// 1단계: 먼저 전부 검사만 합니다. 중간에 실패해도 아무것도 저장되지 않도록(DB 트랜잭션과 같은 효과)
	// 저장은 검사가 모두 끝난 뒤에 합니다. seen은 이번에 함께 넣는 사건끼리의 중복을 잡습니다.
	seen := make(map[string]bool, len(cases))
	for _, c := range cases {
		key := caseKey(c)
		if m.keys[key] || seen[key] {
			return nil, ErrDuplicateCase
		}
		seen[key] = true
	}

	// 2단계: 실제로 저장합니다. ID는 "case-0001"처럼 순서대로 붙입니다(PostgreSQL은 UUID를 씁니다).
	stored := make([]StoredCase, len(cases))
	for i, c := range cases {
		m.nextID++
		stored[i] = StoredCase{ID: fmt.Sprintf("case-%04d", m.nextID), Case: cloneCase(c)}
		m.byID[stored[i].ID] = stored[i]
		m.keys[caseKey(c)] = true
	}
	// 3단계: 이번에 넣은 사건들을 사건번호가 큰 순서로 정렬해 목록 맨 앞에 붙입니다(최신 사건이 위로).
	batch := append([]StoredCase(nil), stored...)
	sort.SliceStable(batch, func(i, j int) bool { return numberAfter(batch[i].Case.Number, batch[j].Case.Number) })
	m.list = append(batch, m.list...)

	result := make([]StoredCase, len(stored))
	for i, item := range stored {
		result[i] = StoredCase{ID: item.ID, Case: cloneCase(item.Case)}
	}
	return result, nil
}

// Count는 저장된 사건 수를 돌려줍니다.
func (m *MemoryRepository) Count(context.Context) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.list), nil
}

// numberAfter는 목록에서 a를 b보다 앞에 둘지 판단합니다(연도, 일련번호가 큰 순서).
func numberAfter(a, b CaseNumber) bool {
	if a.Year != b.Year {
		return a.Year > b.Year
	}
	return a.Serial > b.Serial
}
