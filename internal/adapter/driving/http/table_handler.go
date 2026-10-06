package httpadapter

import (
	"net/http"

	"github.com/Kiryue0/task-api/internal/core/domain"
	"github.com/Kiryue0/task-api/internal/core/port/driving"
)

// TableHandler, tablo yönetimi endpoint'lerini sunar.
type TableHandler struct {
	svc driving.TableService
}

// NewTableHandler, verilen servis ile bir handler oluşturur.
func NewTableHandler(svc driving.TableService) *TableHandler {
	return &TableHandler{svc: svc}
}

// RegisterRoutes, endpoint'leri verilen mux'a kaydeder.
func (h *TableHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /tables", h.create)
	mux.HandleFunc("GET /tables", h.list)
	mux.HandleFunc("DELETE /tables/{table}", h.delete)
}

// tableDTO, domain.Table'ın HTTP üzerindeki JSON temsilidir.
type tableDTO struct {
	Name string `json:"name"`
}

type createTableRequest struct {
	Name string `json:"name"`
}

func (h *TableHandler) create(w http.ResponseWriter, r *http.Request) {
	var req createTableRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	table, err := h.svc.CreateTable(r.Context(), req.Name)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toTableDTO(table))
}

func (h *TableHandler) list(w http.ResponseWriter, r *http.Request) {
	tables, err := h.svc.ListTables(r.Context())
	if err != nil {
		handleServiceError(w, r, err)
		return
	}
	out := make([]tableDTO, 0, len(tables))
	for _, t := range tables {
		out = append(out, toTableDTO(t))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *TableHandler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteTable(r.Context(), r.PathValue("table")); err != nil {
		handleServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func toTableDTO(t domain.Table) tableDTO {
	return tableDTO{Name: t.Name}
}
