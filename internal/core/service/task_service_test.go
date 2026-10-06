package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Kiryue0/task-api/internal/adapter/driven/memory"
	"github.com/Kiryue0/task-api/internal/core/domain"
	"github.com/Kiryue0/task-api/internal/core/service"
)

// newServices, bellek-içi depo üzerinde bir TaskService ve hazır tablolar döner.
// Postgres yerine memory adapter'ı kullanılır: core'un veritabanından bağımsız olduğunun kanıtı.
func newServices(t *testing.T, tables ...string) (*service.TaskService, *memory.Store) {
	t.Helper()
	store := memory.NewStore()
	tableSvc := service.NewTableService(store)
	for _, name := range tables {
		if _, err := tableSvc.CreateTable(context.Background(), name); err != nil {
			t.Fatalf("CreateTable(%q): %v", name, err)
		}
	}
	return service.NewTaskService(store), store
}

func TestTaskCRUD(t *testing.T) {
	ctx := context.Background()
	svc, _ := newServices(t, "alisveris")

	a, err := svc.CreateTask(ctx, "alisveris", "  süt al  ")
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if a.ID == 0 || a.Title != "süt al" {
		t.Fatalf("beklenmeyen task: %+v (title trim edilmeli, id atanmalı)", a)
	}
	b, err := svc.CreateTask(ctx, "alisveris", "ekmek al")
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	got, err := svc.GetTask(ctx, "alisveris", a.ID)
	if err != nil || got != a {
		t.Fatalf("GetTask: %+v, %v", got, err)
	}

	updated, err := svc.UpdateTask(ctx, "alisveris", a.ID, "yarım yağlı süt al")
	if err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}
	if updated.ID != a.ID || updated.Title != "yarım yağlı süt al" {
		t.Fatalf("beklenmeyen güncelleme: %+v", updated)
	}

	if err := svc.DeleteTask(ctx, "alisveris", b.ID); err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}
	list, err := svc.ListTasks(ctx, "alisveris")
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	if len(list) != 1 || list[0] != updated {
		t.Fatalf("beklenmeyen liste: %+v", list)
	}
}

func TestTablesAreIsolated(t *testing.T) {
	ctx := context.Background()
	svc, _ := newServices(t, "alisveris", "okul")

	if _, err := svc.CreateTask(ctx, "alisveris", "süt al"); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	list, err := svc.ListTasks(ctx, "okul")
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("okul tablosu boş olmalı, alınan: %+v", list)
	}
}

func TestTaskInvalidTitle(t *testing.T) {
	ctx := context.Background()
	svc, store := newServices(t, "alisveris")

	for _, title := range []string{"", "   ", "\t\n"} {
		if _, err := svc.CreateTask(ctx, "alisveris", title); !errors.Is(err, domain.ErrInvalidTitle) {
			t.Errorf("CreateTask(%q): ErrInvalidTitle beklendi, alınan: %v", title, err)
		}
	}
	if rows, _ := store.FindAll(ctx, "alisveris"); len(rows) != 0 {
		t.Fatalf("geçersiz title repository'ye ulaşmamalı, satırlar: %+v", rows)
	}

	a, _ := svc.CreateTask(ctx, "alisveris", "süt al")
	if _, err := svc.UpdateTask(ctx, "alisveris", a.ID, " "); !errors.Is(err, domain.ErrInvalidTitle) {
		t.Fatalf("UpdateTask: ErrInvalidTitle beklendi, alınan: %v", err)
	}
}

func TestTaskNotFoundErrors(t *testing.T) {
	ctx := context.Background()
	svc, _ := newServices(t, "alisveris")

	if _, err := svc.GetTask(ctx, "alisveris", 42); !errors.Is(err, domain.ErrTaskNotFound) {
		t.Errorf("GetTask: ErrTaskNotFound beklendi, alınan: %v", err)
	}
	if _, err := svc.UpdateTask(ctx, "alisveris", 42, "x"); !errors.Is(err, domain.ErrTaskNotFound) {
		t.Errorf("UpdateTask: ErrTaskNotFound beklendi, alınan: %v", err)
	}
	if err := svc.DeleteTask(ctx, "alisveris", 42); !errors.Is(err, domain.ErrTaskNotFound) {
		t.Errorf("DeleteTask: ErrTaskNotFound beklendi, alınan: %v", err)
	}
	if _, err := svc.ListTasks(ctx, "yok"); !errors.Is(err, domain.ErrTableNotFound) {
		t.Errorf("ListTasks: ErrTableNotFound beklendi, alınan: %v", err)
	}
	if _, err := svc.CreateTask(ctx, "Bozuk-Ad", "x"); !errors.Is(err, domain.ErrInvalidTableName) {
		t.Errorf("CreateTask: ErrInvalidTableName beklendi, alınan: %v", err)
	}
}
