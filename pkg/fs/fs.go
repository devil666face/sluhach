package fs

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Exists(path string) error {
	_, err := os.Stat(path)
	return err
}

func CreateDirs(path string) error {
	return os.MkdirAll(path, 0o755)
}

func List(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(entries))
	for _, e := range entries {
		result = append(result, e.Name())
	}
	return result, nil
}

func Remove(path string) error {
	return os.RemoveAll(path)
}

func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		// защита от Zip Slip
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			if err := CreateDirs(fpath); err != nil {
				return fmt.Errorf("failed to create dir %s: %w", fpath, err)
			}
			continue
		}

		if err := CreateDirs(filepath.Dir(fpath)); err != nil {
			return fmt.Errorf("failed to create dir %s: %w", filepath.Dir(fpath), err)
		}

		dstFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", fpath, err)
		}

		rc, err := f.Open()
		if err != nil {
			dstFile.Close()
			return fmt.Errorf("failed to open file in zip %s: %w", f.Name, err)
		}

		if _, err := io.Copy(dstFile, rc); err != nil {
			rc.Close()
			dstFile.Close()
			return fmt.Errorf("failed to copy file %s: %w", fpath, err)
		}

		rc.Close()
		if err := dstFile.Close(); err != nil {
			return fmt.Errorf("failed to close file %s: %w", fpath, err)
		}
	}

	return nil
}
