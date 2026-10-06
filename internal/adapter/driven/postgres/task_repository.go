package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Kiryue0/task-api/internal/core/domain"
	"github.com/Kiryue0/task-api/internal/core/port/driven"
)

// TaskRepository, task'ları user_tables şemasındaki ilgili tabloda saklar.
type TaskRepository struct {
	db *sql.DB
}

var _ driven.TaskRepository = (*TaskRepository)(nil)

// NewTaskRepository, verilen *sql.DB ile bir repository oluşturur.
func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// wrap, "tablo yok" hatasını domain.ErrTableNotFound'a çevirir, diğerlerine bağlam ekler.
func wrap(op, table string, err error) error {
	if pgErrorCode(err) == codeUndefinedTable {
		return domain.ErrTableNotFound
	}
	return fmt.Errorf("%s (table %q): %w", op, table, err)
}

func (r *TaskRepository) Insert(ctx context.Context, table string, task domain.Task) (domain.Task, error) {
	q := fmt.Sprintf(`INSERT INTO %s (title) VALUES ($1) RETURNING id, title`, qualifiedTable(table))
	var out domain.Task
	if err := r.db.QueryRowContext(ctx, q, task.Title).Scan(&out.ID, &out.Title); err != nil {
		return domain.Task{}, wrap("insert task", table, err)
	}
	return out, nil
}

func (r *TaskRepository) FindAll(ctx context.Context, table string) ([]domain.Task, error) {
	q := fmt.Sprintf(`SELECT id, title FROM %s ORDER BY id`, qualifiedTable(table))
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, wrap("query tasks", table, err)
	}
	defer rows.Close()

	tasks := make([]domain.Task, 0)
	for rows.Next() {
		var t domain.Task
		if err := rows.Scan(&t.ID, &t.Title); err != nil {
			return nil, wrap("scan task", table, err)
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, wrap("iterate tasks", table, err)
	}
	return tasks, nil
}

func (r *TaskRepository) FindByID(ctx context.Context, table string, id int) (domain.Task, error) {
	q := fmt.Sprintf(`SELECT id, title FROM %s WHERE id = $1`, qualifiedTable(table))
	var out domain.Task
	err := r.db.QueryRowContext(ctx, q, id).Scan(&out.ID, &out.Title)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Task{}, domain.ErrTaskNotFound
	}
	if err != nil {
		return domain.Task{}, wrap("get task", table, err)
	}
	return out, nil
}

func (r *TaskRepository) Update(ctx context.Context, table string, task domain.Task) (domain.Task, error) {
	q := fmt.Sprintf(`UPDATE %s SET title = $1 WHERE id = $2 RETURNING id, title`, qualifiedTable(table))
	var out domain.Task
	err := r.db.QueryRowContext(ctx, q, task.Title, task.ID).Scan(&out.ID, &out.Title)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Task{}, domain.ErrTaskNotFound
	}
	if err != nil {
		return domain.Task{}, wrap("update task", table, err)
	}
	return out, nil
}

func (r *TaskRepository) DeleteByID(ctx context.Context, table string, id int) error {
	q := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, qualifiedTable(table))
	res, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return wrap("delete task", table, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete task rows affected (table %q): %w", table, err)
	}
	if n == 0 {
		return domain.ErrTaskNotFound
	}
	return nil
}
