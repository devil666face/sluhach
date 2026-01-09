package command

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"sluhach/pkg/models"
	"sluhach/pkg/stt"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

type Command struct {
	cmd     *cobra.Command
	stt     *stt.Speach2Text
	manager *models.Manager
}

func (cmd *Command) load() func(*cobra.Command, []string) error {
	return func(c *cobra.Command, s []string) error {
		return cmd.manager.Load(s[0])
	}
}

func (cmd *Command) avail() func(*cobra.Command, []string) error {
	return func(c *cobra.Command, s []string) error {
		models, err := cmd.manager.Avail()
		if err != nil {
			return err
		}
		var (
			sb strings.Builder
			w  = tabwriter.NewWriter(&sb, 1, 1, 1, ' ', 0)
		)
		fmt.Fprintf(w, "#\t%s\t%s\t%s\t%s", "Lang", "Name", "Size", "Desc")
		for i, m := range models {
			fmt.Fprintf(
				w,
				"\n%d\t%s\t%s\t%s\t%s",
				i+1,
				m.Lang,
				m.Name,
				m.Size,
				m.Desc,
			)
		}
		w.Flush()
		fmt.Println(
			lipgloss.NewStyle().
				Padding(0, 1).
				Render(sb.String()),
		)
		return nil
	}
}

func (cmd *Command) remove() func(*cobra.Command, []string) error {
	return func(c *cobra.Command, s []string) error {
		return cmd.manager.Remove(s[0])
	}
}

func (cmd *Command) list() func(*cobra.Command, []string) error {
	return func(c *cobra.Command, s []string) error {
		models, err := cmd.manager.List()
		if err != nil {
			return err
		}
		var (
			sb strings.Builder
			w  = tabwriter.NewWriter(&sb, 1, 1, 1, ' ', 0)
		)
		fmt.Fprintf(w, "#\t%s\t%s", "Name", "Path")
		for i, m := range models {
			fmt.Fprintf(
				w,
				"\n%d\t%s\t%s",
				i+1,
				m.Name,
				m.Path,
			)
		}
		w.Flush()
		fmt.Println(
			lipgloss.NewStyle().
				Padding(0, 1).
				Render(sb.String()),
		)
		return nil
	}
}

func (cmd *Command) reco(model *string, wait *int, noPaste *bool) func(*cobra.Command, []string) error {
	return func(c *cobra.Command, s []string) error {
		c.Println("speak...", fmt.Sprintf("(wait %d sec for close)", *wait))
		out, err := cmd.stt.Start(filepath.Join(cmd.manager.ModelDir, *model), *wait)
		if err != nil {
			return err
		}
		c.Println("finaly:", out)
		return nil
	}
}

func New(
	_stt *stt.Speach2Text,
	_manager *models.Manager,
) *Command {
	_command := &Command{
		stt:     _stt,
		manager: _manager,
		cmd: &cobra.Command{
			Use:   "sluhach",
			Short: "Simple speach to text tool",
			RunE: func(c *cobra.Command, args []string) error {
				return c.Help()
			},
		},
	}

	var (
		modelpath string
		wait      int
		noPaste   bool
	)

	reco := &cobra.Command{
		Use:     "reco (alias:r)",
		Aliases: []string{"r"},
		Short:   "Recognize speach",
		Long:    "",
		Example: ``,
		RunE: _command.reco(
			&modelpath, &wait, &noPaste,
		),
	}
	reco.Flags().StringVarP(&modelpath, "model", "m", "vosk-model-small-ru-0.22", "Model name")
	reco.Flags().IntVarP(&wait, "wait", "w", 5, "Wait second before stop")
	reco.Flags().BoolVarP(&noPaste, "no-paste", "", false, "No paste recognized text")

	_command.cmd.AddCommand(reco)

	model := &cobra.Command{
		Use:     "model (alias:m)",
		Aliases: []string{"m"},
		Short:   "Manage models",
		Long:    "",
		Example: ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	model.AddCommand([]*cobra.Command{
		&cobra.Command{
			Use:   "list",
			Short: "List installed models",
			RunE:  _command.list(),
		},
		&cobra.Command{
			Use:   "load [name]",
			Short: "Download a model",
			Args:  cobra.ExactArgs(1),
			RunE:  _command.load(),
		},
		&cobra.Command{
			Use:   "remove [name]",
			Short: "Remove a model",
			Args:  cobra.ExactArgs(1),
			RunE:  _command.remove(),
		},
		&cobra.Command{
			Use:   "avail",
			Short: "Available models for download",
			RunE:  _command.avail(),
		},
	}...,
	)
	_command.cmd.AddCommand(model)

	return _command
}

func (cmd *Command) Execute(ctx context.Context) error {
	if err := fang.Execute(ctx, cmd.cmd); err != nil {
		return err
	}
	return nil
}
