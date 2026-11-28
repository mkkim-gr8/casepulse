// Package healthcheck는 서비스 상태 확인용 gRPC health 서비스 등록과
// 컨테이너 healthcheck에서 쓰는 점검 함수를 제공합니다.
//
// "health check(상태 확인)"란 서비스가 살아서 요청을 받을 수 있는지 묻는 작은 요청입니다.
// Docker Compose는 이 결과를 보고 컨테이너를 healthy/unhealthy로 표시하고,
// 다른 서비스는 상대가 healthy가 된 뒤에 시작하도록 순서를 맞춥니다.
//
// gRPC에는 표준 health 서비스(grpc.health.v1.Health)가 있어서 직접 만들지 않고 그대로 씁니다.
// HTTP 서버(Aegis)는 GET /healthz에 200으로 응답하는지로 확인합니다.
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
	// 서비스 이름 ""은 "서버 전체"를 뜻합니다. 서버가 켜져 있는 동안 항상 SERVING(정상)으로 답합니다.
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
	// grpc.NewClient는 서버와의 연결 객체를 만듭니다. insecure는 TLS(암호화) 없이 접속한다는 뜻으로,
	// 같은 컨테이너 안이나 내부 네트워크에서만 쓰기 때문에 괜찮습니다.
	conn, err := grpc.NewClient(LocalAddr(listenAddr), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close() // defer: 함수가 끝날 때 연결을 닫도록 예약합니다.
	// 응답이 ProbeTimeout(3초) 안에 오지 않으면 실패로 처리하도록 시간 제한을 겁니다.
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
	response, err := http.DefaultClient.Do(request) // 실제로 HTTP 요청을 보냅니다.
	if err != nil {
		return err
	}
	defer response.Body.Close() // 응답 본문은 다 쓴 뒤 반드시 닫아야 연결 자원이 정리됩니다.
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", response.StatusCode)
	}
	return nil
}

// LocalAddr는 ":9090"이나 "0.0.0.0:9090" 같은 listen 주소를 같은 컨테이너 안에서
// 접속할 수 있는 "localhost:9090" 형식으로 바꿉니다.
func LocalAddr(listenAddr string) string {
	// "0.0.0.0:9090"을 host="0.0.0.0", port="9090"으로 나눕니다.
	host, port, err := net.SplitHostPort(listenAddr)
	if err != nil {
		return listenAddr
	}
	// ""·"0.0.0.0"·"::"는 "모든 주소에서 받기"라는 뜻이라 접속 대상으로는 쓸 수 없으므로 localhost로 바꿉니다.
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "localhost"
	}
	return net.JoinHostPort(host, port)
}
