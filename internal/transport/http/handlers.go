package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"

	"github.com/dimakropachev/image_resizer_service/internal/config"
	"github.com/dimakropachev/image_resizer_service/internal/models"
	"github.com/dimakropachev/image_resizer_service/internal/repository"
	"github.com/google/uuid"
)

type Service interface {
	AddTask(context.Context, *models.Task) error
	GetStatus(context.Context, string) (string, error, error)
	DeleteTask(context.Context, string) error
}

type Handler struct {
	s     Service
	paths config.Storage
}

func NewHandler(s Service, paths config.Storage) Handler {
	return Handler{
		s:     s,
		paths: paths,
	}
}

func (h *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "multipart/form-data" {
		httpError(w, "expected Content-Type multipart/form-data", http.StatusBadRequest)
		return
	}

	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		httpError(w, "couldn't make it out form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		httpError(w, "couldn't get file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		httpError(w, "error read file", http.StatusInternalServerError)
		return
	}

	mimeType := http.DetectContentType(data[:512])
	if mimeType != "image/png" && mimeType != "image/jpeg" {
		httpError(w, "not allowed file format", http.StatusBadRequest)
		return
	}

	id := uuid.NewString()
	ext := ""
	switch mimeType {
	case "image/jpeg":
		ext = "jpg"
	case "image/png":
		ext = "png"
	default:
		httpError(w, "not allowed file format", http.StatusBadRequest)
		return
	}
	path := fmt.Sprintf("%s/%s.%s", h.paths.UploadPath, id, ext)
	if err = os.WriteFile(path, data, 0644); err != nil {
		httpError(w, "error saving file", http.StatusInternalServerError)
		return
	}

	task := &models.Task{
		ID:      id,
		ImgPath: path,
		OutPath: fmt.Sprintf("%s/%s/", h.paths.ProcessedPath, id),
		Status:  models.StatusPending,
	}

	err = h.s.AddTask(ctx, task)
	if err != nil {
		httpError(w, "error add task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if err = json.NewEncoder(w).Encode(struct {
		Id string `json:"task-id"`
	}{
		Id: task.ID,
	}); err != nil {
		httpError(w, "error send response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) CheckStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	taskID := r.URL.Query().Get("id")
	if taskID == "" {
		httpError(w, "expected parameter id", http.StatusBadRequest)
		return
	}

	tStatus, tErr, err := h.s.GetStatus(ctx, taskID)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			httpError(w, "couldn't get task status", http.StatusBadRequest)
			return
		}
		httpError(w, "couldn't get task status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(struct {
		Status string `json:"status"`
		Err    error  `json:"error,omitempty"`
	}{
		Status: tStatus,
		Err:    tErr,
	}); err != nil {
		httpError(w, "error send response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	taskID := r.URL.Query().Get("id")
	if taskID == "" {
		httpError(w, "expected parameter id", http.StatusBadRequest)
		return
	}
	size := r.URL.Query().Get("size")
	if size == "" {
		httpError(w, "expected parameter size", http.StatusBadRequest)
		return
	}

	tStatus, tErr, err := h.s.GetStatus(ctx, taskID)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			httpError(w, "couldn't get task status", http.StatusBadRequest)
			return
		}
		httpError(w, "couldn't get task status", http.StatusInternalServerError)
		return
	}

	switch tStatus {
	case models.StatusDone:
		path := fmt.Sprintf("./storage/images/processed/%s/%s.jpg", taskID, size)
		f, err := os.Open(path)
		if err != nil {
			httpError(w, "file not found", http.StatusNotFound)
			return
		}
		defer f.Close()

		stat, err := f.Stat()
		if err != nil {
			httpError(w, "error reading file", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "image/jpeg")
		http.ServeContent(w, r, f.Name(), stat.ModTime(), f)

	case models.StatusPending:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		if err = json.NewEncoder(w).Encode(struct {
			Msg string `json:"message"`
		}{
			Msg: "task is in queue",
		}); err != nil {
			httpError(w, "error send response", http.StatusInternalServerError)
			return
		}
	case models.StatusProcessing:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		if err = json.NewEncoder(w).Encode(struct {
			Msg string `json:"message"`
		}{
			Msg: "still processing",
		}); err != nil {
			httpError(w, "error send response", http.StatusInternalServerError)
			return
		}

	case models.StatusFailed:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		if err = json.NewEncoder(w).Encode(struct {
			Err error `json:"error"`
		}{
			Err: tErr,
		}); err != nil {
			httpError(w, "error send response", http.StatusInternalServerError)
			return
		}
	}
}

func httpError(w http.ResponseWriter, msg string, status int) {
	http.Error(w, msg, status)
}
