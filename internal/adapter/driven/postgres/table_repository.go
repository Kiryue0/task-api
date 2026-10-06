package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Kiryue0/task-api/internal/core/domain"
	"github.com/Kiryue0/task-api/internal/core/port/driven"
)

// TableRepository, her kullanıcı tablosu için user_tables şemasında gerçek bir
// Postgres tablosu oluşturur.
type TableRepository struct {
	db *sql.DB
}

var _ driven.TableRepository = (*TableRepository)(nil)

// NewTableRepository, verilen *sql.DB ile bir repository oluşturur.
func NewTableRepository(db *sql.DB) *TableRepository {
	return &TableRepository{db: db}
}

func (r *TableRepository) Create(ctx context.Context, name string) error {
	// DDL'de parametre ($1) kullanılamaz; ad bu yüzden tırnaklanarak eklenir.
	q := fmt.Sprintf(`CREATE TABLE %s (
		id    SERIAL PRIMARY KEY,
		title TEXT   NOT NULL CHECK (btrim(title) <> '')
	)`, qualifiedTable(name))

	if _, err := r.db.ExecContext(ctx, q); err != nil {
		switch pgErrorCode(err) {
		// 23505: aynı ad eşzamanlı iki istekle oluşturulmaya çalışılırsa gelebilir.
		case codeDuplicateTable, codeUniqueViolation:
			return domain.ErrTableExists
		}
		return fmt.Errorf("create table %q: %w", name, err)
	}
	return nil
}

func (r *TableRepository) List(ctx context.Context) ([]domain.Table, error) {
	const q = `SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname = $1 ORDER BY tablename`
	rows, err := r.db.QueryContext(ctx, q, userSchema)
	if err != nil {
		return nil, fmt.Errorf("list tables: %w", err)
	}
	defer rows.Close()

	tables := make([]domain.Table, 0)
	for rows.Next() {
		var t domain.Table
		if err := rows.Scan(&t.Name); err != nil {
			return nil, fmt.Errorf("scan table: %w", err)
		}
		tables = append(tables, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tables: %w", err)
	}
	return tables, nil
}

func (r *TableRepository) Drop(ctx context.Context, name string) error {
	q := fmt.Sprintf(`DROP TABLE %s`, qualifiedTable(name))
	if _, err := r.db.ExecContext(ctx, q); err != nil {
		if pgErrorCode(err) == codeUndefinedTable {
			return domain.ErrTableNotFound
		}
		return fmt.Errorf("drop table %q: %w", name, err)
	}
	return nil
}
