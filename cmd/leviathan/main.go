// leviathan은 법원·공소청·중대범죄수사청 정보를 gRPC로 제공하는 프로그램입니다.
package main

import (
	"context"
	"flag"
	"log"
	"net"

	leviathangrpcserver "github.com/mkkim-gr8/casepulse/internal/leviathangrpc/server"
	"google.golang.org/grpc"
)

// main은 gRPC 서버를 지정한 포트에서 시작합니다.
func main() {
	addr := flag.String("addr", ":9090", "gRPC listen address")
	flag.Parse()

	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("listen on %s: %v", *addr, err)
	}

	server := grpc.NewServer(grpc.UnaryInterceptor(logGRPCRequest))
	leviathangrpcserver.RegisterServer(server)
	log.Printf("leviathan gRPC service listening on %s", *addr)
	if err := server.Serve(listener); err != nil {
		log.Fatalf("serve leviathan gRPC: %v", err)
	}
}

// logGRPCRequest는 들어온 메서드와 요청 내용을 터미널에 출력한 뒤 실제 처리를 이어갑니다.
func logGRPCRequest(ctx context.Context, request any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	log.Printf("gRPC request: method=%s request=%v", info.FullMethod, request)
	return handler(ctx, request)
}
