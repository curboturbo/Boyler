package files

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)


func Unzip(src string, dest string) error {
	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("failed to create dest directory: %w", err)
	}
	cmd := exec.Command("tar", "-xzf", src, "-C", dest)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("tar command failed: %w (stderr: %s)", err, stderr.String())
	}
	return nil
}