// Package healthcheck는 서비스 상태 확인용 gRPC health 서비스 등록과
// 컨테이너 healthcheck에서 쓰는 점검 함수를 제공합니다.
package healthcheck

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// ProbeTimeout은 -healthcheck 점검 한 번에 허용하는 최대 시간입니다.
const ProbeTimeout = 3 * time.Second

// RegisterGRPC는 gRPC 서버에 표준 health 서비스를 등록하고 SERVING 상태로 둡니다.
func RegisterGRPC(registrar grpc.ServiceRegistrar) {
	server := health.NewServer()
	server.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(registrar, server)
}

// IsHealthMethod는 요청 로그에서 반복되는 상태 확인 호출을 빼기 위해 health RPC인지 알려 줍니다.
func IsHealthMethod(fullMethod string) bool {
	return fullMethod == healthpb.Health_Check_FullMethodName ||
		fullMethod == healthpb.Health_Watch_FullMethodName
}

// CheckConn은 이미 만든 gRPC 연결로 상대 서비스가 SERVING 상태인지 확인합니다.
// 시작 단계에서는 grpc.WaitForReady(true)를 넘겨 상대가 뜰 때까지 기다릴 수 있습니다.
func CheckConn(ctx context.Context, conn grpc.ClientConnInterface, opts ...grpc.CallOption) error {
	response, err := healthpb.NewHealthClient(conn).Check(ctx, &healthpb.HealthCheckRequest{}, opts...)
	if err != nil {
		return err
	}
	if response.Status != healthpb.HealthCheckResponse_SERVING {
		return fmt.Errorf("service status is %s", response.Status)
	}
	return nil
}

// ProbeGRPC는 listen 주소의 gRPC 서비스가 SERVING 상태인지 한 번 확인합니다.
func ProbeGRPC(listenAddr string) error {
	conn, err := grpc.NewClient(LocalAddr(listenAddr), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), ProbeTimeout)
	defer cancel()
	return CheckConn(ctx, conn)
}

// ProbeHTTP는 listen 주소의 HTTP 서버가 path에 200으로 응답하는지 한 번 확인합니다.
func ProbeHTTP(listenAddr, path string) error {
	ctx, cancel := context.WithTimeout(context.Background(), ProbeTimeout)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+LocalAddr(listenAddr)+path, nil)
	if err != nil {
		return err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", response.StatusCode)
	}
	return nil
}

// LocalAddr는 ":9090"이나 "0.0.0.0:9090" 같은 listen 주소를 같은 컨테이너 안에서
// 접속할 수 있는 "localhost:9090" 형식으로 바꿉니다.
func LocalAddr(listenAddr string) string {
	host, port, err := net.SplitHostPort(listenAddr)
	if err != nil {
		return listenAddr
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "localhost"
	}
	return net.JoinHostPort(host, port)
}
