// Package postgres는 Clio의 사건을 PostgreSQL에 저장하는 저장소와 스키마 마이그레이션을 제공합니다.
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

//go:embed migrations/*.sql
var migrationFiles embed.FS

// uniqueViolation은 PostgreSQL의 고유 제약 위반 오류 코드입니다.
const uniqueViolation = "23505"

// Repository는 pgx 연결 풀로 사건을 저장하고 조회합니다.
type Repository struct {
	pool *pgxpool.Pool
}

var _ clio.Repository = (*Repository)(nil)

// Open은 데이터베이스에 연결하고 접속을 확인합니다. 호출한 쪽에서 Close를 불러야 합니다.
func Open(ctx context.Context, databaseURL string) (*Repository, error) {
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
	db := stdlib.OpenDBFromPool(r.pool)
	defer db.Close()
	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations)
	if err != nil {
		return 0, fmt.Errorf("prepare migrations: %w", err)
	}
	results, err := provider.Up(ctx)
	if err != nil {
		return 0, fmt.Errorf("apply migrations: %w", err)
	}
	return len(results), nil
}

// List는 최근 등록한 사건부터, 같은 때 등록한 사건은 사건번호가 큰 순서로 돌려줍니다.
func (r *Repository) List(ctx context.Context) ([]clio.StoredCase, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::text, year, code, serial, title, category, court_id, panel_name
		FROM cases
		ORDER BY created_at DESC, year DESC, serial DESC`)
	if err != nil {
		return nil, fmt.Errorf("list cases: %w", err)
	}
	items, err := pgx.CollectRows(rows, scanCase)
	if err != nil {
		return nil, fmt.Errorf("list cases: %w", err)
	}
	if err := r.loadParties(ctx, items, nil); err != nil {
		return nil, err
	}
	return items, nil
}

// Get은 ID로 사건 한 건을 찾습니다. ID가 UUID 형식이 아니어도 ErrCaseNotFound를 돌려줍니다.
func (r *Repository) Get(ctx context.Context, id string) (clio.StoredCase, error) {
	var uuid pgtype.UUID
	if err := uuid.Scan(id); err != nil {
		return clio.StoredCase{}, clio.ErrCaseNotFound
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id::text, year, code, serial, title, category, court_id, panel_name
		FROM cases
		WHERE id = $1`, uuid)
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
	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		for i, c := range cases {
			var id string
			err := tx.QueryRow(ctx, `
				INSERT INTO cases (year, code, serial, title, category, court_id, panel_name)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
				RETURNING id::text`,
				c.Number.Year, c.Number.Code, c.Number.Serial, c.Title, string(c.Category),
				c.Court.CourtID, c.Court.PanelName,
			).Scan(&id)
			if err != nil {
				return err
			}
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
	if err := rows.Err(); err != nil {
		return fmt.Errorf("load parties: %w", err)
	}
	return nil
}

// scanCase는 cases 테이블의 한 행을 StoredCase로 읽습니다.
func scanCase(row pgx.CollectableRow) (clio.StoredCase, error) {
	var item clio.StoredCase
	var category string
	err := row.Scan(&item.ID, &item.Case.Number.Year, &item.Case.Number.Code, &item.Case.Number.Serial,
		&item.Case.Title, &category, &item.Case.Court.CourtID, &item.Case.Court.PanelName)
	item.Case.Category = clio.CaseCategory(category)
	return item, err
}
