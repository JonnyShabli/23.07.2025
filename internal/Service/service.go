package Service

import (
	"context"
	"errors"

	"github.com/JonnyShabli/23.07.2025/internal/models"
	"github.com/JonnyShabli/23.07.2025/internal/repository"
	"github.com/JonnyShabli/23.07.2025/pkg/logster"
	"github.com/JonnyShabli/23.07.2025/pkg/workerpool"
	"github.com/google/uuid"
)

type ServiceInterface interface {
	AddTask(ctx context.Context) (uuid.UUID, error)
	AddLinks(ctx context.Context, links []string, id string) (int, error)
	GetStatus(ctx context.Context, id string) (*models.Status, error)
}

type ServiceObj struct {
	db         repository.StorageInterface
	logger     logster.Logger
	downloader DownloaderInterface
	zipper     ZipperInterface
}

func NewServiceObj(db repository.StorageInterface, logger logster.Logger, allowedTypes []string, downloader, zipper workerpool.WorkerPoolInterface) *ServiceObj {
	return &ServiceObj{
		db:         db,
		logger:     logger.WithField("Layer", "Service"),
		downloader: NewDownloader(allowedTypes, downloader),
		zipper:     zipper,
	}
}

func (z *ServiceObj) AddTask(ctx context.Context) (uuid.UUID, error) {
	return z.db.AddTask(ctx)
}

func (z *ServiceObj) AddLinks(ctx context.Context, links []string, id string) (int, error) {
	linksCount, err := z.db.AddLinks(ctx, links, id)
	if err != nil {
		z.logger.WithError(err).Errorf("AddLinks error")
		return 0, err
	}

	return linksCount, nil
}

func (z *ServiceObj) GetStatus(ctx context.Context, id string) (*models.Status, error) {
	task, err := z.db.GetTask(ctx, id)
	if err != nil {
		z.logger.WithError(err).Errorf("GetTask error")
		return nil, err
	}

	if task == nil {
		err = errors.New("task is nil")
		z.logger.WithError(err).Errorf("GetTask error")
		return nil, err
	}

	if task.Status == models.StatusDone {
		status := &models.Status{
			TaskId:     task.TaskId,
			LinksDone:  task.Links,
			LinksError: task.LinksError,
			Status:     task.Status,
			Zip:        task.Zip,
		}
		z.logger.Infof("Get status succesfully with status: %s", task.Status)
		return status, nil
	}

	status := &models.Status{
		TaskId:     task.TaskId,
		LinksDone:  task.Links,
		LinksError: task.LinksError,
		Status:     task.Status,
	}
	z.logger.Infof("Get status succesfully with status: %s", task.Status)
	return status, nil
}
