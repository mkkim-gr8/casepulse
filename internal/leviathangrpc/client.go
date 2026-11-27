// Package leviathangrpc는 다른 서비스가 Leviathan의 기관 정보를 조회할 때 쓰는 클라이언트를 제공합니다.
package leviathangrpc

import (
	"context"

	leviathanv1 "github.com/mkkim-gr8/casepulse/gen/go/leviathan/v1"
	"google.golang.org/grpc"
)

// Client는 자동 생성된 gRPC 클라이언트를 감싸 호출 코드를 간단하게 합니다.
type Client struct {
	rpc leviathanv1.LeviathanServiceClient
}

// NewClient는 기존 gRPC 연결을 사용해 Leviathan 클라이언트를 만듭니다.
func NewClient(conn grpc.ClientConnInterface) *Client {
	return &Client{rpc: leviathanv1.NewLeviathanServiceClient(conn)}
}

// ListCourts는 전체 법원과 재판부 목록을 가져옵니다.
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
