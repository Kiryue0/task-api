package driving

import (
	"context"

	"github.com/Kiryue0/task-api/internal/core/domain"
)

// TableService, kullanıcı tablolarının yönetimine dair use-case'lerdir.
type TableService interface {
	// CreateTable yeni bir tablo oluşturur. Ad geçersizse domain.ErrInvalidTableName,
	// zaten varsa domain.ErrTableExists döner.
	CreateTable(ctx context.Context, name string) (domain.Table, error)
	// ListTables tüm tabloları ada göre sıralı döner.
	ListTables(ctx context.Context) ([]domain.Table, error)
	// DeleteTable tabloyu içindeki tüm satırlarla birlikte siler.
	// Yoksa domain.ErrTableNotFound döner.
	DeleteTable(ctx context.Context, name string) error
}
