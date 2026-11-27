package main

// 이 파일은 사건·법원 JSON API를 담습니다.
//
//	GET  /api/cases       사건 전체 목록(화면이 검색·정렬·필터를 직접 합니다)
//	GET  /api/cases/{id}  사건 한 건
//	POST /api/cases       사건 등록(casepulse.Case.Validate로 검사한 뒤 저장)
//	GET  /api/courts      사건 등록 화면에서 고를 법원과 그 재판부 이름
//
// casepulse.Case에는 JSON 태그가 없고 법원은 ID로만 들고 있으므로,
// 화면에 바로 쓸 수 있도록 법원 이름·주소 등을 붙인 caseView로 바꿔 내보냅니다.

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/mkkim-gr8/casepulse"
)

// categoryLabels는 사건 종류의 화면 표시 이름입니다.
var categoryLabels = map[casepulse.CaseCategory]string{
	casepulse.CaseCategoryCivil:            "민사",
	casepulse.CaseCategoryCivilApplication: "민사신청",
	casepulse.CaseCategoryFamily:           "가사",
	casepulse.CaseCategoryAdministrative:   "행정",
	casepulse.CaseCategoryCriminal:         "형사",
}

type partyView struct {
	Role   string `json:"role"`
	Name   string `json:"name"`
	Client bool   `json:"client"`
}

type courtView struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Panel      string `json:"panel"`
	Seat       string `json:"seat"`
	PostalCode string `json:"postalCode"`
	Address    string `json:"address"`
}

type caseView struct {
	ID            string      `json:"id"`
	Number        string      `json:"number"`
	Year          int         `json:"year"`
	Code          string      `json:"code"`
	Serial        int         `json:"serial"`
	Title         string      `json:"title"`
	Category      string      `json:"category"`
	CategoryLabel string      `json:"categoryLabel"`
	Parties       []partyView `json:"parties"`
	Court         courtView   `json:"court"`
}

// courtOption은 사건 등록 화면의 법원 선택 항목입니다.
type courtOption struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Panels []string `json:"panels"`
}

// createCaseRequest는 POST /api/cases 요청 본문입니다.
type createCaseRequest struct {
	Number struct {
		Year   int    `json:"year"`
		Code   string `json:"code"`
		Serial int    `json:"serial"`
	} `json:"number"`
	Title    string      `json:"title"`
	Category string      `json:"category"`
	Parties  []partyView `json:"parties"`
	Court    struct {
		ID    string `json:"id"`
		Panel string `json:"panel"`
	} `json:"court"`
}

// caseStore는 메모리에 든 사건 목록입니다. 데이터베이스를 붙이면 이 부분을 바꿉니다.
// 등록한 사건은 메모리에만 있으므로 서버를 다시 켜면 사라집니다.
type caseStore struct {
	mu     sync.RWMutex
	courts map[string]casepulse.Court
	list   []caseView
	byID   map[string]caseView
	// keys는 "법원 ID/사건번호"로, 같은 법원에 같은 사건번호를 두 번 등록하지 못하게 합니다.
	keys   map[string]bool
	nextID int
}

// newCaseStore는 사건 목록을 화면용 caseView로 바꿔 담습니다.
// 법원 ID가 법원 목록에 없으면 패닉을 일으킵니다(사건은 미리 Validate를 거쳐야 합니다).
func newCaseStore(cases []casepulse.Case) *caseStore {
	s := &caseStore{
		courts: make(map[string]casepulse.Court),
		byID:   make(map[string]caseView, len(cases)),
		keys:   make(map[string]bool, len(cases)),
	}
	for _, c := range casepulse.Courts() {
		s.courts[c.ID] = c
	}
	for _, c := range cases {
		if _, ok := s.courts[c.Court.CourtID]; !ok {
			panic(fmt.Sprintf("사건 %s의 법원 ID %q가 법원 목록에 없습니다", c.Number, c.Court.CourtID))
		}
		s.add(c)
	}
	// 기본 순서: 최근 연도 → 일련번호 큰 것(최근 접수) 순.
	sort.SliceStable(s.list, func(a, b int) bool {
		if s.list[a].Year != s.list[b].Year {
			return s.list[a].Year > s.list[b].Year
		}
		return s.list[a].Serial > s.list[b].Serial
	})
	return s
}

func caseKey(c casepulse.Case) string {
	return c.Court.CourtID + "/" + c.Number.String()
}

