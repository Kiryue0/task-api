// Package httpadapter, driving port'ları HTTP/JSON üzerinden dışarı açar.
package httpadapter

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/Kiryue0/task-api/internal/core/domain"
)

const maxBodyBytes = 1 << 20 // 1 MiB

type errorResponse struct {
	Error string `json:"error"`
}

// decodeJSON, istek gövdesini boyut sınırıyla okur. Hata durumunda 400 yazar ve false döner.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return false
	}
	return true
}

// handleServiceError, domain hatalarını HTTP status code'larına çevirir.
func handleServiceError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidTitle),
		errors.Is(err, domain.ErrInvalidTableName):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrTaskNotFound),
		errors.Is(err, domain.ErrTableNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrTableExists):
		writeError(w, http.StatusConflict, err.Error())
	default:
		// İç hata detayını istemciye sızdırma; sadece logla.
		log.Printf("ERROR %s %s: %v", r.Method, r.URL.Path, err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("ERROR encode response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}
