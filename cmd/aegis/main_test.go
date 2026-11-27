package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	leviathanv1 "github.com/mkkim-gr8/casepulse/gen/go/leviathan/v1"
	"github.com/mkkim-gr8/casepulse/internal/clio"
	"github.com/mkkim-gr8/casepulse/internal/cliogrpc"
	cliogrpcserver "github.com/mkkim-gr8/casepulse/internal/cliogrpc/server"
	"github.com/mkkim-gr8/casepulse/internal/healthcheck"
	"github.com/mkkim-gr8/casepulse/internal/leviathangrpc"
	leviathangrpcserver "github.com/mkkim-gr8/casepulse/internal/leviathangrpc/server"
	"github.com/mkkim-gr8/casepulse/web"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const expectedDummyCaseCount = 240

func newTestServices(t *testing.T) ([]*leviathanv1.Court, *leviathangrpc.Client, *cliogrpc.Client) {
	t.Helper()
	leviathanListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	clioListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		leviathanListener.Close()
		t.Fatal(err)
	}
	leviathanServer := grpc.NewServer()
	leviathangrpcserver.RegisterServer(leviathanServer)
	go leviathanServer.Serve(leviathanListener)
	t.Cleanup(func() {
		leviathanServer.Stop()
	})

	leviathanConn, err := grpc.NewClient(leviathanListener.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	leviathanClient := leviathangrpc.NewClient(leviathanConn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	courts, err := leviathanClient.ListCourts(ctx, grpc.WaitForReady(true))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		leviathanConn.Close()
	})

	courtData := make([]clio.Court, len(courts))
	for i, court := range courts {
		panels := make([]clio.CourtPanel, len(court.Panels))
		for j, panel := range court.Panels {
			panels[j] = clio.CourtPanel{ID: panel.Id, Name: panel.Name}
		}
		courtData[i] = clio.Court{
			ID: court.Id, Name: court.Name, Seat: court.Seat,
			PostalCode: court.PostalCode, Address: court.Address,
			Level: court.Level, CourtType: court.CourtType, UnitType: court.UnitType,
			Panels: panels,
		}
	}
	clioServer := grpc.NewServer()
	store := clio.NewStore(clio.NewMemoryRepository(), courtData)
	if _, err := store.SeedIfEmpty(context.Background(), clio.BuildDummyCases(courtData)); err != nil {
		t.Fatal(err)
	}
	cliogrpcserver.RegisterServer(clioServer, store)
	healthcheck.RegisterGRPC(clioServer)
	go clioServer.Serve(clioListener)
	clioConn, err := grpc.NewClient(clioListener.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		clioServer.Stop()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		clioConn.Close()
		clioServer.Stop()
	})
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := healthcheck.CheckConn(ctx, clioConn, grpc.WaitForReady(true)); err != nil {
		t.Fatalf("Clio health check: %v", err)
	}
	return courts, leviathanClient, cliogrpc.NewClient(clioConn)
}

func newTestHTTPServer(t *testing.T) *httptest.Server {
	t.Helper()
	courts, _, clioClient := newTestServices(t)
	server := httptest.NewServer(newHandler(courts, clioClient))
	t.Cleanup(server.Close)
	return server
}

func TestLeviathanDirectoryRPCs(t *testing.T) {
	_, client, _ := newTestServices(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	offices, err := client.ListProsecutionOffices(ctx, grpc.WaitForReady(true))
	if err != nil {
		t.Fatal(err)
	}
	if len(offices) != 67 {
		t.Fatalf("got %d prosecution offices, want 67", len(offices))
	}
	office, err := client.GetProsecutionOffice(ctx, "prosecution-national", grpc.WaitForReady(true))
	if err != nil {
		t.Fatal(err)
	}
	if office.Address == "" || office.PostalCode != "06590" {
		t.Errorf("prosecution office address = %q %q", office.PostalCode, office.Address)
	}

	agencies, err := client.ListInvestigationAgencies(ctx, grpc.WaitForReady(true))
	if err != nil {
		t.Fatal(err)
	}
	if len(agencies) != 7 {
		t.Fatalf("got %d investigation agencies, want 7", len(agencies))
	}
	agency, err := client.GetInvestigationAgency(ctx, "investigation-national", grpc.WaitForReady(true))
	if err != nil {
		t.Fatal(err)
	}
	if agency.Address == "" || agency.InstitutionType != "investigation" {
		t.Errorf("investigation agency has missing address or type: %+v", agency)
	}
}

func TestCasesAPI(t *testing.T) {
	srv := newTestHTTPServer(t)

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
	if body.Total != expectedDummyCaseCount || len(body.Cases) != expectedDummyCaseCount {
		t.Fatalf("total=%d len=%d, want %d", body.Total, len(body.Cases), expectedDummyCaseCount)
	}
	head, err := http.Head(srv.URL + "/api/cases")
	if err != nil {
		t.Fatal(err)
	}
	head.Body.Close()
	if head.StatusCode != http.StatusOK {
		t.Errorf("HEAD /api/cases = %d, want 200", head.StatusCode)
	}
	req, err := http.NewRequest(http.MethodPatch, srv.URL+"/api/cases", nil)
	if err != nil {
		t.Fatal(err)
	}
	methodNotAllowed, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	methodNotAllowed.Body.Close()
	if methodNotAllowed.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("PATCH /api/cases = %d, want 405", methodNotAllowed.StatusCode)
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
	if _, err := web.Dist.Open("dist/index.html"); err == nil {
		want = http.StatusOK
	}
	if page.StatusCode != want {
		t.Errorf("GET / = %d, want %d", page.StatusCode, want)
	}
	if _, err := web.Dist.Open("dist/casepulse-logo.svg"); err == nil {
		asset, err := http.Get(srv.URL + "/casepulse-logo.svg")
		if err != nil {
			t.Fatal(err)
		}
		asset.Body.Close()
		if asset.StatusCode != http.StatusOK {
			t.Errorf("GET /casepulse-logo.svg = %d, want 200", asset.StatusCode)
		}
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
	srv := newTestHTTPServer(t)

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
	unknownCourt := strings.Replace(valid, `"court-district-seoul-central"`, `"court-unknown"`, 1)
	status, body = postCase(t, srv.URL, unknownCourt)
	if status != http.StatusBadRequest || !strings.Contains(fmt.Sprint(body["errors"]), "법원 목록에 없습니다") {
		t.Errorf("unknown court = %d %v", status, body)
	}
	unknownPanel := strings.Replace(valid, `"제91민사단독"`, `"없는재판부"`, 1)
	status, body = postCase(t, srv.URL, unknownPanel)
	if status != http.StatusBadRequest || !strings.Contains(fmt.Sprint(body["errors"]), "소속이 아닙니다") {
		t.Errorf("unknown panel = %d %v", status, body)
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
	if list.Total != expectedDummyCaseCount+1 || list.Cases[0].Number != "2026가합99999" {
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

func TestHealthz(t *testing.T) {
	srv := newTestHTTPServer(t)
	if err := healthcheck.ProbeHTTP(srv.Listener.Addr().String(), "/healthz"); err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
}

func TestCourtsAPI(t *testing.T) {
	srv := newTestHTTPServer(t)
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
