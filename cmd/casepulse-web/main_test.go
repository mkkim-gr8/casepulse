package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDummyCases(t *testing.T) {
	cases := buildDummyCases()
	if len(cases) != dummyCaseCount {
		t.Fatalf("got %d cases, want %d", len(cases), dummyCaseCount)
	}
	counts := map[string]int{}
	partialSide := 0 // 의뢰인이 아닌 당사자가 있는 사건 수
	for _, c := range cases {
		if err := c.Validate(); err != nil {
			t.Fatalf("%s: %v", c.Number, err)
		}
		counts[string(c.Category)]++
		if len(c.Clients()) < len(c.Parties) {
			partialSide++
		}
	}
	if partialSide == 0 {
		t.Error("모든 당사자가 의뢰인인 사건만 만들어졌습니다")
	}
	for category := range categoryLabels {
		if counts[string(category)] == 0 {
			t.Errorf("no dummy cases of category %s", category)
		}
	}
}

func TestCasesAPI(t *testing.T) {
	srv := httptest.NewServer(newHandler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/cases")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var body struct {
		Cases []caseView `json:"cases"`
		Total int        `json:"total"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Total != dummyCaseCount || len(body.Cases) != dummyCaseCount {
		t.Fatalf("total=%d len=%d, want %d", body.Total, len(body.Cases), dummyCaseCount)
	}
	first := body.Cases[0]
	if first.Court.Name == "" || first.Court.Panel == "" || first.CategoryLabel == "" || len(first.Parties) == 0 {
		t.Errorf("incomplete case view: %+v", first)
	}

	one, err := http.Get(srv.URL + "/api/cases/" + first.ID)
	if err != nil {
		t.Fatal(err)
	}
	one.Body.Close()
	if one.StatusCode != http.StatusOK {
		t.Errorf("GET /api/cases/%s = %d", first.ID, one.StatusCode)
	}
	missing, err := http.Get(srv.URL + "/api/cases/nope")
	if err != nil {
		t.Fatal(err)
	}
	missing.Body.Close()
	if missing.StatusCode != http.StatusNotFound {
		t.Errorf("GET /api/cases/nope = %d, want 404", missing.StatusCode)
	}

	page, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	page.Body.Close()
	// 화면을 빌드했으면 200, 빌드 전이면 안내 페이지와 함께 503입니다.
	want := http.StatusServiceUnavailable
	if _, err := webDist.Open("web/dist/index.html"); err == nil {
		want = http.StatusOK
	}
	if page.StatusCode != want {
		t.Errorf("GET / = %d, want %d", page.StatusCode, want)
	}
}

// postCase는 사건 등록 요청을 보내고 응답 코드와 본문을 돌려줍니다.
func postCase(t *testing.T, url, body string) (int, map[string]any) {
	t.Helper()
	resp, err := http.Post(url+"/api/cases", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, out
}

func TestCreateCaseAPI(t *testing.T) {
	srv := httptest.NewServer(newHandler())
	defer srv.Close()

	valid := `{
		"number": {"year": 2026, "code": "가합", "serial": 99999},
		"title": "손해배상(기)",
		"category": "civil",
		"parties": [
			{"role": "원고1", "name": "홍길동", "client": true},
			{"role": "원고2", "name": "김철수", "client": false},
			{"role": "피고", "name": "주식회사 가나다", "client": false}
		],
		"court": {"id": "court-district-seoul-central", "panel": "제91민사단독"}
	}`
	status, body := postCase(t, srv.URL, valid)
	if status != http.StatusCreated {
		t.Fatalf("POST valid = %d %v", status, body)
	}
	if body["number"] != "2026가합99999" || body["court"].(map[string]any)["name"] != "서울중앙지방법원" {
		t.Errorf("unexpected created case: %v", body)
	}

	// 등록한 사건이 목록 맨 앞에 나와야 합니다.
	resp, err := http.Get(srv.URL + "/api/cases")
	if err != nil {
		t.Fatal(err)
	}
	var list struct {
		Cases []caseView `json:"cases"`
		Total int        `json:"total"`
	}
	json.NewDecoder(resp.Body).Decode(&list)
	resp.Body.Close()
	if list.Total != dummyCaseCount+1 || list.Cases[0].Number != "2026가합99999" {
		t.Errorf("after create: total=%d first=%s", list.Total, list.Cases[0].Number)
	}

	if status, _ := postCase(t, srv.URL, valid); status != http.StatusConflict {
		t.Errorf("duplicate POST = %d, want 409", status)
	}

	noClient := strings.ReplaceAll(valid, `"client": true`, `"client": false`)
	noClient = strings.ReplaceAll(noClient, "99999", "88888")
	status, body = postCase(t, srv.URL, noClient)
	if status != http.StatusBadRequest || !strings.Contains(fmt.Sprint(body["errors"]), "의뢰인") {
		t.Errorf("no-client POST = %d %v", status, body)
	}

	// 지위가 다른 당사자를 함께 의뢰인으로 하면(이해충돌) 거부합니다.
	bothSides := strings.ReplaceAll(valid, `{"role": "피고", "name": "주식회사 가나다", "client": false}`, `{"role": "피고", "name": "주식회사 가나다", "client": true}`)
	bothSides = strings.ReplaceAll(bothSides, "99999", "77777")
	status, body = postCase(t, srv.URL, bothSides)
	if status != http.StatusBadRequest || !strings.Contains(fmt.Sprint(body["errors"]), "이해충돌") {
		t.Errorf("both-sides POST = %d %v", status, body)
	}

	// 여러 문제가 있으면 모두 한 줄씩 돌려줍니다.
	status, body = postCase(t, srv.URL, `{"number": {"year": 26, "code": "", "serial": 0}, "title": "", "category": "civil", "parties": [], "court": {"id": "", "panel": ""}}`)
	if errs, _ := body["errors"].([]any); status != http.StatusBadRequest || len(errs) < 5 {
		t.Errorf("empty POST = %d %v", status, body)
	}

	if status, _ := postCase(t, srv.URL, `{"unknown": 1}`); status != http.StatusBadRequest {
		t.Errorf("unknown field POST = %d, want 400", status)
	}
}

func TestCourtsAPI(t *testing.T) {
	srv := httptest.NewServer(newHandler())
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/api/courts")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var body struct {
		Courts []courtOption `json:"courts"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body.Courts) == 0 {
		t.Fatal("no courts")
	}
	for _, c := range body.Courts {
		if len(c.Panels) == 0 {
			t.Errorf("%s has no panels but was listed", c.ID)
		}
	}
}
