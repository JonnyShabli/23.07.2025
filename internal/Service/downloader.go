package Service

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"

	"github.com/JonnyShabli/23.07.2025/pkg/workerpool"
)

type DownloaderInterface interface {
	DownloadFromURL(url string) ([]byte, error)
}

type Downloader struct {
	wp           workerpool.WorkerPoolInterface
	allowedTypes []string
}

func NewDownloader(allowedTypes []string, wp workerpool.WorkerPoolInterface) *Downloader {
	return &Downloader{
		wp:           wp,
		allowedTypes: allowedTypes,
	}
}

func (d *Downloader) DownloadFromURL(url string) ([]byte, error) {
	var allowed bool = false
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	contentType := response.Header.Get("Content-Type")
	if contentType == "" {
		err = errors.New("content type is empty")
		return nil, err
	}
	mimeType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, err
	}
	for _, t := range d.allowedTypes {
		if t == mimeType {
			allowed = true
			break
		}
	}
	if allowed {
		data, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		return data, nil
	} else {
		return nil, fmt.Errorf("%s type  is not allowed", mimeType)
	}
}