// add는 검사를 마친 사건을 목록 맨 앞에 넣고 화면용 값을 돌려줍니다. 호출하는 쪽이 잠금을 잡아야 합니다.
func (s *caseStore) add(c casepulse.Case) caseView {
	court := s.courts[c.Court.CourtID]
	parties := make([]partyView, len(c.Parties))
	for j, p := range c.Parties {
		parties[j] = partyView{Role: p.Role, Name: p.Name, Client: p.Client}
	}
	s.nextID++
	v := caseView{
		ID:            fmt.Sprintf("case-%04d", s.nextID),
		Number:        c.Number.String(),
		Year:          c.Number.Year,
		Code:          c.Number.Code,
		Serial:        c.Number.Serial,
		Title:         c.Title,
		Category:      string(c.Category),
		CategoryLabel: categoryLabels[c.Category],
		Parties:       parties,
		Court: courtView{
			ID: court.ID, Name: court.Name, Panel: c.Court.PanelName,
			Seat: court.Seat, PostalCode: court.PostalCode, Address: court.Address,
		},
	}
	s.list = append([]caseView{v}, s.list...)
	s.byID[v.ID] = v
	s.keys[caseKey(c)] = true
	return v
}

// errDuplicateCase는 같은 법원에 같은 사건번호가 이미 있을 때의 오류입니다.
var errDuplicateCase = errors.New("같은 법원에 같은 사건번호가 이미 등록되어 있습니다")

// create는 요청을 casepulse.Case로 바꿔 검사하고, 통과하면 저장합니다.
func (s *caseStore) create(req createCaseRequest) (caseView, error) {
	parties := make([]casepulse.Party, len(req.Parties))
	for i, p := range req.Parties {
		parties[i] = casepulse.Party{Role: strings.TrimSpace(p.Role), Name: strings.TrimSpace(p.Name), Client: p.Client}
	}
	c := casepulse.Case{
		Number: casepulse.CaseNumber{
			Year: req.Number.Year, Code: strings.TrimSpace(req.Number.Code), Serial: req.Number.Serial,
		},
		Title:    strings.TrimSpace(req.Title),
		Category: casepulse.CaseCategory(req.Category),
		Parties:  parties,
		Court:    casepulse.CourtAssignment{CourtID: req.Court.ID, PanelName: req.Court.Panel},
	}
	if err := c.Validate(); err != nil {
		return caseView{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.keys[caseKey(c)] {
		return caseView{}, errDuplicateCase
	}
	return s.add(c), nil
}

// flattenErrors는 errors.Join으로 묶인 오류를 화면에 한 줄씩 보여 줄 수 있도록 풀어 냅니다.
func flattenErrors(err error) []string {
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		var msgs []string
		for _, e := range joined.Unwrap() {
			msgs = append(msgs, flattenErrors(e)...)
		}
		return msgs
	}
	return []string{err.Error()}
}

// courtOptions는 재판부가 하나 이상 있는 법원만 돌려줍니다.
// 사건에는 재판부가 반드시 있어야 하므로, 재판부 정보가 없는 법원은 고를 수 없습니다.
func (s *caseStore) courtOptions() []courtOption {
	var opts []courtOption
	for _, c := range casepulse.Courts() {
		if len(c.Panels) == 0 {
			continue
		}
		panels := make([]string, len(c.Panels))
		for i, p := range c.Panels {
			panels[i] = p.Name
		}
		opts = append(opts, courtOption{ID: c.ID, Name: c.Name, Panels: panels})
	}
	return opts
}

func (s *caseStore) routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/cases", func(w http.ResponseWriter, r *http.Request) {
		s.mu.RLock()
		list := append([]caseView(nil), s.list...)
		s.mu.RUnlock()
		writeJSON(w, http.StatusOK, map[string]any{"cases": list, "total": len(list)})
	})
	mux.HandleFunc("GET /api/cases/{id}", func(w http.ResponseWriter, r *http.Request) {
		s.mu.RLock()
		v, ok := s.byID[r.PathValue("id")]
		s.mu.RUnlock()
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "사건을 찾을 수 없습니다"})
			return
		}
		writeJSON(w, http.StatusOK, v)
	})
	mux.HandleFunc("POST /api/cases", func(w http.ResponseWriter, r *http.Request) {
		var req createCaseRequest
		dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"errors": []string{"요청 형식이 올바르지 않습니다: " + err.Error()}})
			return
		}
		v, err := s.create(req)
		switch {
		case errors.Is(err, errDuplicateCase):
			writeJSON(w, http.StatusConflict, map[string]any{"errors": []string{err.Error()}})
		case err != nil:
			writeJSON(w, http.StatusBadRequest, map[string]any{"errors": flattenErrors(err)})
		default:
			writeJSON(w, http.StatusCreated, v)
		}
	})
	mux.HandleFunc("GET /api/courts", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"courts": s.courtOptions()})
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
