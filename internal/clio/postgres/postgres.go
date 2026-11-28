// Package postgres는 Clio의 사건을 PostgreSQL에 저장하는 저장소와 스키마 마이그레이션을 제공합니다.
//
// 테이블은 두 개입니다(migrations/00001_create_cases.sql).
//   - cases: 사건 한 건당 한 행(사건번호, 사건명, 종류, 법원 ID, 재판부 이름)
//   - case_parties: 당사자 한 명당 한 행. case_id로 어느 사건의 당사자인지, position으로 입력 순서를 기억합니다.
//
// 사용하는 라이브러리:
//   - pgx: Go에서 PostgreSQL에 접속하는 드라이버입니다. pgxpool은 연결을 여러 개 만들어 두고 재사용합니다.
//   - goose: 마이그레이션(스키마 변경 SQL 파일)을 번호 순서대로 한 번씩만 적용해 주는 도구입니다.
package postgres

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/mkkim-gr8/casepulse/internal/clio"
	"github.com/pressly/goose/v3"
)

// go:embed 지시문은 빌드할 때 migrations 폴더의 SQL 파일을 실행 파일 안에 넣습니다.
// 그래서 배포할 때 SQL 파일을 따로 복사하지 않아도 됩니다.
//
//go:embed migrations/*.sql
var migrationFiles embed.FS

// uniqueViolation은 PostgreSQL의 고유 제약 위반 오류 코드입니다.
const uniqueViolation = "23505"

// Repository는 pgx 연결 풀로 사건을 저장하고 조회합니다.
type Repository struct {
	pool *pgxpool.Pool
}

// 이 줄은 실행되는 코드가 아니라 컴파일 시간 검사입니다. Repository가 clio.Repository 인터페이스의
// 메서드를 모두 갖추지 못하면 컴파일 오류가 나서 실수를 바로 알 수 있습니다.
var _ clio.Repository = (*Repository)(nil)

// Open은 데이터베이스에 연결하고 접속을 확인합니다. 호출한 쪽에서 Close를 불러야 합니다.
func Open(ctx context.Context, databaseURL string) (*Repository, error) {
	// pgxpool.New는 URL을 해석해 연결 풀을 만들지만 실제 접속은 필요할 때 합니다.
	// 그래서 바로 Ping을 보내 DB에 정말 접속되는지 확인합니다.
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database URL: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("connect to database: %w", err)
	}
	return &Repository{pool: pool}, nil
}

// Close는 연결 풀을 닫습니다.
func (r *Repository) Close() {
	r.pool.Close()
}

// Migrate는 실행 파일에 포함된 마이그레이션 중 아직 적용하지 않은 것을 순서대로 적용하고,
// 새로 적용한 마이그레이션 수를 돌려줍니다.
func (r *Repository) Migrate(ctx context.Context) (int, error) {
	migrations, err := fs.Sub(migrationFiles, "migrations")
	if err != nil {
		return 0, err
	}
	// goose는 Go 표준 database/sql 방식의 연결을 쓰므로, pgx 풀을 그 형식으로 바꿔 넘깁니다.
	db := stdlib.OpenDBFromPool(r.pool)
	defer db.Close()
	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations)
	if err != nil {
		return 0, fmt.Errorf("prepare migrations: %w", err)
	}
	// Up은 DB의 goose_db_version 테이블을 보고 아직 적용하지 않은 파일만 차례로 실행합니다.
	results, err := provider.Up(ctx)
	if err != nil {
		return 0, fmt.Errorf("apply migrations: %w", err)
	}
	return len(results), nil
}

// List는 최근 등록한 사건부터, 같은 때 등록한 사건은 사건번호가 큰 순서로 돌려줍니다.
func (r *Repository) List(ctx context.Context) ([]clio.StoredCase, error) {
	// id::text는 UUID 값을 문자열로 바꿔 읽는다는 PostgreSQL 문법입니다.
	rows, err := r.pool.Query(ctx, `
		SELECT id::text, year, code, serial, title, category, court_id, panel_name
		FROM cases
		ORDER BY created_at DESC, year DESC, serial DESC`)
	if err != nil {
		return nil, fmt.Errorf("list cases: %w", err)
	}
	// CollectRows는 결과의 모든 행을 scanCase로 하나씩 읽어 슬라이스로 모읍니다.
	items, err := pgx.CollectRows(rows, scanCase)
	if err != nil {
		return nil, fmt.Errorf("list cases: %w", err)
	}
	// 사건마다 당사자를 따로 조회하면 사건 수만큼 쿼리가 나가므로(N+1 문제),
	// 당사자는 한 번의 쿼리로 모두 읽어 사건에 나눠 붙입니다.
	if err := r.loadParties(ctx, items, nil); err != nil {
		return nil, err
	}
	return items, nil
}

