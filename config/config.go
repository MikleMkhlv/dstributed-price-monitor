package config

import (
	"encoding/json"
	"fmt"
	"os"

	"go.yaml.in/yaml/v3"
)

type Config struct {
	Scheduler SchedulerConfig `yaml:"schedulerConfiguration"`
	Sources   []Sources       `yaml:"sources"`
}
type SchedulerConfig struct {
	Interval int `yaml:"interval"`
	Timeout  int `yaml:"timeout"`
}

type Sources struct {
	Type   string `yaml:"type"`
	URL    string `yaml:"url"`
	Method string `yaml:"method"`
	Data   any    `yaml:"data"`
}

type UnidataFLUL struct {
	MdmIds []string `yaml:"mdm_ids"`
}

type rawSources struct {
	Type   string    `yaml:"type"`
	URL    string    `yaml:"url"`
	Method string    `yaml:"method"`
	Data   yaml.Node `yaml:"data"`
}

func (s *Sources) UnmarshalYAML(value *yaml.Node) error {
	var raw rawSources
	if err := value.Decode(&raw); err != nil {
		return fmt.Errorf("Sources.UnmarshalYAML: decode raw source: %w", err)
	}

	s.Type = raw.Type
	s.URL = raw.URL
	s.Method = raw.Method

	switch raw.Type {
	case "unidata_fl", "unidata_ul":
		var d UnidataFLUL
		if err := raw.Data.Decode(&d); err != nil {
			return fmt.Errorf("Sources.UnmarshalYAML: decode UnidataFLUL data: %w", err)
		}
		s.Data = d

	default:
		var d map[string]any
		if err := raw.Data.Decode(&d); err != nil {
			return fmt.Errorf("Sources.UnmarshalYAML: decode generic data: %w", err)
		}
		s.Data = d
	}

	return nil
}

func (c *Config) Marshaling() ([]byte, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("config.Marshaling: error marhal config object")
	}
	return data, nil
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
