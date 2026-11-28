// leviathan은 법원·공소청·중대범죄수사청 정보를 gRPC로 제공하는 프로그램입니다.
//
// gRPC는 다른 프로그램(서비스)의 함수를 네트워크 너머에서 호출할 수 있게 해 주는 통신 방식입니다.
// 어떤 함수(RPC)가 있고 무엇을 주고받는지는 api/proto/casepulse/leviathan/v1/leviathan.proto에 정의되어 있습니다.
//
// 실행 방법(저장소 루트에서):
//
//	go run ./cmd/leviathan              # :9090 포트에서 서버 시작
//	go run ./cmd/leviathan -healthcheck # 이미 떠 있는 서버가 정상인지 한 번 확인하고 종료
package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"

	"github.com/mkkim-gr8/casepulse/internal/healthcheck"
	leviathangrpcserver "github.com/mkkim-gr8/casepulse/internal/leviathangrpc/server"
	"google.golang.org/grpc"
)

// main은 gRPC 서버를 지정한 포트에서 시작합니다.
// Go 프로그램은 main 패키지의 main 함수부터 실행됩니다.
func main() {
	// flag 패키지는 "-addr :9000"처럼 명령줄에서 넘긴 옵션을 읽습니다.
	// flag.String/flag.Bool은 값 자체가 아니라 값을 담을 변수의 포인터를 돌려주므로,
	// 실제 값을 쓸 때는 *addr처럼 앞에 *를 붙입니다.
	addr := flag.String("addr", ":9090", "gRPC listen address")
	probe := flag.Bool("healthcheck", false, "check the running service at -addr and exit")
	flag.Parse() // 실제로 명령줄을 읽어 위 변수들에 값을 채웁니다.

	// -healthcheck 옵션이 있으면 서버를 켜지 않고, 이미 실행 중인 서버의 상태만 확인합니다.
	// Docker Compose의 healthcheck가 이 방식으로 컨테이너가 정상인지 검사합니다.
	if *probe {
		if err := healthcheck.ProbeGRPC(*addr); err != nil {
			log.Printf("healthcheck failed: %v", err)
			os.Exit(1) // 종료 코드 1 = 실패. Docker는 이를 "unhealthy"로 판단합니다.
		}
		return // 종료 코드 0 = 성공
	}

	// net.Listen은 지정한 주소(예: ":9090" = 모든 네트워크 카드의 9090 포트)에서
	// TCP 연결을 받을 준비를 합니다. 포트가 이미 사용 중이면 오류가 납니다.
	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		// log.Fatalf는 메시지를 출력한 뒤 프로그램을 바로 종료합니다.
		log.Fatalf("listen on %s: %v", *addr, err)
	}

	// gRPC 서버를 만듭니다. UnaryInterceptor는 모든 요청 앞에 끼워 넣는 "중간 처리 함수"로,
	// 여기서는 요청 로그를 남기는 logGRPCRequest를 등록합니다.
	server := grpc.NewServer(grpc.UnaryInterceptor(logGRPCRequest))
	leviathangrpcserver.RegisterServer(server) // 기관 조회 RPC들을 서버에 등록
	healthcheck.RegisterGRPC(server)           // 표준 상태 확인(health) RPC 등록
	log.Printf("leviathan gRPC service listening on %s", *addr)
	// Serve는 요청을 계속 받아 처리하며, 서버가 멈출 때까지 돌아오지 않습니다.
	if err := server.Serve(listener); err != nil {
		log.Fatalf("serve leviathan gRPC: %v", err)
	}
}

// logGRPCRequest는 들어온 메서드와 요청 내용을 터미널에 출력한 뒤 실제 처리를 이어갑니다.
// 주기적으로 반복되는 health 확인 요청은 기록하지 않습니다.
//
// 매개변수 설명:
//   - ctx: 요청의 취소·시간 제한 정보를 담은 값입니다(Go의 context).
//   - request: 클라이언트가 보낸 요청 메시지입니다.
//   - info.FullMethod: 호출된 RPC 이름입니다(예: "/casepulse.leviathan.v1.LeviathanService/ListCourts").
//   - handler: 실제 RPC를 처리하는 함수입니다. 이를 호출해야 요청이 처리됩니다.
func logGRPCRequest(ctx context.Context, request any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if healthcheck.IsHealthMethod(info.FullMethod) {
		return handler(ctx, request)
	}
	log.Printf("gRPC request: method=%s request=%v", info.FullMethod, request)
	return handler(ctx, request)
}
