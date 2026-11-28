package main

// 이 파일은 사건·법원 JSON API를 담습니다.
//
//	GET  /api/cases       Clio gRPC 서비스의 사건 전체 목록
//	GET  /api/cases/{id}  Clio gRPC 서비스의 사건 한 건
//	POST /api/cases       Clio gRPC 서비스에 사건 등록
//	GET  /api/courts      사건 등록 화면에서 고를 법원과 재판부 이름
//
// 핸들러는 공통적으로 다음 일을 합니다.
//  1. 브라우저가 보낸 HTTP 요청(경로·JSON 본문)을 읽습니다.
//  2. gRPC 요청 메시지로 바꿔 Clio를 호출합니다.
//  3. gRPC 응답을 화면이 쓰기 좋은 JSON(caseView 등)으로 바꿔 돌려줍니다.
//  4. gRPC 오류 코드를 알맞은 HTTP 상태 코드로 바꿉니다(NotFound→404, InvalidArgument→400,
//     AlreadyExists→409, 그 밖→503).

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	cliov1 "github.com/mkkim-gr8/casepulse/gen/go/clio/v1"
	leviathanv1 "github.com/mkkim-gr8/casepulse/gen/go/leviathan/v1"
	"github.com/mkkim-gr8/casepulse/internal/cliogrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// categoryLabels는 Clio가 보내는 사건 종류 코드를 화면에 보여 줄 한글 이름으로 바꿉니다.
var categoryLabels = map[string]string{
	"civil":             "민사",
	"civil_application": "민사신청",
	"family":            "가사",
	"administrative":    "행정",
	"criminal":          "형사",
}

// partyView는 HTTP JSON으로 주고받을 당사자 정보입니다.
//
// 필드 뒤의 `json:"role"` 같은 표시(태그)는 JSON으로 바꿀 때 쓸 키 이름입니다.
// Go 필드는 대문자(Role)로 시작해야 밖에서 읽을 수 있지만, JSON은 소문자 키(role)를 쓰기 위해 붙입니다.
type partyView struct {
	Role   string `json:"role"`
	Name   string `json:"name"`
	Client bool   `json:"client"`
}

// courtView는 사건 화면에 표시할 법원과 재판부 정보입니다.
type courtView struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Panel      string `json:"panel"`
	Seat       string `json:"seat"`
	PostalCode string `json:"postalCode"`
	Address    string `json:"address"`
}

// caseView는 gRPC 사건 응답을 웹 화면에 맞는 JSON 모양으로 바꾼 자료입니다.
// 화면에서 바로 쓰도록 사건번호 문자열(Number)과 사건 종류 한글 이름(CategoryLabel)을 미리 만들어 넣습니다.
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

// courtOption은 신건등록 화면에서 법원과 재판부를 고를 때 쓰는 항목입니다.
type courtOption struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Panels []string `json:"panels"`
}

// createCaseRequest는 브라우저가 POST /api/cases로 보내는 JSON 형식입니다.
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

// apiHandlers는 사건 gRPC 클라이언트와 화면에 보여 줄 법원 목록을 보관합니다.
type apiHandlers struct {
	clio   *cliogrpc.Client
	courts []*leviathanv1.Court
}

// newAPIHandlers는 HTTP 요청을 처리하는 데 필요한 클라이언트와 법원 목록을 저장합니다.
func newAPIHandlers(clioClient *cliogrpc.Client, courts []*leviathanv1.Court) *apiHandlers {
	return &apiHandlers{
		clio:   clioClient,
		courts: append([]*leviathanv1.Court(nil), courts...), // 호출한 쪽 슬라이스와 분리하기 위해 복사
	}
}

