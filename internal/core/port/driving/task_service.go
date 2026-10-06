// Package driving, dış dünyanın (HTTP vb.) çekirdeği nasıl kullanacağını tanımlar.
package driving

import (
	"context"

	"github.com/Kiryue0/task-api/internal/core/domain"
)

// TaskService, bir tablo içindeki task'lar üzerindeki use-case'lerdir.
// Tüm metotlar, tablo yoksa domain.ErrTableNotFound, tablo adı geçersizse
// domain.ErrInvalidTableName döner.
type TaskService interface {
	// CreateTask yeni bir task oluşturur. Title boşsa domain.ErrInvalidTitle döner.
	CreateTask(ctx context.Context, table, title string) (domain.Task, error)
	// ListTasks tablodaki tüm task'ları id sırasıyla döner.
	ListTasks(ctx context.Context, table string) ([]domain.Task, error)
	// GetTask tek bir task döner. Yoksa domain.ErrTaskNotFound döner.
	GetTask(ctx context.Context, table string, id int) (domain.Task, error)
	// UpdateTask task'ın title'ını değiştirir. Title boşsa domain.ErrInvalidTitle,
	// task yoksa domain.ErrTaskNotFound döner.
	UpdateTask(ctx context.Context, table string, id int, title string) (domain.Task, error)
	// DeleteTask task'ı siler. Yoksa domain.ErrTaskNotFound döner.
	DeleteTask(ctx context.Context, table string, id int) error
}
