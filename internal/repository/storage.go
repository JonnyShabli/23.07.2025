package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/JonnyShabli/23.07.2025/internal/models"
	"github.com/JonnyShabli/23.07.2025/pkg/logster"
	"github.com/google/uuid"
)

type StorageInterface interface {
	AddTask(ctx context.Context) (uuid.UUID, error)
	GetTask(ctx context.Context, id string) (*models.Task, error)
	AddLinks(ctx context.Context, links []string, id string) (int, error)
}
type Storage struct {
	mu     sync.RWMutex
	db     sync.Map
	logger logster.Logger
}

type valueErr struct {
	v   interface{}
	err error
}

func NewStorage(logger logster.Logger) *Storage {
	return &Storage{
		db:     sync.Map{},
		logger: logger.WithField("Layer", "Repository"),
	}
}

func (s *Storage) AddTask(ctx context.Context) (uuid.UUID, error) {
	select {
	default:
	case <-ctx.Done():
		s.logger.WithError(ctx.Err()).Errorf("AddTask: context expire")
		return uuid.Nil, ctx.Err()
	}

	doneCh := make(chan uuid.UUID)
	defer close(doneCh)
	go func() {
		task := models.Task{
			TaskId:     uuid.New(),
			Links:      make([]string, 0, 3),
			LinksDone:  make([]string, 0, 3),
			LinksError: make([]string, 0),
			Status:     models.StatusIdle,
		}
		s.db.Store(task.TaskId.String(), task)

		doneCh <- task.TaskId
	}()

	select {
	case <-ctx.Done():
		s.logger.WithError(ctx.Err()).Errorf("AddTask: context expire")
		return uuid.Nil, ctx.Err()
	case id := <-doneCh:
		s.logger.Infof("AddTask: task added with Id: %v", id)
		return id, nil
	}
}

func (s *Storage) GetTask(ctx context.Context, id string) (*models.Task, error) {
	select {
	default:
	case <-ctx.Done():
		s.logger.WithError(ctx.Err()).Errorf("GetTask: context expire")
		return nil, ctx.Err()
	}

	doneCh := make(chan valueErr)

	go func() {
		task, ok := s.db.Load(id)
		if !ok {
			err := errors.New("task not found")
			s.logger.WithError(err).Errorf("GetTask: task not found")
			doneCh <- valueErr{v: nil, err: err}
			return
		}

		s.logger.Infof("GetTask: task found with Id: %v", task.(models.Task).TaskId)
		doneCh <- valueErr{v: task, err: nil}
	}()

	select {
	case <-ctx.Done():
		s.logger.WithError(ctx.Err()).Errorf("GetTask: context expire")
		return nil, ctx.Err()
	case result := <-doneCh:
		if result.err != nil {
			s.logger.WithError(result.err).Errorf("GetTask: task not found")
			return nil, result.err
		}
		if result.v == nil {
			err := errors.New("task and error are nil")
			s.logger.WithError(err).Errorf("GetTask: task not found")
			return nil, errors.New("task not found")
		}

		s.logger.Infof("GetTask: task found, returning", result)
		task := result.v.(models.Task)
		return &task, nil
	}
}

func (s *Storage) AddLinks(ctx context.Context, links []string, id string) (int, error) {
	select {
	default:
	case <-ctx.Done():
		s.logger.WithError(ctx.Err()).Errorf("AddLink: context expire")
		return 0, ctx.Err()
	}

	doneCh := make(chan valueErr)
	defer close(doneCh)

	go func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		task, err := s.GetTask(ctx, id)
		if err != nil {
			s.logger.WithError(err).Errorf("GetTask: error")
			doneCh <- valueErr{v: nil, err: err}
			return
		}
		delta := 3 - len(task.Links)
		if len(links) > delta {
			task.Links = append(task.Links, links[:delta]...)
			s.logger.Infof("AddLinks: Added %v link(s)", delta)
			s.db.Store(task.TaskId.String(), *task)
			doneCh <- valueErr{v: delta, err: nil}
			return
		}
		task.Links = append(task.Links, links...)
		s.logger.Infof("AddLinks: Added %v link(s)", len(links))
		s.db.Store(task.TaskId.String(), *task)
		doneCh <- valueErr{v: len(links), err: nil}
	}()

	select {
	case <-ctx.Done():
		s.logger.WithError(ctx.Err()).Errorf("AddLinks: context expire")
		return 0, ctx.Err()
	case result := <-doneCh:
		if result.err != nil {
			s.logger.WithError(result.err).Errorf("AddLinks: links not added")
			return 0, result.err
		}
		if result.v == nil {
			err := errors.New("links number and error are nil")
			s.logger.WithError(err).Errorf("AddLinks: links not added")
			return 0, errors.New("task not found")
		}
		s.logger.Infof("AddLinks: links added, returning")
		return result.v.(int), nil
	}
}
