package zipper

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/JonnyShabli/23.07.2025/internal/models"
	"github.com/JonnyShabli/23.07.2025/internal/repository"
	"github.com/JonnyShabli/23.07.2025/pkg/logster"
)

type ZipperInterface interface {
	Start(ctx context.Context) error
}

type ZipperConfig struct {
	ArchivePath string `yaml:"archive_path"`
	MaxFiles    int    `yaml:"max_files"`
}

type Zipper struct {
	in       chan models.ZipJob
	db       repository.StorageInterface
	logger   logster.Logger
	baseBath string
	maxFiles int
}

func NewZipper(cfg ZipperConfig, logger logster.Logger, in chan models.ZipJob, db repository.StorageInterface) *Zipper {
	return &Zipper{
		baseBath: cfg.ArchivePath,
		in:       in,
		db:       db,
		logger:   logger,
		maxFiles: cfg.MaxFiles,
	}
}

func (z *Zipper) Start(ctx context.Context, logger logster.Logger) error {
	doneCh := make(chan models.ValueAndError)
	var result models.ValueAndError
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		archives := make(map[string]*os.File)
		zipWriters := make(map[string]*zip.Writer)
		counterMap := make(map[string]int)
		linksCodes := make(map[string]map[string]string)
		linksErrors := make(map[string]map[string]string)
		pathStorage := make(map[string]string)
		filesInArchive := make(map[string]map[string]int)
		var filename string

		for job := range z.in {
			z.logger.Infof("job recived from downloader")
			// сохраняем ошибки для каждой ссылки
			if job.Err != nil {
				//TODO обработать тип ошибки
				if linksErrors[job.TaskId] == nil {
					linksErrors[job.TaskId] = make(map[string]string)
				}
				linksErrors[job.TaskId][job.Url] = job.Err.Error()
				doneCh <- models.ValueAndError{}
				return
			}

			filename = job.FileName

			// сохраняем код ответа для каждой ссылки
			if linksCodes[job.TaskId] == nil {
				linksCodes[job.TaskId] = make(map[string]string)
			}
			linksCodes[job.TaskId][job.Url] = job.ResponseStatus

			// создаем файл архива, если он не существует и сохраняем соответствующий ему
			// zipWriter в zipWriters

			archivePath := filepath.Join(z.baseBath, job.TaskId+".zip")
			_, ok := archives[job.TaskId]
			if !ok {
				file, err := os.Create(archivePath)
				if err != nil {
					result.Err = fmt.Errorf("create zip file %s failed: %w", archivePath, err)
					doneCh <- result
					return
				}
				archives[job.TaskId] = file
				zipWriters[job.TaskId] = zip.NewWriter(file)
				pathStorage[job.TaskId] = archivePath
			}

			if filesInArchive[job.TaskId] == nil {
				filesInArchive[job.TaskId] = make(map[string]int)
			}
			n, ok := filesInArchive[job.TaskId][job.FileName]
			filesInArchive[job.TaskId][filename]++
			if ok {
				ext := filepath.Ext(job.FileName)
				base := strings.TrimSuffix(job.FileName, ext)
				filename = fmt.Sprintf("%s_%d.%s", base, n, ext)
			}

			// создаем файл в соответствующем taskId архиве
			zipFileWriter, err := zipWriters[job.TaskId].Create(filename)
			if err != nil {
				result.Err = fmt.Errorf("create zipFileWriter %s failed: %w", filename, err)
				doneCh <- result
				return
			}

			// записываем файл в архив
			buf := bytes.NewBuffer(*job.Data)
			_, err = io.Copy(zipFileWriter, buf)
			if err != nil {
				result.Err = fmt.Errorf("copy zipFileWriter %s failed: %w", filename, err)
				doneCh <- result
				return
			}

			// считаем кол-во запакованных файлов для каждой таски
			counterMap[job.TaskId]++

			// если скачано 3 файла, то сбрасываем счетчик и обновляем Task в хранилище
			if counterMap[job.TaskId] == z.maxFiles {
				delete(counterMap, job.TaskId)

				// закрываем zipWriter
				err = zipWriters[job.TaskId].Close()
				if err != nil {
					result.Err = fmt.Errorf("close zipWriter %s failed: %w", filename, err)
					doneCh <- result
					return
				}
				// удаляем zipWriter из мапы
				delete(zipWriters, job.TaskId)

				// закрываем файл
				err = archives[job.TaskId].Close()
				if err != nil {
					result.Err = fmt.Errorf("close archive file %s.zip failed: %w", archives[job.TaskId].Name(), err)
					doneCh <- result
					return
				}
				// удаляем zip файл из мапы
				delete(archives, job.TaskId)

				data := models.Task{
					LinksStatuses: linksCodes[job.TaskId],
					LinksError:    linksErrors[job.TaskId],
					ZipPath:       pathStorage[job.TaskId],
					Status:        models.StatusDone,
				}

				// обновляем запись в хранилище
				err = z.db.AddZip(ctx, data, job.TaskId)
				if err != nil {
					result.Err = fmt.Errorf("add zip file %s to db failed: %w", filename, err)
					doneCh <- result
					return
				}
			}
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case v := <-doneCh:
		if v.Err != nil {
			return v.Err
		}
		return nil
	}
}

func exist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
