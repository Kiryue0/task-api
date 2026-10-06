package domain

import (
	"errors"
	"regexp"
	"strings"
)

// Table, kullanıcının oluşturduğu ve içinde Task satırları tutan bir tablodur.
type Table struct {
	Name string
}

var (
	// ErrInvalidTableName, tablo adı kurallara uymadığında döner.
	ErrInvalidTableName = errors.New("table name must start with a letter and contain only a-z, 0-9 and _ (max 63 chars)")
	// ErrTableNotFound, istenen tablo yoksa döner.
	ErrTableNotFound = errors.New("table not found")
	// ErrTableExists, aynı adla bir tablo zaten varsa döner.
	ErrTableExists = errors.New("table already exists")
)

// Postgres identifier sınırı 63 karakterdir. Yalnızca ASCII küçük harf, rakam ve _
// kabul edilir; böylece ad SQL içinde hiçbir zaman beklenmedik bir anlam taşıyamaz.
var tableNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

// ParseTableName, ham girdiyi normalize eder (boşlukları kırpar, küçük harfe çevirir)
// ve geçerliyse döner. Geçersizse ErrInvalidTableName döner.
func ParseTableName(raw string) (string, error) {
	name := strings.ToLower(strings.TrimSpace(raw))
	if !tableNamePattern.MatchString(name) {
		return "", ErrInvalidTableName
	}
	return name, nil
}
