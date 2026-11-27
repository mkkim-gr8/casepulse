// Package cliogrpc는 웹 서버가 Clio gRPC 서비스를 호출할 때 쓰는 클라이언트를 제공합니다.
package cliogrpc

import (
	"context"

	cliov1 "github.com/mkkim-gr8/casepulse/gen/go/clio/v1"
	"google.golang.org/grpc"
)

// Client는 자동 생성된 gRPC 클라이언트를 감싸 호출 코드를 간단하게 합니다.
type Client struct {
	rpc cliov1.ClioServiceClient
}

// NewClient는 기존 gRPC 연결을 사용해 Clio 클라이언트를 만듭니다.
func NewClient(conn grpc.ClientConnInterface) *Client {
	return &Client{rpc: cliov1.NewClioServiceClient(conn)}
}

// ListCases는 Clio에 저장된 사건과 전체 사건 수를 가져옵니다.
func (c *Client) ListCases(ctx context.Context, opts ...grpc.CallOption) ([]*cliov1.Case, int, error) {
	response, err := c.rpc.ListCases(ctx, &cliov1.ListCasesRequest{}, opts...)
	if err != nil {
		return nil, 0, err
	}
	return response.Cases, int(response.Total), nil
}

// GetCase는 ID로 사건 한 건을 가져옵니다.
func (c *Client) GetCase(ctx context.Context, id string, opts ...grpc.CallOption) (*cliov1.Case, error) {
	return c.rpc.GetCase(ctx, &cliov1.GetCaseRequest{Id: id}, opts...)
}

// CreateCase는 입력한 사건을 Clio에 등록합니다.
func (c *Client) CreateCase(ctx context.Context, input *cliov1.CaseInput, opts ...grpc.CallOption) (*cliov1.Case, error) {
	return c.rpc.CreateCase(ctx, &cliov1.CreateCaseRequest{Input: input}, opts...)
}
