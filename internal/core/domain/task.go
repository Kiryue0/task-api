// Package domain, uygulamanın çekirdek varlıklarını ve domain hatalarını içerir.
// Hiçbir dış pakete (HTTP, SQL vb.) bağımlı değildir.
package domain

import "errors"

// Task, yönetilen tek varlıktır.
type Task struct {
	ID    int
	Title string
}

var (
	// ErrInvalidTitle, title boş (veya yalnızca boşluk) olduğunda döner.
	ErrInvalidTitle = errors.New("title must not be empty")
	// ErrTaskNotFound, istenen id'ye sahip task yoksa döner.
	ErrTaskNotFound = errors.New("task not found")
)
