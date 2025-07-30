package config

import (
	"os"

	"github.com/JonnyShabli/23.07.2025/internal/Service/downloader"
	"github.com/JonnyShabli/23.07.2025/internal/Service/zipper"
	pkghttp "github.com/JonnyShabli/23.07.2025/pkg/http"
	"github.com/JonnyShabli/23.07.2025/pkg/logster"
	"gopkg.in/yaml.v3"
)

type Config struct {
	HttpServer pkghttp.HTTPServer    `yaml:"http_server"`
	Logger     logster.Config        `yaml:"logger"`
	Pool       downloader.PoolConfig `yaml:"worker_pool"`
	Zipper     zipper.ZipperConfig   `yaml:"zipper"`
}

func LoadConfig(filename string, cfg interface{}) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		return err
	}

	return nil
}
