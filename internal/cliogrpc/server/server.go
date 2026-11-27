// Package server는 Clio의 사건 자료형과 gRPC 요청 사이를 변환합니다.
package server

import (
	"context"
	"errors"

	cliov1 "github.com/mkkim-gr8/casepulse/gen/go/clio/v1"
	"github.com/mkkim-gr8/casepulse/internal/clio"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// server는 gRPC 요청을 Clio의 메모리 저장소 기능으로 연결합니다.
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
	records := s.store.List(ctx)
	cases := make([]*cliov1.Case, len(records))
	for i, record := range records {
		cases[i] = recordToProto(record)
	}
	return &cliov1.ListCasesResponse{Cases: cases, Total: int32(len(cases))}, nil
}

// GetCase는 ID로 사건을 찾아 반환하고, 없으면 gRPC NotFound 오류를 돌려줍니다.
func (s server) GetCase(ctx context.Context, request *cliov1.GetCaseRequest) (*cliov1.Case, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "사건 ID가 비어 있습니다")
	}
	record, err := s.store.Get(ctx, request.Id)
	if errors.Is(err, clio.ErrCaseNotFound) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
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
	switch {
	case errors.Is(err, clio.ErrDuplicateCase):
		return nil, status.Error(codes.AlreadyExists, err.Error())
	case err != nil:
		if ctx.Err() != nil {
			return nil, status.FromContextError(ctx.Err()).Err()
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	default:
		return recordToProto(record), nil
	}
}

// caseFromProto는 gRPC 요청의 필드를 Clio 내부 사건 자료형으로 옮깁니다.
func caseFromProto(input *cliov1.CaseInput) clio.Case {
	c := clio.Case{
		Title:    input.Title,
		Category: clio.CaseCategory(input.Category),
	}
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
