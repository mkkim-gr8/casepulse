// clio는 사건을 보관하고 사건 목록·조회·등록 gRPC 요청을 처리하는 프로그램입니다.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	leviathanv1 "github.com/mkkim-gr8/casepulse/gen/go/leviathan/v1"
	"github.com/mkkim-gr8/casepulse/internal/clio"
	cliogrpcserver "github.com/mkkim-gr8/casepulse/internal/cliogrpc/server"
	"github.com/mkkim-gr8/casepulse/internal/healthcheck"
	"github.com/mkkim-gr8/casepulse/internal/leviathangrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// main은 실행 옵션을 읽고 Clio 서버를 시작합니다.
func main() {
	addr := flag.String("addr", ":9091", "gRPC listen address")
	leviathanAddr := flag.String("leviathan-addr", envOrDefault("LEVIATHAN_GRPC_ADDR", "localhost:9090"), "Leviathan gRPC address")
	probe := flag.Bool("healthcheck", false, "check the running service at -addr and exit")
	flag.Parse()

	if *probe {
		if err := healthcheck.ProbeGRPC(*addr); err != nil {
			log.Printf("healthcheck failed: %v", err)
			os.Exit(1)
		}
		return
	}
	if err := run(*addr, *leviathanAddr); err != nil {
		log.Fatal(err)
	}
}

// run은 Leviathan에서 법원 목록을 받은 뒤 더미 사건을 준비하고 gRPC 서버를 엽니다.
func run(addr, leviathanAddr string) error {
	conn, err := grpc.NewClient(leviathanAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("create Leviathan gRPC client: %w", err)
	}
	defer conn.Close()

	leviathanClient := leviathangrpc.NewClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	source, err := leviathanClient.ListCourts(ctx, grpc.WaitForReady(true))
	if err != nil {
		return fmt.Errorf("load courts from Leviathan at %q: %w", leviathanAddr, err)
	}
	courts, err := courtsFromProto(source)
	if err != nil {
		return err
	}
	if len(courts) == 0 {
		return fmt.Errorf("Leviathan gRPC service at %q returned no courts", leviathanAddr)
	}
	store := clio.NewStore(clio.BuildDummyCases(courts), courts)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}
	server := grpc.NewServer(grpc.UnaryInterceptor(logGRPCRequest))
	cliogrpcserver.RegisterServer(server, store)
	healthcheck.RegisterGRPC(server)
	log.Printf("clio gRPC service listening on %s", addr)
	if err := server.Serve(listener); err != nil {
		return fmt.Errorf("serve clio gRPC: %w", err)
	}
	return nil
}

// courtsFromProto는 Leviathan 응답을 Clio가 검증에 쓰는 간단한 법원 자료형으로 바꿉니다.
func courtsFromProto(source []*leviathanv1.Court) ([]clio.Court, error) {
	courts := make([]clio.Court, len(source))
	for i, court := range source {
		if court == nil {
			return nil, fmt.Errorf("Leviathan returned an empty court at index %d", i)
		}
		panels := make([]clio.CourtPanel, len(court.Panels))
		for j, panel := range court.Panels {
			if panel == nil {
				return nil, fmt.Errorf("Leviathan returned an empty panel for court %q at index %d", court.Id, j)
			}
			panels[j] = clio.CourtPanel{ID: panel.Id, Name: panel.Name}
		}
		courts[i] = clio.Court{
			ID: court.Id, Name: court.Name, Seat: court.Seat,
			PostalCode: court.PostalCode, Address: court.Address,
			Level: court.Level, CourtType: court.CourtType, UnitType: court.UnitType,
			Panels: panels,
		}
	}
	return courts, nil
}

// logGRPCRequest는 요청 메서드만 기록해 사건 개인정보가 로그에 남지 않게 합니다.
// 주기적으로 반복되는 health 확인 요청은 기록하지 않습니다.
func logGRPCRequest(ctx context.Context, request any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if healthcheck.IsHealthMethod(info.FullMethod) {
		return handler(ctx, request)
	}
	log.Printf("gRPC request: method=%s", info.FullMethod)
	return handler(ctx, request)
}

// envOrDefault는 환경변수가 비어 있을 때 사용할 기본 주소를 돌려줍니다.
func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
