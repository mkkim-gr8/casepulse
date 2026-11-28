// Package server는 Leviathan의 기관 자료형을 gRPC 메시지로 바꿔 제공합니다.
//
// 이런 코드를 "어댑터(adapter)"라고 부릅니다. 내부 자료형(leviathan.Court 등)과
// 네트워크로 보내는 자료형(leviathanv1.Court 등, .proto에서 자동 생성)이 서로 다르기 때문에
// 중간에서 한쪽 모양을 다른 쪽 모양으로 옮겨 담는 역할을 합니다.
package server

import (
	"context"

	leviathanv1 "github.com/mkkim-gr8/casepulse/gen/go/leviathan/v1"
	"github.com/mkkim-gr8/casepulse/internal/leviathan"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// server는 gRPC 요청을 Leviathan의 기관 디렉터리 함수로 연결합니다.
type server struct {
	// UnimplementedLeviathanServiceServer를 넣어 두면(임베딩) 아직 구현하지 않은 RPC가 있어도
	// 컴파일이 되고, 그 RPC를 호출하면 "Unimplemented" 오류가 돌아갑니다. gRPC 생성 코드의 관례입니다.
	leviathanv1.UnimplementedLeviathanServiceServer
}

// RegisterServer는 전달받은 gRPC 서버에 Leviathan 서비스를 등록합니다.
func RegisterServer(registrar grpc.ServiceRegistrar) {
	leviathanv1.RegisterLeviathanServiceServer(registrar, server{})
}

// ListCourts는 법원 전체 목록을 Protobuf 응답으로 바꿔 반환합니다.
func (server) ListCourts(context.Context, *leviathanv1.ListCourtsRequest) (*leviathanv1.ListCourtsResponse, error) {
	source := leviathan.Courts()
	// make로 결과를 담을 슬라이스를 원본과 같은 길이로 미리 만든 뒤, 한 칸씩 변환해 채웁니다.
	courts := make([]*leviathanv1.Court, len(source))
	for i, court := range source {
		courts[i] = courtToProto(court)
	}
	return &leviathanv1.ListCourtsResponse{Courts: courts}, nil
}

// GetCourt는 ID로 법원을 찾아 반환하고, 없으면 gRPC NotFound 오류를 돌려줍니다.
func (server) GetCourt(_ context.Context, request *leviathanv1.GetCourtRequest) (*leviathanv1.Court, error) {
	// 법원은 200여 곳뿐이라 처음부터 끝까지 차례로 비교해도 충분히 빠릅니다.
	for _, court := range leviathan.Courts() {
		if court.ID == request.Id {
			return courtToProto(court), nil
		}
	}
	// status.Errorf로 만든 오류는 gRPC 상태 코드(NotFound 등)를 함께 담아 클라이언트에 전달됩니다.
	return nil, status.Errorf(codes.NotFound, "court %q not found", request.Id)
}

// ListProsecutionOffices는 공소청 전체 목록을 Protobuf 응답으로 바꿔 반환합니다.
func (server) ListProsecutionOffices(context.Context, *leviathanv1.ListProsecutionOfficesRequest) (*leviathanv1.ListProsecutionOfficesResponse, error) {
	source := leviathan.ProsecutionOffices()
	offices := make([]*leviathanv1.ProsecutionOffice, len(source))
	for i, office := range source {
		offices[i] = prosecutionOfficeToProto(office)
	}
	return &leviathanv1.ListProsecutionOfficesResponse{Offices: offices}, nil
}

// GetProsecutionOffice는 ID로 공소청을 찾아 반환하고, 없으면 gRPC NotFound 오류를 돌려줍니다.
func (server) GetProsecutionOffice(_ context.Context, request *leviathanv1.GetProsecutionOfficeRequest) (*leviathanv1.ProsecutionOffice, error) {
	for _, office := range leviathan.ProsecutionOffices() {
		if office.ID == request.Id {
			return prosecutionOfficeToProto(office), nil
		}
	}
	return nil, status.Errorf(codes.NotFound, "prosecution office %q not found", request.Id)
}

// ListInvestigationAgencies는 수사기관 전체 목록을 Protobuf 응답으로 바꿔 반환합니다.
func (server) ListInvestigationAgencies(context.Context, *leviathanv1.ListInvestigationAgenciesRequest) (*leviathanv1.ListInvestigationAgenciesResponse, error) {
	source := leviathan.InvestigationAgencies()
	agencies := make([]*leviathanv1.InvestigationAgency, len(source))
	for i, agency := range source {
		agencies[i] = investigationAgencyToProto(agency)
	}
	return &leviathanv1.ListInvestigationAgenciesResponse{Agencies: agencies}, nil
}

// GetInvestigationAgency는 ID로 수사기관을 찾아 반환하고, 없으면 gRPC NotFound 오류를 돌려줍니다.
func (server) GetInvestigationAgency(_ context.Context, request *leviathanv1.GetInvestigationAgencyRequest) (*leviathanv1.InvestigationAgency, error) {
	for _, agency := range leviathan.InvestigationAgencies() {
		if agency.ID == request.Id {
			return investigationAgencyToProto(agency), nil
		}
	}
	return nil, status.Errorf(codes.NotFound, "investigation agency %q not found", request.Id)
}

// courtToProto는 내부 법원 자료형을 gRPC에서 주고받을 자료형으로 바꿉니다.
func courtToProto(court leviathan.Court) *leviathanv1.Court {
	panels := make([]*leviathanv1.CourtPanel, len(court.Panels))
	for i, panel := range court.Panels {
		panels[i] = &leviathanv1.CourtPanel{Id: panel.ID, Name: panel.Name, Seat: panel.Seat}
	}
	return &leviathanv1.Court{
		Id: court.ID, Name: court.Name, InstitutionType: string(court.Type),
		PostalCode: court.PostalCode, Address: court.Address, Phone: court.Phone, Email: court.Email,
		Level: string(court.Level), CourtType: string(court.CourtType), UnitType: string(court.UnitType),
		ParentCourtId: court.ParentCourtID, Seat: court.Seat,
		// append([]string(nil), ...)은 슬라이스 복사본을 만듭니다. 원본을 그대로 넘기면
		// 받는 쪽에서 값을 바꿀 때 원본 데이터까지 바뀔 수 있어서 복사해 넘깁니다.
		Jurisdictions: append([]string(nil), court.Jurisdictions...), Panels: panels,
	}
}

// prosecutionOfficeToProto는 내부 공소청 자료형을 gRPC 응답 자료형으로 바꿉니다.
func prosecutionOfficeToProto(office leviathan.ProsecutionOffice) *leviathanv1.ProsecutionOffice {
	return &leviathanv1.ProsecutionOffice{
		Id: office.ID, Name: office.Name, InstitutionType: string(office.Type),
		PostalCode: office.PostalCode, Address: office.Address, Phone: office.Phone, Email: office.Email,
		Level: string(office.Level), UnitType: string(office.UnitType),
		ParentOfficeId: office.ParentOfficeID, Seat: office.Seat,
	}
}

// investigationAgencyToProto는 내부 수사기관 자료형을 gRPC 응답 자료형으로 바꿉니다.
func investigationAgencyToProto(agency leviathan.InvestigationAgency) *leviathanv1.InvestigationAgency {
	return &leviathanv1.InvestigationAgency{
		Id: agency.ID, Name: agency.Name, InstitutionType: string(agency.Type),
		PostalCode: agency.PostalCode, Address: agency.Address, Phone: agency.Phone, Email: agency.Email,
		ShortName: agency.ShortName, Level: string(agency.Level),
		ParentAgencyId: agency.ParentAgencyID, Seat: agency.Seat,
		Jurisdictions: append([]string(nil), agency.Jurisdictions...),
	}
}
