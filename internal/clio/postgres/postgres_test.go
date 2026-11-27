package postgres_test

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/mkkim-gr8/casepulse/internal/clio"
	"github.com/mkkim-gr8/casepulse/internal/clio/postgres"
	"github.com/mkkim-gr8/casepulse/internal/clio/repotest"
)

// testDatabaseEnv는 통합 테스트에 쓸 PostgreSQL 주소를 담는 환경변수입니다.
// 이 주소의 사용자는 테스트마다 임시 DB를 만들고 지울 수 있어야 합니다(CREATEDB 권한).
const testDatabaseEnv = "CLIO_TEST_DATABASE_URL"

var databaseSeq atomic.Int64

// newTestRepository는 비어 있는 임시 DB를 만들고 마이그레이션을 적용한 저장소를 돌려줍니다.
// 테스트가 끝나면 임시 DB를 지웁니다.
func newTestRepository(t *testing.T) *postgres.Repository {
	t.Helper()
	adminURL := os.Getenv(testDatabaseEnv)
	if adminURL == "" {
		t.Skipf("%s is not set; skipping PostgreSQL integration test", testDatabaseEnv)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	admin, err := pgx.Connect(ctx, adminURL)
	if err != nil {
		t.Fatalf("connect to %s: %v", testDatabaseEnv, err)
	}
	defer admin.Close(context.Background())
	name := fmt.Sprintf("clio_test_%d_%d", os.Getpid(), databaseSeq.Add(1))
	if _, err := admin.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{name}.Sanitize()); err != nil {
		t.Fatalf("create test database: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		conn, err := pgx.Connect(ctx, adminURL)
		if err != nil {
			t.Errorf("connect to drop test database: %v", err)
			return
		}
		defer conn.Close(ctx)
		if _, err := conn.Exec(ctx, "DROP DATABASE IF EXISTS "+pgx.Identifier{name}.Sanitize()+" WITH (FORCE)"); err != nil {
			t.Errorf("drop test database: %v", err)
		}
	})

	testURL, err := url.Parse(adminURL)
	if err != nil {
		t.Fatal(err)
	}
	testURL.Path = "/" + name
	repo, err := postgres.Open(ctx, testURL.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(repo.Close)
	if applied, err := repo.Migrate(ctx); err != nil || applied == 0 {
		t.Fatalf("Migrate() = %d, %v", applied, err)
	}
	return repo
}

func TestRepository(t *testing.T) {
	repotest.Run(t, func(t *testing.T) clio.Repository { return newTestRepository(t) })
}

func TestMigrateIsIdempotent(t *testing.T) {
	repo := newTestRepository(t)
	applied, err := repo.Migrate(context.Background())
	if err != nil || applied != 0 {
		t.Fatalf("second Migrate() = %d, %v; want 0", applied, err)
	}
}
