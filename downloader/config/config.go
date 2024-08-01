package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DownloadDir string `yaml:"downloadDir"`
	TempDir     string `yaml:"tempDir"`
	DBPath      string `yaml:"dbPath"`
	FFMPEGPath  string `yaml:"ffmpegPath"`
	Spotify     struct {
		ClientID     string `yaml:"clientID"`
		ClientSecret string `yaml:"clientSecret"`
	} `yaml:"spotify"`
	BaseURL string `yaml:"baseURL"`
	Ports   struct {
		Downloader int `yaml:"downloader"`
		Genre      int `yaml:"genre"`
		MusicAPI   int `yaml:"musicAPI"`
	} `yaml:"ports"`
	Endpoints struct {
		Genre string `yaml:"genre"`
		Meta  string `yaml:"meta"`
	} `yaml:"endpoints"`
	Concurrency struct {
		Download   int `yaml:"download"`
		Conversion int `yaml:"conversion"`
		Genre      int `yaml:"genre"`
	} `yaml:"concurrency"`
}

func LoadConfigFromFile(path string) (*Config, error) {
	if path == "" {
		path = "config/config.yaml"
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	var config Config
	dec := yaml.NewDecoder(file)
	err = dec.Decode(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
