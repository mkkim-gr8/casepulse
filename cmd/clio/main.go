// clio는 사건을 보관하고 사건 목록·조회·등록 gRPC 요청을 처리하는 프로그램입니다.
//
// 시작 순서:
//  1. Leviathan에 접속해 법원·재판부 목록을 받습니다(사건의 법원 검증용).
//  2. 사건 저장소를 엽니다. DATABASE_URL이 있으면 PostgreSQL, 없으면 메모리.
//  3. 저장소가 비어 있으면 더미 사건 240건을 넣습니다.
//  4. :9091 포트에서 gRPC 서버를 엽니다.
//
// 실행 예(저장소 루트에서, Leviathan을 먼저 켜 둔 상태):
//
//	go run ./cmd/clio
//	DATABASE_URL='postgres://casepulse:casepulse@localhost:5432/casepulse?sslmode=disable' go run ./cmd/clio
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
	cliopostgres "github.com/mkkim-gr8/casepulse/internal/clio/postgres"
	cliogrpcserver "github.com/mkkim-gr8/casepulse/internal/cliogrpc/server"
	"github.com/mkkim-gr8/casepulse/internal/healthcheck"
	"github.com/mkkim-gr8/casepulse/internal/leviathangrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// main은 실행 옵션을 읽고 Clio 서버를 시작합니다.
func main() {
	// 각 옵션의 기본값은 환경변수에서 먼저 찾고, 없으면 코드에 적힌 값을 씁니다.
	// Docker Compose는 환경변수로 설정을 넘기고, 로컬에서는 -옵션으로도 바꿀 수 있습니다.
	addr := flag.String("addr", ":9091", "gRPC listen address")
	leviathanAddr := flag.String("leviathan-addr", envOrDefault("LEVIATHAN_GRPC_ADDR", "localhost:9090"), "Leviathan gRPC address")
	databaseURL := flag.String("database-url", os.Getenv("DATABASE_URL"), "PostgreSQL URL; empty keeps cases in memory")
	seedDummy := flag.Bool("seed-dummy", envOrDefault("CLIO_SEED_DUMMY", "true") == "true", "add dummy cases when the store is empty")
	probe := flag.Bool("healthcheck", false, "check the running service at -addr and exit")
	flag.Parse()

	if *probe {
		if err := healthcheck.ProbeGRPC(*addr); err != nil {
			log.Printf("healthcheck failed: %v", err)
			os.Exit(1)
		}
		return
	}
	// 실제 시작 과정은 run에 두고 오류를 돌려받습니다. 이렇게 하면 run 안의 defer(정리 작업)가
	// 모두 실행된 뒤에 프로그램이 종료됩니다(log.Fatal은 defer를 실행하지 않고 바로 끝내기 때문).
	if err := run(*addr, *leviathanAddr, *databaseURL, *seedDummy); err != nil {
		log.Fatal(err)
	}
}

// run은 Leviathan에서 법원 목록을 받고 사건 저장소를 준비한 뒤 gRPC 서버를 엽니다.
func run(addr, leviathanAddr, databaseURL string, seedDummy bool) error {
	conn, err := grpc.NewClient(leviathanAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("create Leviathan gRPC client: %w", err)
	}
	defer conn.Close()

	leviathanClient := leviathangrpc.NewClient(conn)
	// 시작 준비(법원 목록 받기, DB 접속, 더미 넣기)에 최대 30초를 허용합니다.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// WaitForReady(true): Leviathan이 아직 켜지는 중이면 바로 실패하지 않고 연결될 때까지 기다립니다.
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

	// openRepository는 저장소와 함께 "닫는 함수"를 돌려줍니다. 프로그램이 끝날 때 defer로 호출합니다.
	repo, closeRepo, err := openRepository(ctx, databaseURL)
	if err != nil {
		return err
	}
	defer closeRepo()
	store := clio.NewStore(repo, courts)
	if seedDummy {
		seeded, err := store.SeedIfEmpty(ctx, clio.BuildDummyCases(courts))
		if err != nil {
			return fmt.Errorf("seed dummy cases: %w", err)
		}
		if seeded > 0 {
			log.Printf("seeded %d dummy cases", seeded)
		}
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}
	server := grpc.NewServer(grpc.UnaryInterceptor(logGRPCRequest))
	cliogrpcserver.RegisterServer(server, store) // 사건 RPC(ListCases, GetCase, CreateCase) 등록
	healthcheck.RegisterGRPC(server)
	log.Printf("clio gRPC service listening on %s", addr)
	if err := server.Serve(listener); err != nil {
		return fmt.Errorf("serve clio gRPC: %w", err)
	}
	return nil
}

// openRepository는 databaseURL이 있으면 PostgreSQL에 연결해 마이그레이션을 적용하고,
// 없으면 메모리 저장소를 씁니다. DB가 아직 준비되지 않았으면 ctx가 끝날 때까지 다시 시도합니다.
func openRepository(ctx context.Context, databaseURL string) (clio.Repository, func(), error) {
	if databaseURL == "" {
		log.Printf("DATABASE_URL is empty; cases are kept in memory and lost on restart")
		// 메모리 저장소는 닫을 것이 없으므로 아무 일도 하지 않는 함수를 돌려줍니다.
		return clio.NewMemoryRepository(), func() {}, nil
	}
	// Compose에서 PostgreSQL과 Clio가 거의 동시에 켜질 수 있으므로, 접속에 실패하면 1초 쉬고 다시 시도합니다.
	var repo *cliopostgres.Repository
	for {
		var err error
		repo, err = cliopostgres.Open(ctx, databaseURL)
		if err == nil {
			break
		}
		// select는 여러 채널 중 먼저 신호가 오는 쪽을 실행합니다.
		//   - ctx.Done(): 30초 제한이 끝나면 신호가 옵니다 → 포기하고 오류를 돌려줍니다.
		//   - time.After(1초): 1초 뒤 신호가 옵니다 → 반복문 처음으로 돌아가 다시 접속합니다.
		select {
		case <-ctx.Done():
			return nil, nil, fmt.Errorf("open PostgreSQL: %w", err)
		case <-time.After(time.Second):
		}
	}
	// 스키마(테이블)가 최신 상태가 되도록 아직 적용하지 않은 마이그레이션을 실행합니다.
	applied, err := repo.Migrate(ctx)
	if err != nil {
		repo.Close()
		return nil, nil, err
	}
	log.Printf("connected to PostgreSQL; applied %d new migrations", applied)
	// repo.Close처럼 괄호 없이 쓰면 메서드를 "호출"하지 않고 함수 값으로 넘깁니다.
	return repo, repo.Close, nil
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
// (Leviathan은 공개 기관 정보만 다루므로 요청 내용까지 기록하지만, Clio의 요청에는 당사자 이름이 들어 있습니다.)
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
