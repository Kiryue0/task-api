package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadDotEnv(t *testing.T) {
	path := writeFile(t, `
# yorum satırı
TEST_DOTENV_PLAIN=localhost
export TEST_DOTENV_EXPORT=5432
TEST_DOTENV_DOUBLE="p@ss w=rd"
TEST_DOTENV_SINGLE='tek'
TEST_DOTENV_EXISTING=dosyadan
`)
	for _, k := range []string{"TEST_DOTENV_PLAIN", "TEST_DOTENV_EXPORT", "TEST_DOTENV_DOUBLE", "TEST_DOTENV_SINGLE"} {
		t.Setenv(k, "") // test bitince eski haline dönsün
		os.Unsetenv(k)
	}
	t.Setenv("TEST_DOTENV_EXISTING", "ortamdan")

	n, err := loadDotEnv(path)
	if err != nil {
		t.Fatalf("loadDotEnv: %v", err)
	}
	if n != 4 {
		t.Errorf("4 değişken yüklenmeli, yüklenen: %d", n)
	}

	want := map[string]string{
		"TEST_DOTENV_PLAIN":    "localhost",
		"TEST_DOTENV_EXPORT":   "5432",
		"TEST_DOTENV_DOUBLE":   "p@ss w=rd",
		"TEST_DOTENV_SINGLE":   "tek",
		"TEST_DOTENV_EXISTING": "ortamdan", // ortamdaki değer ezilmemeli
	}
	for k, v := range want {
		if got := os.Getenv(k); got != v {
			t.Errorf("%s = %q, beklenen %q", k, got, v)
		}
	}
}

func TestLoadDotEnvMissingFile(t *testing.T) {
	n, err := loadDotEnv(filepath.Join(t.TempDir(), "yok.env"))
	if err != nil || n != 0 {
		t.Fatalf("dosya yoksa sessizce geçmeli, alınan: n=%d err=%v", n, err)
	}
}

func TestLoadDotEnvInvalidLine(t *testing.T) {
	path := writeFile(t, "GECERLI=1\nbozuk satir\n")
	t.Setenv("GECERLI", "")
	os.Unsetenv("GECERLI")

	if _, err := loadDotEnv(path); err == nil {
		t.Fatal("geçersiz satır için hata beklendi")
	}
}
