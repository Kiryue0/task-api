// Package driven, çekirdeğin ihtiyaç duyduğu dış servisleri (kalıcılık vb.) tanımlar.
package driven

import (
	"context"

	"github.com/Kiryue0/task-api/internal/core/domain"
)

// TaskRepository, task'ların belirli bir tabloda saklanmasından sorumludur.
// Tüm metotlar, tablo yoksa domain.ErrTableNotFound döner.
type TaskRepository interface {
	// Insert task'ı kaydeder ve id'si atanmış halini döner (task.ID yok sayılır).
	Insert(ctx context.Context, table string, task domain.Task) (domain.Task, error)
	// FindAll tüm task'ları id sırasıyla döner. Kayıt yoksa boş slice döner.
	FindAll(ctx context.Context, table string) ([]domain.Task, error)
	// FindByID tek bir task döner. Yoksa domain.ErrTaskNotFound döner.
	FindByID(ctx context.Context, table string, id int) (domain.Task, error)
	// Update task.ID'li kaydın title'ını günceller. Yoksa domain.ErrTaskNotFound döner.
	Update(ctx context.Context, table string, task domain.Task) (domain.Task, error)
	// DeleteByID task'ı siler. Yoksa domain.ErrTaskNotFound döner.
	DeleteByID(ctx context.Context, table string, id int) error
}
