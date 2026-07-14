package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server      ServerConfig      `yaml:"server"`
	SSL         SSLConfig         `yaml:"ssl"`
	Download    DownloadConfig    `yaml:"download"`
	DataComplete DataCompleteConfig `yaml:"data_complete"`
	Netease     NeteaseConfig     `yaml:"netease"`
}

type ServerConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	AutoRedirectSSL bool   `yaml:"auto_redirect_ssl"`
}

type SSLConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Mode     string `yaml:"mode"` // upload, path, acme
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
	ACME     ACMEConfig `yaml:"acme"`
}

type ACMEConfig struct {
	Provider   string `yaml:"provider"`
	Email      string `yaml:"email"`
	Domain     string `yaml:"domain"`
	AccountID  string `yaml:"account_id"`
	SecretKey  string `yaml:"secret_key"`
	RegionID   string `yaml:"region_id"`
}

type DownloadConfig struct {
	Path         string `yaml:"path"`
	Quality      string `yaml:"quality"` // standard, high, lossless
	Concurrency  int    `yaml:"concurrency"`
	AutoMetadata bool   `yaml:"auto_metadata"`
}

type DataCompleteConfig struct {
	AutoComplete     bool   `yaml:"auto_complete"`
	Interval         int    `yaml:"interval"` // in hours
	Unit             string `yaml:"unit"`     // hour, day
	CompleteCover    bool   `yaml:"complete_cover"`
	CompleteLyrics   bool   `yaml:"complete_lyrics"`
	CompleteArtist   bool   `yaml:"complete_artist"`
}

type NeteaseConfig struct {
	APIURL string `yaml:"api_url"`
}

func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:            "0.0.0.0",
			Port:            33550,
			AutoRedirectSSL: false,
		},
		SSL: SSLConfig{
			Enabled: false,
			Mode:    "upload",
		},
		Download: DownloadConfig{
			Path:         "/music",
			Quality:      "high",
			Concurrency:  2,
			AutoMetadata: true,
		},
		Netease: NeteaseConfig{
			APIURL: "http://127.0.0.1:3000",
		},
	}
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
