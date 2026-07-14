package config

import (
	"fmt"
	"os"
	"time"

	"go.yaml.in/yaml/v3"
)

type Config struct {
	Scheduler SchedulerConfig `yaml:"schedulerConfiguration"`
	Sources   []Sources       `yaml:"sources"`
}
type SchedulerConfig struct {
	Interval int           `yaml:"interval"`
	Timeout  time.Duration `yaml:"timeout"`
}

type Sources struct {
	URL    string   `yaml:"url"`
	Method string   `yaml:"method"`
	Data   []string `yaml:"data"`
}

func MustLoadConfig(pathConfig string) *Config {
	if pathConfig == "" {
		panic("config.MustLoadConfig: path configuration is Empty")
	}
	_, err := os.Stat(pathConfig)
	if err != nil {
		if os.IsNotExist(err) {
			panic(fmt.Sprintf("config.MustLoadConfig: configuration file {%s} is not exsist", pathConfig))
		}
	}

	cfg := &Config{}

	file, err := os.Open(pathConfig)
	if err != nil {
		panic("config.MustLoadConfig: could not open configuration file:" + pathConfig)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err = decoder.Decode(cfg); err != nil {
		panic("config.MustLoadConfig: could not decode configuretion YAML:" + err.Error())
	}

	return cfg
}
