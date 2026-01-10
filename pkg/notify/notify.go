package notify

import (
	"fmt"

	"github.com/gen2brain/beeep"
)

const appName = "Sluhach"

func Notify(title, s string, icon ...string) error {
	if len(icon) == 0 {
		icon = append(icon, "media-record-symbolic")
	}
	beeep.AppName = appName
	if err := beeep.Notify(title, s, icon[0]); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}
	return nil
}
