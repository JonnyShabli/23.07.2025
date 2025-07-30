package downloader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"sync"
	"time"

	"github.com/JonnyShabli/23.07.2025/internal/models"
	"github.com/JonnyShabli/23.07.2025/pkg/logster"
)

type DownloaderInterface interface {
	StartDownloader(ctx context.Context)
	AddJob(job models.DownloadJob)
}

type PoolConfig struct {
	timeout      time.Duration `yaml:"timeout"`
	NumWorkers   int           `yaml:"num_workers"`
	AllowedTypes []string      `yaml:"allowed_types"`
	MaxFileSize  int64         `yaml:"max_file_size"`
}

type Downloader struct {
	client       http.Client
	numWorkers   int
	In           chan models.DownloadJob
	Out          chan models.ZipJob
	wg           *sync.WaitGroup
	logger       logster.Logger
	allowedTypes []string
	maxFileSize  int64
}

func NewDownloader(cfg PoolConfig, logger logster.Logger) *Downloader {
	logger.Infof("Downloader pool created")
	client := http.Client{
		Timeout: cfg.timeout,
	}
	return &Downloader{
		client:       client,
		numWorkers:   cfg.NumWorkers,
		In:           make(chan models.DownloadJob),
		Out:          make(chan models.ZipJob),
		wg:           &sync.WaitGroup{},
		logger:       logger,
		allowedTypes: cfg.AllowedTypes,
		maxFileSize:  cfg.MaxFileSize,
	}
}

func (d *Downloader) AddJob(job models.DownloadJob) {
	d.In <- job
	d.logger.Infof("Job added, %-v", job)
}

func (d *Downloader) StartDownloader(ctx context.Context) {
	d.logger.Infof("Starting downloader worker pool")
	d.wg.Add(d.numWorkers)
	for range d.numWorkers {
		go func() {
			defer d.wg.Done()
			for job := range d.In {
				select {
				case <-ctx.Done():
					return
				case d.Out <- d.downloadFromURL(ctx, job, d.allowedTypes):

				}
			}
		}()
	}

	go func() {
		d.wg.Wait()
		d.logger.Infof("Stopping downloader worker pool")
		close(d.Out)
	}()
}

func (d *Downloader) processData(ctx context.Context, job models.DownloadJob) models.ZipJob {
	doneCh := make(chan models.ZipJob)
	result := models.ZipJob{
		TaskId: job.TaskId,
	}
	go func() {
		doneCh <- d.downloadFromURL(ctx, job, d.allowedTypes)
	}()

	select {
	case v := <-doneCh:
		result.Data = v.Data
		result.ResponseStatus = v.ResponseStatus
		return result
	case <-ctx.Done():
		d.logger.Infof("Job not finished: %s", ctx.Err().Error())
		result.Err = ctx.Err()
		return result
	}
}

func (d *Downloader) downloadFromURL(ctx context.Context, job models.DownloadJob, allowedTypes []string) models.ZipJob {
	var allowed bool = false
	result := models.ZipJob{
		TaskId: job.TaskId,
		Url:    job.Url,
	}
	doneCh := make(chan models.ZipJob)
	select {
	default:
	case <-ctx.Done():
		result.Err = ctx.Err()
		return result
	}

	go func() {
		// получаем заголовки
		respHead, err := d.client.Head(job.Url)
		if err != nil {
			result.Err = err
			doneCh <- result
			return
		}
		defer respHead.Body.Close()
		// проверяем размер файла
		contentLength := respHead.ContentLength
		if contentLength > d.maxFileSize {
			result.Err = fmt.Errorf("file size %d exceeds maximum allowed size %d", contentLength, d.maxFileSize)
			doneCh <- result
			return
		}
		// запрашиваем файл
		response, err := d.client.Get(job.Url)
		if err != nil {
			result.Err = err
			doneCh <- result
			return
		}
		defer response.Body.Close()

		// получаем статус код ответа
		result.ResponseStatus = response.Status
		// получаем имя файл
		parsedURL, err := url.Parse(job.Url)
		if err != nil {
			result.Err = err
			doneCh <- result
			return
		}
		filename := filepath.Base(parsedURL.Path)
		if filename == "." || filename == "/" {
			filename = "unnamed_file"
		}
		result.FileName = filename

		// получаем тип файла
		contentType := response.Header.Get("Content-Type")
		if contentType == "" {
			err = errors.New("content type is empty")
			result.Err = err
			doneCh <- result
			return
		}

		mimeType, _, err := mime.ParseMediaType(contentType)
		if err != nil {
			result.Err = err
			doneCh <- result
			return
		}
		for _, t := range allowedTypes {
			if t == mimeType {
				allowed = true
				break
			}
		}

		// проверяем допустимость типа файла
		if allowed {
			data, err := io.ReadAll(response.Body)
			if err != nil {
				result.Err = err
				doneCh <- result
				return
			}
			result.Data = &data
			doneCh <- result
			return
		} else {
			result.Err = fmt.Errorf("%s type  is not allowed", mimeType)
			doneCh <- result
			return
		}
	}()

	select {
	case <-ctx.Done():
		result.Err = fmt.Errorf("job not finished: %w", ctx.Err().Error())
		return result
	case v := <-doneCh:
		return v
	}
}
