package notifier

import (
	"fmt"
	"os"
	"time"

	"go.yaml.in/yaml/v3"
)

type NotifierConfig struct {
	Channels []ChannelConfig `yaml:"channels"`
}

type ChannelConfig struct {
	Type    string `yaml:"type"`
	Enabled bool   `yaml:"enabled"`

	Telegram *TelegramConfig `yaml:"telegram,omitempty"`
	Webhook  *WebhookConfig  `yaml:"webhook,omitempty"`
	Email    *EmailConfig    `yaml:"email,omitempty"`
}

type TelegramConfig struct {
	Token  string `yaml:"token"`
	ChatID string `yaml:"chat_id"`
}

type WebhookConfig struct {
	URL     string            `yaml:"url"`
	Method  string            `yaml:"method"`
	Timeout time.Duration     `yaml:"timeout"`
	Headers map[string]string `yaml:"headers"`
}

type EmailConfig struct {
	SMTPHost string `yaml:"smtp_host"`
	SMTPPort int    `yaml:"smtp_port"`
	From     string `yaml:"from"`
}

func LoadNotifierConfig(pathConf string) (*NotifierConfig, error) {
	if pathConf == "" {
		panic("notifier.config.LoadNotifierConfig: path configuration is Empty")
	}
	_, err := os.Stat(pathConf)
	if err != nil {
		if os.IsNotExist(err) {
			panic(fmt.Sprintf("notifier.config.LoadNotifierConfig: configuration file {%s} is not exsist", pathConf))
		}
	}

	cfg := NotifierConfig{}

	file, err := os.Open(pathConf)
	if err != nil {
		return nil, fmt.Errorf("notifier.config.LoadNotifierConfig: %v", err)
	}
	decoder := yaml.NewDecoder(file)
	if err = decoder.Decode(&cfg); err != nil {
		panic("notifier.config.LoadNotifierConfig: could not decode configuretion YAML:" + err.Error())
	}

	return &cfg, nil
}
