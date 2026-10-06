package service

import (
	"context"

	"github.com/Kiryue0/task-api/internal/core/domain"
	"github.com/Kiryue0/task-api/internal/core/port/driven"
	"github.com/Kiryue0/task-api/internal/core/port/driving"
)

// TableService, driving.TableService'in çekirdek uygulamasıdır.
type TableService struct {
	repo driven.TableRepository
}

var _ driving.TableService = (*TableService)(nil)

// NewTableService, verilen repository ile bir TableService oluşturur.
func NewTableService(repo driven.TableRepository) *TableService {
	return &TableService{repo: repo}
}

func (s *TableService) CreateTable(ctx context.Context, name string) (domain.Table, error) {
	name, err := domain.ParseTableName(name)
	if err != nil {
		return domain.Table{}, err
	}
	if err := s.repo.Create(ctx, name); err != nil {
		return domain.Table{}, err
	}
	return domain.Table{Name: name}, nil
}

func (s *TableService) ListTables(ctx context.Context) ([]domain.Table, error) {
	return s.repo.List(ctx)
}

func (s *TableService) DeleteTable(ctx context.Context, name string) error {
	name, err := domain.ParseTableName(name)
	if err != nil {
		return err
	}
	return s.repo.Drop(ctx, name)
}
