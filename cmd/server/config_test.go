package main

import (
	"strings"
	"testing"
)

func setDBEnv(t *testing.T, vals map[string]string) {
	t.Helper()
	for _, k := range dbEnvKeys {
		t.Setenv(k, vals[k]) // boş string = tanımsız sayılır
	}
}

func TestDSNFromEnvNoneSetUsesMemory(t *testing.T) {
	setDBEnv(t, nil)

	dsn, configured, err := dsnFromEnv()
	if err != nil || configured || dsn != "" {
		t.Fatalf("hiçbiri yoksa bellek modu beklenir, alınan: dsn=%q configured=%v err=%v", dsn, configured, err)
	}
}

func TestDSNFromEnvAllSet(t *testing.T) {
	setDBEnv(t, map[string]string{
		"DB_HOST": "postgres", "DB_PORT": "5432", "DB_USER": "app",
		"DB_PASSWORD": "p@ss:w/rd", "DB_NAME": "tasks",
	})

	dsn, configured, err := dsnFromEnv()
	if err != nil || !configured {
		t.Fatalf("hepsi varsa postgres modu beklenir, alınan: configured=%v err=%v", configured, err)
	}
	if want := "postgres://app:p%40ss%3Aw%2Frd@postgres:5432/tasks"; dsn != want {
		t.Fatalf("dsn = %q, beklenen %q", dsn, want)
	}
}

func TestDSNFromEnvPartialIsError(t *testing.T) {
	setDBEnv(t, map[string]string{"DB_HOST": "postgres", "DB_PORT": "5432", "DB_USER": "app", "DB_NAME": "tasks"})

	_, _, err := dsnFromEnv()
	if err == nil || !strings.Contains(err.Error(), "DB_PASSWORD") {
		t.Fatalf("eksik DB_PASSWORD için hata beklenir, alınan: %v", err)
	}
}
