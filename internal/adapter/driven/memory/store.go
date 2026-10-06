// Package memory, driven port'ların bellek-içi adapter'ıdır.
// Veritabanı yapılandırılmadığında kullanılır. Veriler yalnızca sürecin belleğinde durur:
// pod yeniden başlarsa kaybolur ve her replika kendi ayrı verisini tutar.
package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/Kiryue0/task-api/internal/core/domain"
	"github.com/Kiryue0/task-api/internal/core/port/driven"
)

// Store, hem driven.TableRepository hem driven.TaskRepository'yi uygular.
// Davranışı Postgres adapter'ıyla aynıdır: tablo başına ayrı id sayacı, silinen id'ler
// tekrar kullanılmaz, tablo silinip yeniden oluşturulursa id'ler 1'den başlar.
type Store struct {
	mu     sync.RWMutex
	tables map[string]*table
}

type table struct {
	nextID int
	rows   map[int]domain.Task
}

var (
	_ driven.TableRepository = (*Store)(nil)
	_ driven.TaskRepository  = (*Store)(nil)
)

// NewStore, boş bir bellek-içi depo oluşturur.
func NewStore() *Store {
	return &Store{tables: map[string]*table{}}
}

// --- driven.TableRepository ---

func (s *Store) Create(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tables[name]; ok {
		return domain.ErrTableExists
	}
	s.tables[name] = &table{nextID: 1, rows: map[int]domain.Task{}}
	return nil
}

func (s *Store) List(_ context.Context) ([]domain.Table, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Table, 0, len(s.tables))
	for name := range s.tables {
		out = append(out, domain.Table{Name: name})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (s *Store) Drop(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tables[name]; !ok {
		return domain.ErrTableNotFound
	}
	delete(s.tables, name)
	return nil
}

// --- driven.TaskRepository ---

// table, adı verilen tabloyu döner. Çağıran kilidi tutmalıdır.
func (s *Store) table(name string) (*table, error) {
	t, ok := s.tables[name]
	if !ok {
		return nil, domain.ErrTableNotFound
	}
	return t, nil
}

func (s *Store) Insert(_ context.Context, tableName string, task domain.Task) (domain.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, err := s.table(tableName)
	if err != nil {
		return domain.Task{}, err
	}
	task.ID = t.nextID
	t.nextID++
	t.rows[task.ID] = task
	return task, nil
}

func (s *Store) FindAll(_ context.Context, tableName string) ([]domain.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, err := s.table(tableName)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Task, 0, len(t.rows))
	for _, r := range t.rows {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (s *Store) FindByID(_ context.Context, tableName string, id int) (domain.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, err := s.table(tableName)
	if err != nil {
		return domain.Task{}, err
	}
	r, ok := t.rows[id]
	if !ok {
		return domain.Task{}, domain.ErrTaskNotFound
	}
	return r, nil
}

func (s *Store) Update(_ context.Context, tableName string, task domain.Task) (domain.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, err := s.table(tableName)
	if err != nil {
		return domain.Task{}, err
	}
	if _, ok := t.rows[task.ID]; !ok {
		return domain.Task{}, domain.ErrTaskNotFound
	}
	t.rows[task.ID] = task
	return task, nil
}

func (s *Store) DeleteByID(_ context.Context, tableName string, id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, err := s.table(tableName)
	if err != nil {
		return err
	}
	if _, ok := t.rows[id]; !ok {
		return domain.ErrTaskNotFound
	}
	delete(t.rows, id)
	return nil
}
