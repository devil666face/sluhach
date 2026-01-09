package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sluhach/pkg/fs"
)

const (
	modelDir = "~/.local/share/sluhach"
)

type Config struct {
	SessionType string
	ModelDir    string
}

func getModelDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home dir: %w", err)
	}

	expanded := modelDir
	if len(modelDir) >= 2 && modelDir[:2] == "~/" {
		expanded = filepath.Join(home, modelDir[2:])
	}

	_modelDir, err := filepath.Abs(expanded)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for modelDir: %w", err)
	}
	return _modelDir, nil
}

func New() (*Config, error) {
	_sessionType := os.Getenv("XDG_SESSION_TYPE")
	if _sessionType == "" {
		return nil, fmt.Errorf("XDG_SESSION_TYPE is not set")
	}
	_modelDir, err := getModelDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get dir with models: %w", err)
	}
	if err := fs.CreateDirs(_modelDir); err != nil {
		return nil, fmt.Errorf("failed to create model dirs: %w", err)
	}
	return &Config{
		SessionType: _sessionType,
		ModelDir:    _modelDir,
	}, nil
}
