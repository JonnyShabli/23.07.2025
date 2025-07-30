package workerpool

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"sync"

	"github.com/JonnyShabli/23.07.2025/internal/models"
	"github.com/JonnyShabli/23.07.2025/pkg/logster"
)

type WorkerPoolInterface interface {
	Start(ctx context.Context)
	AddJob(data models.Job)
}

type PoolConfig struct {
	NumWorkers   int      `json:"num_workers"`
	AllowedTypes []string `json:"allowed_types"`
}

type WorkerPool struct {
	numWorkers int
	In         chan models.Job
	Out        chan models.Job
	wg         *sync.WaitGroup
	logger     logster.Logger
	f          func(models.Job) models.Job
}

func NewWorkerPool(numWorkers int, logger logster.Logger, f func(models.Job) models.Job) *WorkerPool {
	logger.Infof("Pool created")
	return &WorkerPool{
		numWorkers: numWorkers,
		In:         make(chan models.Job),
		Out:        make(chan models.Job),
		wg:         &sync.WaitGroup{},
		logger:     logger,
		f:          f,
	}
}

func (wp *WorkerPool) processData(ctx context.Context, job models.Job) (models.Job, error) {
	doneCh := make(chan models.ValueAndError)
	go func() {
		doneCh <- downloadFromURL(ctx, job)
	}()

	select {
	case v := <-doneCh:
		return v.Value.(models.Job), nil
	case <-ctx.Done():
		wp.logger.Infof("Job not finished: %s", ctx.Err().Error())
		result := models.Job{
			Id:        job.Id,
			Data:      nil,
			JobStatus: "not done",
		}
		return result, ctx.Err()
	}
}

func (wp *WorkerPool) AddJob(data models.Job) {
	wp.In <- data
	wp.logger.Infof("Job added")
}

func (wp *WorkerPool) Start(ctx context.Context) {
	wg := &sync.WaitGroup{}
	wp.logger.Infof("Starting worker pool")
	for range wp.numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case v, ok := <-wp.In:
					if !ok {
						wp.logger.Infof("incoming channel closed")
						return
					}
					val, err := wp.processData(ctx, v)
					if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
						return
					}
					select {
					case wp.Out <- val:
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
		wp.logger.Infof("Stopping worker pool")
		close(wp.Out)
	}()
}

func downloadFromURL(ctx context.Context, job models.Job, allowedTypes *[]string) models.Job {
	var allowed bool = false
	result := models.Job{
		Id: job.Id,
	}
	doneCh := make(chan models.ValueAndError)
	select {
	default:
	case <-ctx.Done():
		result.Err = ctx.Err()
		return result
	}

	// Downloader
	go func() {
		response, err := http.Get(job.Data.(jobData).url)
		if err != nil {
			result.Err = err
			doneCh <- models.ValueAndError{Err: err}
			return
		}
		defer response.Body.Close()

		contentType := response.Header.Get("Content-Type")
		if contentType == "" {
			err = errors.New("content type is empty")
			result.Err = err
			doneCh <- models.ValueAndError{Err: err}
			return
		}
		mimeType, _, err := mime.ParseMediaType(contentType)
		if err != nil {
			result.Err = err
			doneCh <- models.ValueAndError{Err: err}
			return
		}
		for _, t := range *allowedTypes {
			if t == mimeType {
				allowed = true
				break
			}
		}
		if allowed {
			data, err := io.ReadAll(response.Body)
			if err != nil {
				result.Err = err
				doneCh <- models.ValueAndError{Err: err}
				return
			}

			doneCh <- models.ValueAndError{
				Value: data}
			return
		} else {
			result.Err = fmt.Errorf("%s type  is not allowed", mimeType)
			doneCh <- models.ValueAndError{Err: err}
			return
		}
	}()

	select {
	case <-ctx.Done():
		result.Err = ctx.Err()
		return result
	case data := <-doneCh:
		if data.Err != nil {
			result.Err = data.Err
			return result
		}
		result.Data = data.Value
		return result
	}
}
