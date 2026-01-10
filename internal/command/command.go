package command

import (
	"context"
	"fmt"
	"strings"
	"text/tabwriter"

	"sluhach/pkg/clip"
	"sluhach/pkg/models"
	"sluhach/pkg/notify"
	"sluhach/pkg/stt"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

const (
	recordStarted  = "▶️ recording started"
	recordFinished = "⏹️ recording finished"
	copiedToClip   = "📋 text copied to clipboard"
	listen         = "🎤 listening"
)

type Command struct {
	cmd     *cobra.Command
	stt     *stt.Speach2Text
	manager *models.Manager
}

func (cmd *Command) reco(model *string, wait *int, noPaste *bool) func(*cobra.Command, []string) error {
	return func(c *cobra.Command, s []string) error {
		c.Println(listen, fmt.Sprintf("(waiting for %d seconds of silence to stop)", *wait))
		if err := notify.Notify(recordStarted, listen); err != nil {
			return err
		}

		out, err := cmd.stt.Start(*model, *wait)
		if err != nil {
			return err
		}
		if out != "" {
			c.Println(out)

			if !*noPaste {
				if err := clip.Clip(out); err != nil {
					return err
				}
				c.Println(copiedToClip)
			}

			if err := notify.Notify(recordFinished, out); err != nil {
				return err
			}
		}
		return nil
	}
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
		c.Println(
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
		c.Println(
			lipgloss.NewStyle().
				Padding(0, 1).
				Render(sb.String()),
		)
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
			Short: "Simple speech-to-text tool",
			Long: `sluhach is a simple speech-to-text CLI tool based on Vosk.

It can:
  - record audio from the default microphone
  - recognize speech using locally installed Vosk models
  - copy recognized text to the clipboard
  - show desktop notifications about recording status

Main commands:
  sluhach reco
      Record from microphone and recognize speech.

  sluhach model ...
      Manage speech recognition models (list, load, remove, avail).`,
			Example: `  sluhach reco
  sluhach reco -m vosk-model-small-ru-0.22
  sluhach model list
  sluhach model avail`,
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
		Short:   "Recognize speech",
		Long: `Records speech from the microphone and recognizes it using the selected model.

By default:
  - uses the "vosk-model-small-ru-0.22" model
  - stops recording after several seconds of silence (see the --wait flag)
  - prints the recognized text to the terminal and copies it to the clipboard

Examples:
  sluhach reco
      Record from microphone, recognize and copy the text.

  sluhach reco -m vosk-model-en-us-0.22
      Use the English model.

  sluhach reco -w 8
      Wait up to 8 seconds of silence before stopping recording.

  sluhach reco --no-paste
      Do not copy the result to the clipboard, only print it to the terminal.`,
		Example: `  sluhach reco
  sluhach reco -m vosk-model-small-ru-0.22
  sluhach reco -m vosk-model-en-us-0.22 -w 8
  sluhach reco --no-paste`,
		RunE: _command.reco(
			&modelpath, &wait, &noPaste,
		),
	}
	reco.Flags().StringVarP(&modelpath, "model", "m", "vosk-model-small-ru-0.22", "Model name")
	reco.Flags().IntVarP(&wait, "wait", "w", 5, "Seconds of silence before stop")
	reco.Flags().BoolVarP(&noPaste, "no-paste", "", false, "Do not copy recognized text to clipboard")

	_command.cmd.AddCommand(reco)

	model := &cobra.Command{
		Use:     "model (alias:m)",
		Aliases: []string{"m"},
		Short:   "Manage models",
		Long: `Manage Vosk speech recognition models.

  sluhach model list
      Show models that are already downloaded and available locally.

  sluhach model load [name]
      Download and unpack a model with the given name into the models directory.

  sluhach model remove [name]
      Remove a previously downloaded model from the models directory.

  sluhach model avail
      Show models that are available for download from the remote repository.`,
		Example: `  sluhach model list
  sluhach model load vosk-model-small-ru-0.22
  sluhach model remove vosk-model-small-ru-0.22
  sluhach model avail`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	model.AddCommand([]*cobra.Command{
		{
			Use:   "list",
			Short: "List installed models",
			Long: `List models that are already downloaded and available locally.

This command inspects the models directory and prints all models that can be
used for recognition without additional downloads.`,
			Example: `  sluhach model list`,
			RunE:    _command.list(),
		},
		{
			Use:   "load [name]",
			Short: "Download a model",
			Long: `Download and unpack a Vosk model by name.

The model archive is fetched from the configured base URL and extracted into
the local models directory. If the model is already present, the command
returns an error.`,
			Args: cobra.ExactArgs(1),
			Example: `  sluhach model load vosk-model-small-ru-0.22
  sluhach model load vosk-model-en-us-0.22`,
			RunE: _command.load(),
		},
		{
			Use:   "remove [name]",
			Short: "Remove a model",
			Long: `Remove a previously downloaded model from the local models directory.

This does not affect any remote resources, only local files are deleted.`,
			Args:    cobra.ExactArgs(1),
			Example: "  sluhach model remove vosk-model-small-ru-0.22",
			RunE:    _command.remove(),
		},
		{
			Use:   "avail",
			Short: "List models available for download",
			Long: `Show models that are available for download from the remote repository.

The list is fetched from the remote server and includes language, name, size
and a short description for each model.`,
			Example: "  sluhach model avail",
			RunE:    _command.avail(),
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