// routes는 사건 API와 법원 선택 API의 URL 경로를 Gin 서버에 등록합니다.
//
// 각 핸들러는 func(c *gin.Context) 모양의 함수입니다. c에서 요청 정보를 읽고(c.Param, c.Request),
// c.JSON(상태코드, 값)으로 응답을 씁니다. HEAD 요청은 본문 없이 헤더만 돌려주는 GET이라 같은 함수를 씁니다.
func (h *apiHandlers) routes(router *gin.Engine) {
	// GET /api/cases: 사건 전체 목록
	listCases := func(c *gin.Context) {
		// 브라우저 요청의 context를 바탕으로 5초 제한을 겁니다. 브라우저가 연결을 끊으면 Clio 호출도 취소됩니다.
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		cases, total, err := h.clio.ListCases(ctx, grpc.WaitForReady(true))
		if err != nil {
			h.serviceUnavailable(c, "list cases", err)
			return
		}
		views := make([]caseView, len(cases))
		for i, item := range cases {
			view, err := caseToView(item)
			if err != nil {
				h.serviceUnavailable(c, "decode case list", err)
				return
			}
			views[i] = view
		}
		// map[string]any는 키가 문자열이고 값은 아무 타입이나 되는 사전입니다. JSON 객체 {"cases": [...], "total": 240}이 됩니다.
		c.JSON(http.StatusOK, map[string]any{"cases": views, "total": total})
	}
	router.GET("/api/cases", listCases)
	router.HEAD("/api/cases", listCases)

	// GET /api/cases/:id: 사건 한 건. ":id" 자리의 값은 c.Param("id")로 읽습니다.
	getCase := func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		item, err := h.clio.GetCase(ctx, c.Param("id"), grpc.WaitForReady(true))
		if err != nil {
			if status.Code(err) == codes.NotFound {
				c.JSON(http.StatusNotFound, map[string]string{"error": "사건을 찾을 수 없습니다"})
				return
			}
			h.serviceUnavailable(c, "get case", err)
			return
		}
		view, err := caseToView(item)
		if err != nil {
			h.serviceUnavailable(c, "decode case", err)
			return
		}
		c.JSON(http.StatusOK, view)
	}
	router.GET("/api/cases/:id", getCase)
	router.HEAD("/api/cases/:id", getCase)

	// POST /api/cases: 사건 등록
	router.POST("/api/cases", func(c *gin.Context) {
		var req createCaseRequest
		// 요청 본문 크기를 1MB(1<<20 바이트)로 제한해 너무 큰 요청으로 서버가 힘들어지는 것을 막습니다.
		dec := json.NewDecoder(http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20))
		// 정의하지 않은 필드가 오면 오류로 처리합니다(오타 난 필드를 조용히 무시하지 않기 위해).
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			c.JSON(http.StatusBadRequest, map[string]any{"errors": []string{"요청 형식이 올바르지 않습니다: " + err.Error()}})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		item, err := h.clio.CreateCase(ctx, caseInputFromRequest(req), grpc.WaitForReady(true))
		if err != nil {
			// status.Code(err)는 gRPC 오류에서 상태 코드를 꺼냅니다.
			switch status.Code(err) {
			case codes.InvalidArgument:
				c.JSON(http.StatusBadRequest, map[string]any{"errors": errorMessages(status.Convert(err).Message())})
			case codes.AlreadyExists:
				c.JSON(http.StatusConflict, map[string]any{"errors": []string{status.Convert(err).Message()}})
			default:
				h.serviceUnavailable(c, "create case", err)
			}
			return
		}
		view, err := caseToView(item)
		if err != nil {
			h.serviceUnavailable(c, "decode created case", err)
			return
		}
		c.JSON(http.StatusCreated, view) // 201 Created: 새 자원을 만들었다는 뜻
	})

	// GET /api/courts: 등록 화면의 법원·재판부 선택지. 시작할 때 받은 목록을 그대로 씁니다.
	listCourts := func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]any{"courts": courtOptions(h.courts)})
	}
	router.GET("/api/courts", listCourts)
	router.HEAD("/api/courts", listCourts)
}

