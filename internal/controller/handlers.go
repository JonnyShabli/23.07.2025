package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/JonnyShabli/23.07.2025/internal/Service"
	"github.com/JonnyShabli/23.07.2025/internal/models"
	"github.com/JonnyShabli/23.07.2025/pkg/logster"
	"github.com/go-chi/chi/v5"
)

type HandlerInterface interface {
	AddTask(w http.ResponseWriter, r *http.Request)
	AddLinks(w http.ResponseWriter, r *http.Request)
	GetStatus(w http.ResponseWriter, r *http.Request)
}

type HandlerObj struct {
	Service Service.ZipperInterface
	Logger  logster.Logger
}

func NewHandlers(service Service.ZipperInterface, logger logster.Logger) *HandlerObj {
	return &HandlerObj{
		Service: service,
		Logger:  logger.WithField("Layer", "Handlers"),
	}
}

func (h *HandlerObj) AddTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := h.Service.AddTask(ctx)
	if err != nil {
		h.Logger.WithError(err).Errorf("Add task failed")
		ErrorResponse(w, h.Logger, err.Error(), nil)
	}
	h.Logger.Infof("Add task successfully with Id: %s", id)
	SuccessResponse(w, h.Logger, "Success", id)
}

func (h *HandlerObj) AddLinks(w http.ResponseWriter, r *http.Request) {
	linksReq := models.AddLinksRequest{}
	ctx := r.Context()
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		h.Logger.WithError(err).Infof("fail read request body")
		BadRequestResponse(w, h.Logger, "fail read request body", "")
		return
	}
	err = json.Unmarshal(reqBody, &linksReq)
	if err != nil {
		err = fmt.Errorf("fail to unmarshal request body '%w'", err)
		h.Logger.WithError(err).Infof("fail to unmarshal request body")
		BadRequestResponse(w, h.Logger, err.Error(), "")
		return
	}

	n, err := h.Service.AddLinks(ctx, linksReq.Links, linksReq.TaskId.String())
	if err != nil {
		h.Logger.WithError(err).Infof("fail to add links")
		BadRequestResponse(w, h.Logger, err.Error(), "")
		return
	}
	h.Logger.Infof("Add links successfully")
	SuccessResponse(w, h.Logger, "Success", n)
}

func (h *HandlerObj) GetStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskId := chi.URLParam(r, "task_id")
	if taskId == "" {
		h.Logger.Infof("fail to get task_id from url params")
		BadRequestResponse(w, h.Logger, fmt.Errorf("fail to get task_id").Error(), "")
		return
	}

	status, err := h.Service.GetStatus(ctx, taskId)
	if err != nil {
		h.Logger.WithError(err).Errorf("fail to get task status")
		BadRequestResponse(w, h.Logger, err.Error(), nil)
		return
	}
	h.Logger.Infof("Get task status successfully")
	SuccessResponse(w, h.Logger, "Success", status)
}