// Get은 ID로 사건 한 건을 찾습니다. ID가 UUID 형식이 아니어도 ErrCaseNotFound를 돌려줍니다.
func (r *Repository) Get(ctx context.Context, id string) (clio.StoredCase, error) {
	// 먼저 문자열 id가 UUID 형식인지 확인합니다. 형식이 틀린 값을 그대로 DB에 보내면
	// "없음"이 아니라 DB 오류가 나기 때문에, 여기서 바로 "찾을 수 없음"으로 처리합니다.
	var uuid pgtype.UUID
	if err := uuid.Scan(id); err != nil {
		return clio.StoredCase{}, clio.ErrCaseNotFound
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id::text, year, code, serial, title, category, court_id, panel_name
		FROM cases
		WHERE id = $1`, uuid) // $1 자리에 uuid 값이 안전하게 들어갑니다(SQL 인젝션 방지).
	if err != nil {
		return clio.StoredCase{}, fmt.Errorf("get case: %w", err)
	}
	item, err := pgx.CollectExactlyOneRow(rows, scanCase)
	if errors.Is(err, pgx.ErrNoRows) {
		return clio.StoredCase{}, clio.ErrCaseNotFound
	}
	if err != nil {
		return clio.StoredCase{}, fmt.Errorf("get case: %w", err)
	}
	items := []clio.StoredCase{item}
	if err := r.loadParties(ctx, items, &uuid); err != nil {
		return clio.StoredCase{}, err
	}
	return items[0], nil
}

// Insert는 사건과 당사자를 한 트랜잭션으로 저장합니다.
func (r *Repository) Insert(ctx context.Context, cases ...clio.Case) ([]clio.StoredCase, error) {
	stored := make([]clio.StoredCase, len(cases))
	// BeginFunc는 트랜잭션을 열고 안의 함수를 실행합니다. 함수가 nil을 돌려주면 커밋(모두 저장),
	// 오류를 돌려주면 롤백(아무것도 저장하지 않음)합니다. 그래서 사건 여러 건 중 하나만 실패해도 전부 취소됩니다.
	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		for i, c := range cases {
			var id string
			err := tx.QueryRow(ctx, `
				INSERT INTO cases (year, code, serial, title, category, court_id, panel_name)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
				RETURNING id::text`, // DB가 만든 새 사건 ID(uuidv7)를 돌려받습니다.
				c.Number.Year, c.Number.Code, c.Number.Serial, c.Title, string(c.Category),
				c.Court.CourtID, c.Court.PanelName,
			).Scan(&id)
			if err != nil {
				return err
			}
			// 당사자 행을 [case_id, position, role, name, client] 모양으로 모은 뒤,
			// CopyFrom(PostgreSQL COPY)으로 한 번에 넣습니다. INSERT를 여러 번 하는 것보다 빠릅니다.
			parties := make([][]any, len(c.Parties))
			for j, p := range c.Parties {
				parties[j] = []any{id, j, p.Role, p.Name, p.Client}
			}
			if _, err := tx.CopyFrom(ctx, pgx.Identifier{"case_parties"},
				[]string{"case_id", "position", "role", "name", "client"},
				pgx.CopyFromRows(parties)); err != nil {
				return err
			}
			c.Parties = append([]clio.Party(nil), c.Parties...)
			stored[i] = clio.StoredCase{ID: id, Case: c}
		}
		return nil
	})
	// errors.As는 err 안에 PostgreSQL 오류(*pgconn.PgError)가 들어 있으면 pgErr에 꺼내 줍니다.
	// 고유 제약(같은 법원·같은 사건번호) 위반이면 Clio가 이해하는 ErrDuplicateCase로 바꿉니다.
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
		return nil, clio.ErrDuplicateCase
	}
	if err != nil {
		return nil, fmt.Errorf("insert cases: %w", err)
	}
	return stored, nil
}

// Count는 저장된 사건 수를 돌려줍니다.
func (r *Repository) Count(ctx context.Context) (int, error) {
	var count int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM cases`).Scan(&count); err != nil {
		return 0, fmt.Errorf("count cases: %w", err)
	}
	return count, nil
}

// loadParties는 사건들의 당사자를 입력 순서대로 채웁니다. only가 nil이면 모든 사건의 당사자를 읽습니다.
func (r *Repository) loadParties(ctx context.Context, items []clio.StoredCase, only *pgtype.UUID) error {
	if len(items) == 0 {
		return nil
	}
	// index: 사건 ID → items 안의 위치. 당사자 행을 읽을 때 어느 사건에 붙일지 바로 찾기 위해 씁니다.
	index := make(map[string]int, len(items))
	for i, item := range items {
		index[item.ID] = i
	}
	query := `SELECT case_id::text, role, name, client FROM case_parties ORDER BY case_id, position`
	args := []any{}
	if only != nil {
		query = `SELECT case_id::text, role, name, client FROM case_parties
			WHERE case_id = $1 ORDER BY position`
		args = append(args, *only)
	}
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("load parties: %w", err)
	}
	defer rows.Close()
	// rows.Next()는 다음 행이 있으면 true를 돌려주며, 반복할 때마다 한 행씩 읽습니다.
	for rows.Next() {
		var caseID string
		var p clio.Party
		if err := rows.Scan(&caseID, &p.Role, &p.Name, &p.Client); err != nil {
			return fmt.Errorf("load parties: %w", err)
		}
		if i, ok := index[caseID]; ok {
			items[i].Case.Parties = append(items[i].Case.Parties, p)
		}
	}
	// 반복 도중 생긴 오류(연결 끊김 등)는 반복이 끝난 뒤 rows.Err()로 확인해야 합니다.
	if err := rows.Err(); err != nil {
		return fmt.Errorf("load parties: %w", err)
	}
	return nil
}

// scanCase는 cases 테이블의 한 행을 StoredCase로 읽습니다.
func scanCase(row pgx.CollectableRow) (clio.StoredCase, error) {
	var item clio.StoredCase
	var category string
	// Scan은 SELECT에 적은 열 순서대로 값을 각 변수의 주소(&)에 채워 넣습니다.
	err := row.Scan(&item.ID, &item.Case.Number.Year, &item.Case.Number.Code, &item.Case.Number.Serial,
		&item.Case.Title, &category, &item.Case.Court.CourtID, &item.Case.Court.PanelName)
	item.Case.Category = clio.CaseCategory(category)
	return item, err
}
