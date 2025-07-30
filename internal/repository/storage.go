package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/JonnyShabli/23.07.2025/internal/models"
	"github.com/JonnyShabli/23.07.2025/pkg/logster"
	"github.com/google/uuid"
)

const serverBusyErr = "to many active tasks"

type StorageInterface interface {
	AddTask(ctx context.Context) (string, error)
	GetTask(ctx context.Context, id string) (models.Task, error)
	AddLinks(ctx context.Context, links []string, id string) (int, error)
	AddZip(ctx context.Context, data models.Task, id string) error
}
type Storage struct {
	mu     sync.RWMutex
	db     sync.Map
	logger logster.Logger
}

func NewStorage(logger logster.Logger) *Storage {
	return &Storage{
		db:     sync.Map{},
		logger: logger.WithField("Layer", "Repository"),
	}
}

func (s *Storage) AddTask(ctx context.Context) (string, error) {
	select {
	default:
	case <-ctx.Done():
		s.logger.WithError(ctx.Err()).Errorf("AddTask: context expire")
		return "", ctx.Err()
	}

	doneCh := make(chan models.ValueAndError)
	defer close(doneCh)
	var activeCount int
	go func() {
		var result models.ValueAndError
		s.mu.Lock()
		s.db.Range(func(k, v interface{}) bool {
			if v.(models.Task).Status != models.StatusDone {
				activeCount++
				fmt.Println(activeCount)
			}
			return true
		})
		s.mu.Unlock()
		if activeCount >= 3 {
			result.Err = errors.New(serverBusyErr)
			doneCh <- result
			return
		}

		task := models.Task{
			TaskId: uuid.New().String(),
			Links:  make([]string, 0),
			Status: models.StatusIdle,
		}
		s.db.Store(task.TaskId, task)

		result.Value = task.TaskId

		doneCh <- result
	}()

	select {
	case <-ctx.Done():
		s.logger.WithError(ctx.Err()).Errorf("AddTask: context expire")
		return "", ctx.Err()
	case val := <-doneCh:
		if val.Err != nil {
			s.logger.WithError(ctx.Err()).Errorf(val.Err.Error())
			return "", val.Err
		}
		id := val.Value.(string)
		s.logger.Infof("AddTask: task added with Id: %v", id)
		return id, nil
	}
}

func (s *Storage) GetTask(ctx context.Context, id string) (models.Task, error) {
	select {
	default:
	case <-ctx.Done():
		s.logger.WithError(ctx.Err()).Errorf("GetTask: context expire")
		return models.Task{}, ctx.Err()
	}

	doneCh := make(chan models.ValueAndError)

	go func() {
		task, ok := s.db.Load(id)
		if !ok {
			err := errors.New("task not found")
			s.logger.WithError(err).Errorf("GetTask: task not found")
			doneCh <- models.ValueAndError{Value: nil, Err: err}
			return
		}

		doneCh <- models.ValueAndError{Value: task, Err: nil}
	}()

	select {
	case <-ctx.Done():
		s.logger.WithError(ctx.Err()).Errorf("GetTask: context expire")
		return models.Task{}, ctx.Err()
	case result := <-doneCh:
		if result.Err != nil {
			s.logger.WithError(result.Err).Errorf("GetTask: task not found")
			return models.Task{}, result.Err
		}
		if result.Value == nil {
			err := errors.New("task and error are nil")
			s.logger.WithError(err).Errorf("GetTask: task not found")
			return models.Task{}, errors.New("task not found")
		}

		s.logger.Infof("GetTask: task found, Id: %s", result.Value.(models.Task).TaskId)
		task := result.Value.(models.Task)
		return task, nil
	}
}

func (s *Storage) AddLinks(ctx context.Context, links []string, id string) (int, error) {
	select {
	default:
	case <-ctx.Done():
		s.logger.WithError(ctx.Err()).Errorf("AddLink: context expire")
		return 0, ctx.Err()
	}

	doneCh := make(chan models.ValueAndError)
	defer close(doneCh)

	go func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		task, err := s.GetTask(ctx, id)
		if err != nil {
			s.logger.WithError(err).Errorf("GetTask: error")
			doneCh <- models.ValueAndError{Value: nil, Err: err}
			return
		}

		task.Links = append(task.Links, links...)
		task.Status = models.StatusProcessing
		s.db.Store(task.TaskId, task)
		doneCh <- models.ValueAndError{Value: len(links), Err: nil}
	}()

	select {
	case <-ctx.Done():
		s.logger.WithError(ctx.Err()).Errorf("AddLinks: context expire")
		return 0, ctx.Err()
	case result := <-doneCh:
		if result.Err != nil {
			s.logger.WithError(result.Err).Errorf("AddLinks: links not added")
			return 0, result.Err
		}
		if result.Value == nil {
			err := errors.New("links number and error are nil")
			s.logger.WithError(err).Errorf("AddLinks: links not added")
			return 0, errors.New("task not found")
		}
		s.logger.Infof("AddLinks: links added, returning")
		return result.Value.(int), nil
	}
}

func (s *Storage) AddZip(ctx context.Context, data models.Task, id string) error {
	select {
	default:
	case <-ctx.Done():
		s.logger.WithError(ctx.Err()).Errorf("AddZip: context expire")
		return ctx.Err()
	}

	doneCh := make(chan error)
	defer close(doneCh)

	go func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		task, err := s.GetTask(ctx, id)
		if err != nil {
			s.logger.WithError(err).Errorf("GetTask: error")
			doneCh <- err
			return
		}

		task.TaskId = id
		task.Status = data.Status
		task.LinksError = data.LinksError
		task.LinksStatuses = data.LinksStatuses
		task.ZipPath = data.ZipPath

		s.db.Store(task.TaskId, task)
		doneCh <- err
	}()

	select {
	case <-ctx.Done():
		s.logger.WithError(ctx.Err()).Errorf("AddZip: context expire")
		return ctx.Err()
	case err := <-doneCh:
		if err != nil {
			s.logger.WithError(err).Errorf("AddZip: failed to add")
			return err
		}

		s.logger.Infof("AddLinks: links added, returning")
		return nil
	}
}
