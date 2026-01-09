package models

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"sluhach/pkg/fs"

	"github.com/PuerkitoBio/goquery"
)

const (
	_base = "https://alphacephei.com/vosk/models"
)

type Manager struct {
	base     string
	ModelDir string
	client   *http.Client
}

type Model struct {
	Lang, Name, Size, Desc string
	Path                   string
}

func New(
	modelDir string,
) *Manager {
	// стандартный http.Client по умолчанию использует системные настройки,
	// включая переменные окружения HTTP_PROXY / HTTPS_PROXY / NO_PROXY.
	client := &http.Client{}

	return &Manager{
		base:     _base,
		ModelDir: modelDir,
		client:   client,
	}
}

// Load скачивает zip‑архив модели по имени, распаковывает его в ModelDir
// и удаляет временный файл.
func (m *Manager) Load(model string) error {
	// формируем URL архива
	url := fmt.Sprintf("%s/%s.zip", strings.TrimRight(m.base, "/"), model)

	tmp, err := os.CreateTemp(m.ModelDir, model+"-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	path := tmp.Name()
	defer func() {
		tmp.Close()
		_ = os.Remove(path)
	}()

	// создаём запрос
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	// стандартные заголовки для всех запросов
	defaultHeaders(req)

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download model: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download model, status: %s", resp.Status)
	}

	if resp.ContentLength > 0 {
		var downloaded int64
		total := resp.ContentLength

		buf := make([]byte, 32*1024)
		for {
			n, _err := resp.Body.Read(buf)
			if n > 0 {
				if _, err := tmp.Write(buf[:n]); err != nil {
					return fmt.Errorf("failed to save model archive: %w", err)
				}
				downloaded += int64(n)
				percent := float64(downloaded) * 100 / float64(total)
				remaining := total - downloaded
				fmt.Printf("\rdownloaded: %.1f%% (%d / %d bytes, %d bytes remaining)", percent, downloaded, total, remaining)
			}
			if _err == io.EOF {
				break
			}
			if _err != nil {
				return fmt.Errorf("failed to read response body: %w", _err)
			}
		}
		fmt.Println()
	} else {
		// если размер неизвестен — просто копируем без процентов
		if _, err := io.Copy(tmp, resp.Body); err != nil {
			return fmt.Errorf("failed to save model archive: %w", err)
		}
	}

	if err := tmp.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := fs.Unzip(path, m.ModelDir); err != nil {
		return fmt.Errorf("failed to unzip model: %w", err)
	}

	return nil
}

func (m *Manager) Avail() ([]Model, error) {
	// создаём запрос, чтобы можно было навесить те же заголовки
	req, err := http.NewRequest(http.MethodGet, m.base, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	defaultHeaders(req)

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch models page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse html: %w", err)
	}

	table := doc.Find("table.table.table-bordered").First()
	if table.Length() == 0 {
		return nil, fmt.Errorf("models table not found")
	}

	// получаем tbody из таблицы (если он есть)
	tbody := table.Find("tbody")
	if tbody.Length() == 0 {
		return nil, fmt.Errorf("models content not found")
	}

	var (
		models []Model
		lang   string
	)

	tbody.Find("tr").Each(func(i int, tr *goquery.Selection) {
		var (
			tds []*goquery.Selection
		)
		tr.Find("td").Each(func(i int, td *goquery.Selection) {
			if strings.TrimSpace(td.Text()) == "" {
				return
			}
			tds = append(tds, td)
		})
		if len(tds) < 5 {
			if len(tds) > 0 {
				lang = strings.TrimSpace(tds[0].Text())
			}
			return
		}
		models = append(models,
			Model{
				Lang: strings.TrimSpace(lang),
				Name: strings.TrimSpace(tds[0].Text()),
				Size: strings.TrimSpace(tds[1].Text()),
				Desc: strings.TrimSpace(tds[3].Text()),
			},
		)
	})

	if len(models) == 0 {
		return nil, fmt.Errorf("not found any model")
	}

	return models, nil
}

func (m *Manager) Remove(model string) error {
	models, err := fs.List(m.ModelDir)
	if err != nil {
		return fmt.Errorf("failed to get models list: %w", err)
	}
	if !slices.Contains(models, model) {
		return fmt.Errorf("model %s not exists", model)
	}
	if err := fs.Remove(path.Join(m.ModelDir, model)); err != nil {
		return fmt.Errorf("failed to remove: %w", err)
	}
	return nil
}

func (m *Manager) List() ([]Model, error) {
	_models, err := fs.List(m.ModelDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get models list: %w", err)
	}
	var models []Model
	for _, name := range _models {
		models = append(models, Model{
			Name: name,
			Path: filepath.Join(m.ModelDir, name),
		})
	}
	return models, nil
}

func defaultHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:146.0) Gecko/20100101 Firefox/146.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.8,en-US;q=0.5,en;q=0.3")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Priority", "u=0, i")
	// req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
}
