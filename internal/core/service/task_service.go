// Package service, driving port'ları uygulayan çekirdek iş mantığıdır.
// Yalnızca domain ve port paketlerini bilir; hiçbir adapter'ı import etmez.
package service

import (
	"context"
	"strings"

	"github.com/Kiryue0/task-api/internal/core/domain"
	"github.com/Kiryue0/task-api/internal/core/port/driven"
	"github.com/Kiryue0/task-api/internal/core/port/driving"
)

// TaskService, driving.TaskService'in çekirdek uygulamasıdır.
type TaskService struct {
	repo driven.TaskRepository
}

// Derleme zamanında port'u uyguladığımızı garanti et.
var _ driving.TaskService = (*TaskService)(nil)

// NewTaskService, verilen repository ile bir TaskService oluşturur.
func NewTaskService(repo driven.TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) CreateTask(ctx context.Context, table, title string) (domain.Task, error) {
	table, err := domain.ParseTableName(table)
	if err != nil {
		return domain.Task{}, err
	}
	title, err = parseTitle(title)
	if err != nil {
		return domain.Task{}, err
	}
	return s.repo.Insert(ctx, table, domain.Task{Title: title})
}

func (s *TaskService) ListTasks(ctx context.Context, table string) ([]domain.Task, error) {
	table, err := domain.ParseTableName(table)
	if err != nil {
		return nil, err
	}
	return s.repo.FindAll(ctx, table)
}

func (s *TaskService) GetTask(ctx context.Context, table string, id int) (domain.Task, error) {
	table, err := domain.ParseTableName(table)
	if err != nil {
		return domain.Task{}, err
	}
	return s.repo.FindByID(ctx, table, id)
}

func (s *TaskService) UpdateTask(ctx context.Context, table string, id int, title string) (domain.Task, error) {
	table, err := domain.ParseTableName(table)
	if err != nil {
		return domain.Task{}, err
	}
	title, err = parseTitle(title)
	if err != nil {
		return domain.Task{}, err
	}
	return s.repo.Update(ctx, table, domain.Task{ID: id, Title: title})
}

func (s *TaskService) DeleteTask(ctx context.Context, table string, id int) error {
	table, err := domain.ParseTableName(table)
	if err != nil {
		return err
	}
	return s.repo.DeleteByID(ctx, table, id)
}

func parseTitle(raw string) (string, error) {
	title := strings.TrimSpace(raw)
	if title == "" {
		return "", domain.ErrInvalidTitle
	}
	return title, nil
}
