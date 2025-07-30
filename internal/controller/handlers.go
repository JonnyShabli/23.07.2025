package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/JonnyShabli/23.07.2025/internal/Service"
	"github.com/JonnyShabli/23.07.2025/internal/models"
	"github.com/JonnyShabli/23.07.2025/pkg/logster"
	"github.com/go-chi/chi/v5"
)

const archivesDir = "./tmp/archives/"

type HandlerInterface interface {
	AddTask(w http.ResponseWriter, r *http.Request)
	AddLinks(w http.ResponseWriter, r *http.Request)
	GetStatus(w http.ResponseWriter, r *http.Request)
	DownloadZip(w http.ResponseWriter, r *http.Request)
}

type HandlerObj struct {
	Service  Service.ServiceInterface
	Logger   logster.Logger
	hostname string
}

func NewHandlers(service Service.ServiceInterface, logger logster.Logger, hostName string) *HandlerObj {
	return &HandlerObj{
		Service:  service,
		Logger:   logger.WithField("Layer", "Handlers"),
		hostname: hostName,
	}
}

func (h *HandlerObj) AddTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := h.Service.AddTask(ctx)
	if err != nil {
		h.Logger.WithError(err).Errorf("Add task failed")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.Logger.Infof("Add task successfully with Id: %s", id)
	SuccessDataResponse(w, h.Logger, "Success", id)
}

func (h *HandlerObj) AddLinks(w http.ResponseWriter, r *http.Request) {
	linksReq := models.AddLinksRequest{}
	ctx := r.Context()
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		h.Logger.WithError(err).Infof("fail read request body")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(reqBody, &linksReq)
	if err != nil {
		err = fmt.Errorf("fail to unmarshal request body '%w'", err)
		h.Logger.WithError(err).Infof("fail to unmarshal request body")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	n, err := h.Service.AddLinks(ctx, linksReq.Links, linksReq.TaskId)
	if err != nil {
		h.Logger.WithError(err).Infof("fail to add links")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.Logger.Infof("Add links successfully")
	SuccessDataResponse(w, h.Logger, "Success", n)
}

func (h *HandlerObj) GetStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskId := chi.URLParam(r, "task_id")
	if taskId == "" {
		h.Logger.Infof("fail to get task_id from url params")
		http.Error(w, "fail to get task_id", http.StatusBadRequest)
		return
	}

	status, err := h.Service.GetStatus(ctx, taskId)
	if err != nil {
		h.Logger.WithError(err).Errorf("fail to get task status")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fullUrl := fmt.Sprintf("http://%s/download/%s", h.hostname, status.ZipPath)
	status.ZipPath = fullUrl
	h.Logger.Infof("Get task status successfully")
	SuccessDataResponse(w, h.Logger, "Success", status)
}

func (h *HandlerObj) DownloadZip(w http.ResponseWriter, r *http.Request) {
	fileName := path.Base(r.URL.Path)
	if fileName == "" {
		http.Error(w, "filename not specified", http.StatusBadRequest)
		return
	}

	// Проверяем существование файла
	filePath := filepath.Join(archivesDir, fileName)

	if info, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Println(info)
		http.Error(w, "archive not found", http.StatusNotFound)
		return
	}

	// Отдаём файл
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	http.ServeFile(w, r, archivesDir+fileName)
}
