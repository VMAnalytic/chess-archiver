package config

import (
	"os"
	"path/filepath"

	"github.com/joeshaw/envdecode"
	"github.com/pkg/errors"
)

type ENV string

const gCloud ENV = "gcloud"

type Config struct {
	Env     string `env:"ENVIRONMENT"`
	Timeout int    `env:"TIMEOUT,default=60"` //in seconds

	Google struct {
		ProjectID       string `env:"GOOGLE_PROJECT_ID"`
		Secret          string `env:"GOOGLE_APPLICATION_CREDENTIALS"`
		ArchiveFolderID string `env:"ARCHIVE_FOLDER_ID"`
	}

	Lichess struct {
		APIKey      string `env:"LICHESS_API_KEY"`
		Username    string `env:"LICHESS_USERNAME"`
		LimitPerSec int    `env:"LICHESS_API_LIMIT,default=20"`
	}
}

func (c *Config) validate() error {
	var err error

	err = c.validateEnvironment()
	if err != nil {
		return errors.WithStack(err)
	}

	err = c.validateGoogleCredentials()
	if err != nil {
		return errors.WithStack(err)
	}

	err = c.validateLichessCredentials()
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (c *Config) validateEnvironment() error {
	if c.Env == "" {
		return errors.New("credentials file does not exist at the specified path")
	}

	return nil
}

func (c *Config) validateLichessCredentials() error {
	if c.Lichess.APIKey == "" {
		return errors.New("LICHESS_API_KEY ENV: required")
	}

	if c.Lichess.Username == "" {
		return errors.New("LICHESS_USERNAME ENV: required")
	}

	return nil
}

func (c *Config) validateGoogleCredentials() error {
	if c.Env != string(gCloud) {
		return nil
	}

	if len(c.Google.ProjectID) == 0 {
		return errors.New("project id is required")
	}

	if len(c.Google.Secret) == 0 {
		return errors.New("credentials file path is required")
	}

	if filepath.Ext(c.Google.Secret) != ".json" {
		return errors.New("credentials file is not json")
	}

	_, err := os.Stat(c.Google.Secret)
	if os.IsNotExist(err) {
		return errors.New("credentials file does not exist at the specified path")
	}

	return nil
}

func NewConfig() (*Config, error) {
	var config Config

	err := envdecode.Decode(&config)
	if err != nil {
		if err != envdecode.ErrNoTargetFieldsAreSet {
			return nil, errors.WithStack(err)
		}
	}

	err = config.validate()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &config, nil
}
