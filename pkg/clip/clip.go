package clip

import (
	"fmt"

	"github.com/atotto/clipboard"
)

func Clip(s string) error {
	if err := clipboard.WriteAll(s); err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}
	return nil
}
