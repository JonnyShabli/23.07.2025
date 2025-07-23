package config

import (
	"os"
	"time"

	"github.com/JonnyShabli/23.07.2025/pkg/logster"
	"gopkg.in/yaml.v3"
)

type Config struct {
	HTTPClient `yaml:"httpClient"`
	Logger     logster.Config `yaml:"logger"`
}

type HTTPClient struct {
	Timeout time.Duration `yaml:"timeout"`
	Addr    string        `yaml:"Addr"`
	Port    string        `yaml:"Port"`
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
