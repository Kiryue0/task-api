package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

// loadDotEnv, verilen dosyadaki KEY=VALUE satırlarını ortam değişkeni olarak yükler.
// Yerel geliştirme kolaylığı içindir:
//   - Dosya yoksa hiçbir şey yapmaz (Kubernetes'te dosya yoktur, değerler ortamdan gelir).
//   - Ortamda zaten tanımlı bir değişkenin üzerine YAZMAZ; dışarıdan verilen değer kazanır.
//
// Desteklenen biçim: boş satırlar ve # ile başlayan satırlar yok sayılır, isteğe bağlı
// "export " öneki kabul edilir, değer tek veya çift tırnak içindeyse tırnaklar atılır.
// Yüklenen değişken sayısını döner.
func loadDotEnv(path string) (int, error) {
	f, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	loaded := 0
	scanner := bufio.NewScanner(f)
	for lineNo := 1; scanner.Scan(); lineNo++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")

		key, value, ok := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" {
			return loaded, fmt.Errorf("%s:%d: expected KEY=VALUE", path, lineNo)
		}
		value = unquote(strings.TrimSpace(value))

		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return loaded, fmt.Errorf("%s:%d: set %s: %w", path, lineNo, key, err)
		}
		loaded++
	}
	if err := scanner.Err(); err != nil {
		return loaded, fmt.Errorf("read %s: %w", path, err)
	}
	return loaded, nil
}

// unquote, değer eşleşen tek veya çift tırnak içindeyse tırnakları atar.
func unquote(v string) string {
	if len(v) >= 2 && (v[0] == '"' || v[0] == '\'') && v[len(v)-1] == v[0] {
		return v[1 : len(v)-1]
	}
	return v
}
