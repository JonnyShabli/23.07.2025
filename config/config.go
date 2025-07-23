package config

import (
	"os"

	pkghttp "github.com/JonnyShabli/23.07.2025/pkg/http"
	"github.com/JonnyShabli/23.07.2025/pkg/logster"
	"gopkg.in/yaml.v3"
)

type Config struct {
	pkghttp.HTTPClient `yaml:"httpClient"`
	Logger             logster.Config `yaml:"logger"`
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
