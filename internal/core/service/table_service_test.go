package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Kiryue0/task-api/internal/adapter/driven/memory"
	"github.com/Kiryue0/task-api/internal/core/domain"
	"github.com/Kiryue0/task-api/internal/core/service"
)

func TestCreateListDeleteTable(t *testing.T) {
	ctx := context.Background()
	svc := service.NewTableService(memory.NewStore())

	tbl, err := svc.CreateTable(ctx, "  Alisveris ")
	if err != nil {
		t.Fatalf("CreateTable: %v", err)
	}
	if tbl.Name != "alisveris" {
		t.Fatalf("ad normalize edilmeli, alınan: %q", tbl.Name)
	}
	if _, err := svc.CreateTable(ctx, "okul"); err != nil {
		t.Fatalf("CreateTable: %v", err)
	}

	list, err := svc.ListTables(ctx)
	if err != nil {
		t.Fatalf("ListTables: %v", err)
	}
	if len(list) != 2 || list[0].Name != "alisveris" || list[1].Name != "okul" {
		t.Fatalf("beklenmeyen liste: %+v", list)
	}

	if err := svc.DeleteTable(ctx, "okul"); err != nil {
		t.Fatalf("DeleteTable: %v", err)
	}
	if err := svc.DeleteTable(ctx, "okul"); !errors.Is(err, domain.ErrTableNotFound) {
		t.Fatalf("ikinci silmede ErrTableNotFound beklendi, alınan: %v", err)
	}
}

func TestCreateTableDuplicate(t *testing.T) {
	ctx := context.Background()
	svc := service.NewTableService(memory.NewStore())

	if _, err := svc.CreateTable(ctx, "okul"); err != nil {
		t.Fatalf("CreateTable: %v", err)
	}
	// Normalize sonrası aynı ad olduğu için çakışmalı.
	if _, err := svc.CreateTable(ctx, "OKUL"); !errors.Is(err, domain.ErrTableExists) {
		t.Fatalf("ErrTableExists beklendi, alınan: %v", err)
	}
}

func TestCreateTableRejectsInvalidNames(t *testing.T) {
	store := memory.NewStore()
	svc := service.NewTableService(store)

	invalid := []string{
		"", "   ", "1tablo", "_gizli", "alışveriş", "a-b", "a b",
		`x"; DROP TABLE users; --`,
		"a234567890123456789012345678901234567890123456789012345678901234", // 64 karakter
	}
	for _, name := range invalid {
		if _, err := svc.CreateTable(context.Background(), name); !errors.Is(err, domain.ErrInvalidTableName) {
			t.Errorf("%q için ErrInvalidTableName beklendi, alınan: %v", name, err)
		}
	}
	if tables, _ := store.List(context.Background()); len(tables) != 0 {
		t.Fatalf("geçersiz adlar repository'ye ulaşmamalı, tablolar: %v", tables)
	}
}
