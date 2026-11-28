// Package leviathangrpc는 다른 서비스가 Leviathan의 기관 정보를 조회할 때 쓰는 클라이언트를 제공합니다.
//
// Clio와 Aegis는 이 패키지를 통해서만 Leviathan에 요청합니다. 즉 Leviathan 내부 코드(internal/leviathan)를
// 직접 import하지 않고, 네트워크(gRPC)로 "법원 목록 주세요"라고 물어봅니다.
// 이렇게 하면 서비스끼리 서로의 내부 구현에 묶이지 않아 따로 고치고 배포할 수 있습니다.
//
// 사용 예:
//
//	conn, _ := grpc.NewClient("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
//	client := leviathangrpc.NewClient(conn)
//	courts, err := client.ListCourts(ctx)
package leviathangrpc

import (
	"context"

	leviathanv1 "github.com/mkkim-gr8/casepulse/gen/go/leviathan/v1"
	"google.golang.org/grpc"
)

// Client는 자동 생성된 gRPC 클라이언트를 감싸 호출 코드를 간단하게 합니다.
// 생성된 클라이언트(gen/go/leviathan/v1)는 매번 요청 메시지를 만들고 응답에서 필드를 꺼내야 하므로,
// 여기서 그 과정을 함수 하나로 묶어 둡니다.
type Client struct {
	rpc leviathanv1.LeviathanServiceClient
}

// NewClient는 기존 gRPC 연결을 사용해 Leviathan 클라이언트를 만듭니다.
func NewClient(conn grpc.ClientConnInterface) *Client {
	return &Client{rpc: leviathanv1.NewLeviathanServiceClient(conn)}
}

// ListCourts는 전체 법원과 재판부 목록을 가져옵니다.
// opts에는 grpc.WaitForReady(true)처럼 호출 방식을 바꾸는 옵션을 넘길 수 있습니다(...은 0개 이상 가변 인자).
func (c *Client) ListCourts(ctx context.Context, opts ...grpc.CallOption) ([]*leviathanv1.Court, error) {
	response, err := c.rpc.ListCourts(ctx, &leviathanv1.ListCourtsRequest{}, opts...)
	if err != nil {
		return nil, err
	}
	return response.Courts, nil
}

// GetCourt는 ID로 법원 한 곳을 가져옵니다.
func (c *Client) GetCourt(ctx context.Context, id string, opts ...grpc.CallOption) (*leviathanv1.Court, error) {
	response, err := c.rpc.GetCourt(ctx, &leviathanv1.GetCourtRequest{Id: id}, opts...)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// ListProsecutionOffices는 전체 공소청 목록을 가져옵니다.
func (c *Client) ListProsecutionOffices(ctx context.Context, opts ...grpc.CallOption) ([]*leviathanv1.ProsecutionOffice, error) {
	response, err := c.rpc.ListProsecutionOffices(ctx, &leviathanv1.ListProsecutionOfficesRequest{}, opts...)
	if err != nil {
		return nil, err
	}
	return response.Offices, nil
}

// GetProsecutionOffice는 ID로 공소청 한 곳을 가져옵니다.
func (c *Client) GetProsecutionOffice(ctx context.Context, id string, opts ...grpc.CallOption) (*leviathanv1.ProsecutionOffice, error) {
	response, err := c.rpc.GetProsecutionOffice(ctx, &leviathanv1.GetProsecutionOfficeRequest{Id: id}, opts...)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// ListInvestigationAgencies는 전체 중대범죄수사청 기관 목록을 가져옵니다.
func (c *Client) ListInvestigationAgencies(ctx context.Context, opts ...grpc.CallOption) ([]*leviathanv1.InvestigationAgency, error) {
	response, err := c.rpc.ListInvestigationAgencies(ctx, &leviathanv1.ListInvestigationAgenciesRequest{}, opts...)
	if err != nil {
		return nil, err
	}
	return response.Agencies, nil
}

// GetInvestigationAgency는 ID로 수사기관 한 곳을 가져옵니다.
func (c *Client) GetInvestigationAgency(ctx context.Context, id string, opts ...grpc.CallOption) (*leviathanv1.InvestigationAgency, error) {
	response, err := c.rpc.GetInvestigationAgency(ctx, &leviathanv1.GetInvestigationAgencyRequest{Id: id}, opts...)
	if err != nil {
		return nil, err
	}
	return response, nil
}
