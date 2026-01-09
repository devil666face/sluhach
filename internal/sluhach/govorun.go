package sluhach

import (
	"context"
	"fmt"

	"sluhach/internal/command"
	"sluhach/internal/config"

	"sluhach/pkg/models"
	"sluhach/pkg/stt"
)

type Sluhach struct {
	cmd *command.Command
}

func New(
	modelpath string,
) (*Sluhach, error) {
	_config, err := config.New()
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}
	_stt := stt.New()
	_manager := models.New(_config.ModelDir)
	_cmd := command.New(
		_stt,
		_manager,
	)
	return &Sluhach{
		cmd: _cmd,
	}, nil
}

func (s *Sluhach) Start() error {
	return s.cmd.Execute(context.Background())
}
