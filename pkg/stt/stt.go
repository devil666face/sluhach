package stt

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"sluhach/pkg/fs"

	vosk "github.com/alphacep/vosk-api/go"
	"github.com/gordonklaus/portaudio"
)

type Speach2Text struct {
	modelDir string
}

// sudo apt-get install portaudio19-dev
// sudo dnf install portaudio-devel
func New(
	_modelDir string,
) *Speach2Text {
	vosk.SetLogLevel(-1)

	return &Speach2Text{
		modelDir: _modelDir,
	}
}

func (s *Speach2Text) Start(
	path string,
	wait int,
) (string, error) {
	path = filepath.Join(s.modelDir, path)

	if err := fs.Exists(path); err != nil {
		return "", fmt.Errorf("source vosk model not found: %w", err)
	}

	model, err := vosk.NewModel(path)
	if err != nil {
		return "", fmt.Errorf("failed to load vosk model: %w", err)
	}
	defer model.Free()

	rec, err := vosk.NewRecognizer(model, 16000.0)
	if err != nil {
		return "", fmt.Errorf("failed to create recognizer: %w", err)
	}

	defer rec.Free()

	if err := portaudio.Initialize(); err != nil {
		return "", fmt.Errorf("failed to initilaze portaudio: %w", err)
	}
	defer portaudio.Terminate()

	var (
		_err      error
		collected []string
		lastTime  = time.Now()
	)

	stream, err := portaudio.OpenDefaultStream(1, 0, 16000, 8000, func(in []int16) {
		if _err != nil {
			return
		}

		data := int16ToBytes(in)
		if b := rec.AcceptWaveform(data); b > 0 {
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(rec.Result()), &result); err != nil {
				_err = fmt.Errorf("failed to unmarshal result: %w", err)
				return
			}
			text, ok := result["text"].(string)
			if !ok {
				_err = fmt.Errorf("unexpected result format: no 'text' field")
				return
			}
			if text != "" {
				// fmt.Println("recognized:", text)
				collected = append(collected, text)
				lastTime = time.Now()
			}
		} else {
			var partial map[string]interface{}
			if err := json.Unmarshal([]byte(rec.PartialResult()), &partial); err != nil {
				_err = fmt.Errorf("failed to unmarshal partial result: %w", err)
				return
			}
			text, ok := partial["partial"].(string)
			if !ok {
				return
			}
			if text != "" {
				// fmt.Println("temp:", text)
				lastTime = time.Now()
			}
		}
	})
	if err != nil {
		return "", fmt.Errorf("failed to open audio channel: %w", err)
	}
	defer stream.Close()

	if err := stream.Start(); err != nil {
		return "", fmt.Errorf("failed to listen: %w", err)
	}
	defer stream.Stop()

	for {
		if _err != nil {
			return "", _err
		}
		if time.Since(lastTime) >= time.Duration(wait)*time.Second {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if _err != nil {
		return "", _err
	}

	return strings.Join(collected, "\n"), nil
}

func int16ToBytes(input []int16) []byte {
	output := make([]byte, len(input)*2)
	for i, v := range input {
		binary.LittleEndian.PutUint16(output[i*2:], uint16(v))
	}
	return output
}