// serviceUnavailable은 gRPC 호출 실패를 서버 로그에 남기고 브라우저에는 503을 돌려줍니다.
// 자세한 내부 오류는 화면에 보여 주지 않고 로그에만 남깁니다(내부 구조 노출 방지).
func (h *apiHandlers) serviceUnavailable(c *gin.Context, operation string, err error) {
	log.Printf("Clio %s failed: %v", operation, err)
	c.JSON(http.StatusServiceUnavailable, map[string]any{"errors": []string{"사건 서비스를 사용할 수 없습니다"}})
}

// caseInputFromRequest는 HTTP 요청 자료를 Clio gRPC가 받는 자료형으로 옮깁니다.
func caseInputFromRequest(req createCaseRequest) *cliov1.CaseInput {
	parties := make([]*cliov1.Party, len(req.Parties))
	for i, party := range req.Parties {
		parties[i] = &cliov1.Party{Role: party.Role, Name: party.Name, Client: party.Client}
	}
	return &cliov1.CaseInput{
		Number:   &cliov1.CaseNumber{Year: int64(req.Number.Year), Code: req.Number.Code, Serial: int64(req.Number.Serial)},
		Title:    req.Title,
		Category: req.Category,
		Parties:  parties,
		Court:    &cliov1.CourtAssignment{Id: req.Court.ID, Panel: req.Court.Panel},
	}
}

// caseToView는 Clio의 사건 응답을 기존 웹 화면이 사용하는 JSON 자료형으로 바꿉니다.
// 필수 정보가 빠졌으면 빈 응답을 만들지 않고 오류를 돌려줍니다.
func caseToView(item *cliov1.Case) (caseView, error) {
	if item == nil {
		return caseView{}, errors.New("Clio returned an empty case")
	}
	if item.Number == nil {
		return caseView{}, errors.New("Clio returned a case without a number")
	}
	if item.Court == nil {
		return caseView{}, errors.New("Clio returned a case without a court")
	}
	year, serial, code := int(item.Number.Year), int(item.Number.Serial), item.Number.Code
	parties := make([]partyView, len(item.Parties))
	for i, party := range item.Parties {
		if party == nil {
			return caseView{}, errors.New("Clio returned an empty party")
		}
		parties[i] = partyView{Role: party.Role, Name: party.Name, Client: party.Client}
	}
	court := courtView{
		ID: item.Court.Id, Name: item.Court.Name, Panel: item.Court.Panel,
		Seat: item.Court.Seat, PostalCode: item.Court.PostalCode, Address: item.Court.Address,
	}
	return caseView{
		ID: item.Id, Number: caseNumber(year, code, serial), Year: year, Code: code, Serial: serial,
		Title: item.Title, Category: item.Category, CategoryLabel: categoryLabels[item.Category],
		Parties: parties, Court: court,
	}, nil
}

// courtOptions는 재판부가 있는 법원만 골라 등록 화면에서 쓸 목록으로 바꿉니다.
func courtOptions(courts []*leviathanv1.Court) []courtOption {
	var options []courtOption
	for _, court := range courts {
		if court == nil || len(court.Panels) == 0 {
			continue
		}
		panels := make([]string, len(court.Panels))
		for i, panel := range court.Panels {
			if panel != nil {
				panels[i] = panel.Name
			}
		}
		options = append(options, courtOption{ID: court.Id, Name: court.Name, Panels: panels})
	}
	return options
}

// errorMessages는 Clio가 줄바꿈으로 묶어 보낸 검증 오류를 화면용 목록으로 나눕니다.
func errorMessages(message string) []string {
	var messages []string
	for _, line := range strings.Split(message, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			messages = append(messages, line)
		}
	}
	if len(messages) == 0 {
		return []string{"요청을 처리할 수 없습니다"}
	}
	return messages
}

// caseNumber는 연도·사건부호·일련번호를 "2026가합12345" 형태로 합칩니다.
func caseNumber(year int, code string, serial int) string {
	if year == 0 && code == "" && serial == 0 {
		return ""
	}
	return fmt.Sprintf("%04d%s%d", year, code, serial)
}
