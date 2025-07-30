package Service

import (
	"context"

	d "github.com/JonnyShabli/23.07.2025/internal/Service/downloader"
	"github.com/JonnyShabli/23.07.2025/internal/models"
	"github.com/JonnyShabli/23.07.2025/internal/repository"
	"github.com/JonnyShabli/23.07.2025/pkg/logster"
)

type ServiceInterface interface {
	AddTask(ctx context.Context) (string, error)
	AddLinks(ctx context.Context, links []string, id string) (int, error)
	GetStatus(ctx context.Context, id string) (*models.Status, error)
}

type ServiceObj struct {
	db         repository.StorageInterface
	logger     logster.Logger
	Downloader d.DownloaderInterface
}

func NewServiceObj(db repository.StorageInterface, logger logster.Logger, downloader d.DownloaderInterface) *ServiceObj {
	return &ServiceObj{
		db:         db,
		logger:     logger.WithField("Layer", "Service"),
		Downloader: downloader,
	}
}

func (s *ServiceObj) AddTask(ctx context.Context) (string, error) {
	return s.db.AddTask(ctx)
}

func (s *ServiceObj) AddLinks(ctx context.Context, links []string, id string) (int, error) {
	var result models.ValueAndError
	ch := make(chan models.ValueAndError)
	select {
	default:
	case <-ctx.Done():
		s.logger.WithError(ctx.Err()).Errorf("AddLinks: context expire")
		return 0, ctx.Err()
	}

	go func() {
		linksCount, err := s.db.AddLinks(ctx, links, id)
		if err != nil {
			s.logger.WithError(err).Errorf("AddLinks error")
			result.Err = err
			ch <- result
			return
		}
		result.Value = linksCount
		ch <- result
	}()

	select {
	case <-ctx.Done():
		s.logger.WithError(ctx.Err()).Errorf("AddLinks: context expire")
		return 0, ctx.Err()
	case v := <-ch:
		s.logger.Infof("AddLinks: added %v links", v.Value)
		go func() {
			for _, url := range links {
				job := models.DownloadJob{
					TaskId: id,
					Url:    url,
				}
				s.Downloader.AddJob(job)
			}
		}()

		return v.Value.(int), v.Err
	}
}

func (s *ServiceObj) GetStatus(ctx context.Context, id string) (*models.Status, error) {
	ch := make(chan models.ValueAndError)
	select {
	default:
	case <-ctx.Done():
		s.logger.WithError(ctx.Err()).Errorf("GetStatus: context expire")
		return nil, ctx.Err()
	}

	go func() {
		task, err := s.db.GetTask(ctx, id)
		if err != nil {
			s.logger.WithError(err).Errorf("GetTask error")
			ch <- models.ValueAndError{
				Value: nil,
				Err:   err,
			}
			return
		}

		//if task == nil {
		//	err = errors.New("task is nil")
		//	s.logger.WithError(err).Errorf("GetTask error")
		//	ch <- models.ValueAndError{
		//		Value: nil,
		//		Err:   err,
		//	}
		//	return
		//}
		s.logger.Infof("Get status succesfully with status: %s", task.Status)
		result := &models.Status{
			TaskId:        task.TaskId,
			LinksStatuses: task.LinksStatuses,
			LinksError:    task.LinksError,
			Status:        task.Status,
		}
		if task.Status == models.StatusDone {

			result.ZipPath = task.ZipPath
		}
		ch <- models.ValueAndError{
			Value: result,
			Err:   nil,
		}
	}()

	select {
	case <-ctx.Done():
		s.logger.WithError(ctx.Err()).Errorf("GetStatus: context expire")
		return nil, ctx.Err()
	case v := <-ch:
		if v.Err != nil {
			return nil, v.Err
		}
		return v.Value.(*models.Status), nil
	}
}
