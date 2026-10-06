package httpadapter

import (
	"net/http"
	"strconv"

	"github.com/Kiryue0/task-api/internal/core/domain"
	"github.com/Kiryue0/task-api/internal/core/port/driving"
)

// TaskHandler, bir tablo içindeki task endpoint'lerini sunar.
type TaskHandler struct {
	svc driving.TaskService
}

// NewTaskHandler, verilen servis ile bir handler oluşturur.
func NewTaskHandler(svc driving.TaskService) *TaskHandler {
	return &TaskHandler{svc: svc}
}

// RegisterRoutes, endpoint'leri verilen mux'a kaydeder.
func (h *TaskHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /tables/{table}/tasks", h.create)
	mux.HandleFunc("GET /tables/{table}/tasks", h.list)
	mux.HandleFunc("GET /tables/{table}/tasks/{id}", h.get)
	mux.HandleFunc("PUT /tables/{table}/tasks/{id}", h.update)
	mux.HandleFunc("DELETE /tables/{table}/tasks/{id}", h.delete)
}

// taskDTO, domain.Task'ın HTTP üzerindeki JSON temsilidir.
type taskDTO struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// taskRequest, oluşturma ve güncelleme isteklerinin gövdesidir.
type taskRequest struct {
	Title string `json:"title"`
}

func (h *TaskHandler) create(w http.ResponseWriter, r *http.Request) {
	var req taskRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	task, err := h.svc.CreateTask(r.Context(), r.PathValue("table"), req.Title)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toTaskDTO(task))
}

func (h *TaskHandler) list(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.svc.ListTasks(r.Context(), r.PathValue("table"))
	if err != nil {
		handleServiceError(w, r, err)
		return
	}
	out := make([]taskDTO, 0, len(tasks))
	for _, t := range tasks {
		out = append(out, toTaskDTO(t))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *TaskHandler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	task, err := h.svc.GetTask(r.Context(), r.PathValue("table"), id)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toTaskDTO(task))
}

func (h *TaskHandler) update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var req taskRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	task, err := h.svc.UpdateTask(r.Context(), r.PathValue("table"), id, req.Title)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toTaskDTO(task))
}

func (h *TaskHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := h.svc.DeleteTask(r.Context(), r.PathValue("table"), id); err != nil {
		handleServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// parseID, {id} path değerini okur. Geçersizse 400 yazar ve false döner.
func parseID(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "id must be a positive integer")
		return 0, false
	}
	return id, true
}

func toTaskDTO(t domain.Task) taskDTO {
	return taskDTO{ID: t.ID, Title: t.Title}
}
