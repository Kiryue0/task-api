package driven

import (
	"context"

	"github.com/Kiryue0/task-api/internal/core/domain"
)

// TableRepository, kullanıcı tablolarının fiziksel olarak oluşturulup silinmesinden sorumludur.
// Verilen adların domain.ParseTableName ile doğrulanmış olduğu varsayılır.
type TableRepository interface {
	// Create tabloyu oluşturur. Zaten varsa domain.ErrTableExists döner.
	Create(ctx context.Context, name string) error
	// List tüm tabloları ada göre sıralı döner.
	List(ctx context.Context) ([]domain.Table, error)
	// Drop tabloyu siler. Yoksa domain.ErrTableNotFound döner.
	Drop(ctx context.Context, name string) error
}
