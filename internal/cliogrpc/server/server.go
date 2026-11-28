// Package server는 Clio의 사건 자료형과 gRPC 요청 사이를 변환합니다.
//
// 요청 하나가 처리되는 순서:
//  1. gRPC 요청 메시지(cliov1.CreateCaseRequest 등)를 받습니다.
//  2. caseFromProto로 Clio 내부 자료형(clio.Case)으로 옮깁니다.
//  3. clio.Store에 맡겨 검증·저장·조회를 합니다.
//  4. 결과를 recordToProto로 다시 gRPC 응답 메시지로 옮기거나, 오류를 gRPC 상태 코드로 바꿉니다.
//
// 오류 → gRPC 상태 코드 대응표:
//
//	clio.ErrCaseNotFound   → NotFound        (Aegis가 HTTP 404로 바꿈)
//	clio.ValidationError   → InvalidArgument (HTTP 400)
//	clio.ErrDuplicateCase  → AlreadyExists   (HTTP 409)
//	그 밖의 저장소 오류     → Unavailable     (HTTP 503)
package server

import (
	"context"
	"errors"
	"log"

	cliov1 "github.com/mkkim-gr8/casepulse/gen/go/clio/v1"
	"github.com/mkkim-gr8/casepulse/internal/clio"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// server는 gRPC 요청을 Clio의 서비스 계층(clio.Store)으로 연결합니다.
type server struct {
	cliov1.UnimplementedClioServiceServer
	store *clio.Store
}

// RegisterServer는 전달받은 gRPC 서버에 Clio 서비스를 등록합니다.
func RegisterServer(registrar grpc.ServiceRegistrar, store *clio.Store) {
	cliov1.RegisterClioServiceServer(registrar, server{store: store})
}

// ListCases는 저장소의 사건을 Protobuf 응답으로 바꿔 반환합니다.
func (s server) ListCases(ctx context.Context, _ *cliov1.ListCasesRequest) (*cliov1.ListCasesResponse, error) {
	records, err := s.store.List(ctx)
	if err != nil {
		return nil, storeError(ctx, err)
	}
	cases := make([]*cliov1.Case, len(records))
	for i, record := range records {
		cases[i] = recordToProto(record)
	}
	return &cliov1.ListCasesResponse{Cases: cases, Total: int32(len(cases))}, nil
}

// GetCase는 ID로 사건을 찾아 반환하고, 없으면 gRPC NotFound 오류를 돌려줍니다.
func (s server) GetCase(ctx context.Context, request *cliov1.GetCaseRequest) (*cliov1.Case, error) {
	// 생성 코드의 요청 값은 포인터라 nil일 수 있으므로, 필드를 읽기 전에 먼저 확인합니다.
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "사건 ID가 비어 있습니다")
	}
	record, err := s.store.Get(ctx, request.Id)
	if errors.Is(err, clio.ErrCaseNotFound) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, storeError(ctx, err)
	}
	return recordToProto(record), nil
}

// CreateCase는 Protobuf 요청을 사건 자료형으로 바꿔 저장소에 등록합니다.
// 입력 오류와 중복 사건을 서로 다른 gRPC 상태 코드로 알려줍니다.
func (s server) CreateCase(ctx context.Context, request *cliov1.CreateCaseRequest) (*cliov1.Case, error) {
	if request == nil || request.Input == nil {
		return nil, status.Error(codes.InvalidArgument, "사건 정보가 비어 있습니다")
	}
	record, err := s.store.Create(ctx, caseFromProto(request.Input))
	var validation *clio.ValidationError
	// 조건 없는 switch는 위에서부터 차례로 case 조건을 확인해 처음 참인 곳을 실행합니다(if-else 사슬과 같음).
	// errors.Is는 "이 오류 값과 같은가", errors.As는 "이 타입의 오류가 들어 있는가"를 확인합니다.
	switch {
	case errors.Is(err, clio.ErrDuplicateCase):
		return nil, status.Error(codes.AlreadyExists, err.Error())
	case errors.As(err, &validation):
		return nil, status.Error(codes.InvalidArgument, validation.Error())
	case err != nil:
		return nil, storeError(ctx, err)
	default:
		return recordToProto(record), nil
	}
}

// storeError는 저장소 오류를 gRPC 상태로 바꿉니다. DB 오류의 자세한 내용은 Clio 로그에만 남기고
// 호출한 쪽에는 일반적인 메시지만 돌려줍니다.
func storeError(ctx context.Context, err error) error {
	// 요청이 취소되었거나 시간이 초과되어 실패한 것이면 Canceled/DeadlineExceeded로 그대로 알립니다.
	if ctx.Err() != nil {
		return status.FromContextError(ctx.Err()).Err()
	}
	log.Printf("clio store error: %v", err)
	return status.Error(codes.Unavailable, "사건 저장소를 사용할 수 없습니다")
}

// caseFromProto는 gRPC 요청의 필드를 Clio 내부 사건 자료형으로 옮깁니다.
func caseFromProto(input *cliov1.CaseInput) clio.Case {
	c := clio.Case{
		Title:    input.Title,
		Category: clio.CaseCategory(input.Category),
	}
	// 중첩 메시지(Number, Court)는 비어 있으면 nil이므로, 있을 때만 옮깁니다.
	// 비어 있으면 0값으로 남고, 이후 Store의 검증에서 오류로 걸러집니다.
	if input.Number != nil {
		c.Number = clio.CaseNumber{
			Year: int(input.Number.Year), Code: input.Number.Code, Serial: int(input.Number.Serial),
		}
	}
	if input.Court != nil {
		c.Court = clio.CourtAssignment{CourtID: input.Court.Id, PanelName: input.Court.Panel}
	}
	c.Parties = make([]clio.Party, len(input.Parties))
	for i, party := range input.Parties {
		if party != nil {
			c.Parties[i] = clio.Party{Role: party.Role, Name: party.Name, Client: party.Client}
		}
	}
	return c
}

// recordToProto는 저장된 사건과 법원 정보를 gRPC 응답 자료형으로 옮깁니다.
func recordToProto(record clio.Record) *cliov1.Case {
	parties := make([]*cliov1.Party, len(record.Case.Parties))
	for i, party := range record.Case.Parties {
		parties[i] = &cliov1.Party{Role: party.Role, Name: party.Name, Client: party.Client}
	}
	return &cliov1.Case{
		Id: record.ID,
		Number: &cliov1.CaseNumber{
			Year: int64(record.Case.Number.Year), Code: record.Case.Number.Code, Serial: int64(record.Case.Number.Serial),
		},
		Title: record.Case.Title, Category: string(record.Case.Category), Parties: parties,
		Court: &cliov1.CaseCourt{
			Id: record.Court.ID, Panel: record.Case.Court.PanelName,
			Name: record.Court.Name, Seat: record.Court.Seat,
			PostalCode: record.Court.PostalCode, Address: record.Court.Address,
		},
	}
}
