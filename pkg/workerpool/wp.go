package workerpool

import (
	"context"
	"errors"
	"sync"

	"github.com/JonnyShabli/23.07.2025/internal/models"
	"github.com/JonnyShabli/23.07.2025/pkg/logster"
)

type WorkerPoolInterface interface {
	Start(ctx context.Context)
	AddJob(data interface{})
}

type WorkerPool struct {
	numWorkers int
	in         chan models.Job
	out        chan models.Job
	wg         *sync.WaitGroup
	logger     logster.Logger
	f          func(models.Job) models.Job
}

func NewWorkerPool(numWorkers int, logger logster.Logger) *WorkerPool {
	return &WorkerPool{
		numWorkers: numWorkers,
		in:         make(chan models.Job),
		out:        make(chan models.Job),
		wg:         &sync.WaitGroup{},
		logger:     logger,
	}
}

func (wp *WorkerPool) processData(ctx context.Context, job models.Job) (models.Job, error) {
	ch := make(chan models.Job)
	go func() {
		ch <- wp.f(job)
		wp.logger.Infof("Finish proccessing job")
		close(ch)
	}()

	select {
	case v := <-ch:
		return v, nil
	case <-ctx.Done():
		wp.logger.Infof("Job not finish: %s", ctx.Err().Error())
		result := models.Job{
			Id:        job.Id,
			Data:      nil,
			JobStatus: "not done",
		}
		return result, ctx.Err()
	}
}

func (wp *WorkerPool) AddJob(data models.Job) {
	wp.in <- data
	wp.logger.Infof("Job added")
}

func (wp *WorkerPool) Start(ctx context.Context, name string) {
	wg := &sync.WaitGroup{}
	wp.logger.Infof("Starting %s worker pool", name)
	for range wp.numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case v, ok := <-wp.in:
					if !ok {
						wp.logger.Infof("incoming channel closed")
						return
					}
					val, err := wp.processData(ctx, v)
					if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
						return
					}
					select {
					case wp.out <- val:
					case <-ctx.Done():
						return
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		wp.logger.Infof("Stopping %s worker pool", name)
		close(wp.out)
	}()
}
