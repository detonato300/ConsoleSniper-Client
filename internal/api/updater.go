package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
)

// DownloadAndReplace downloads the new binary and replaces the current executable.
func DownloadAndReplace(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download: %d", resp.StatusCode)
	}

	executable, err := os.Executable()
	if err != nil {
		return err
	}

	// On Windows, we can't overwrite a running file.
	// We rename the current file to .old and save the new one.
	// RENAMING is allowed even if the file is open.
	oldPath := executable + ".old"
	if runtime.GOOS == "windows" {
		_ = os.Remove(oldPath)
		if err := os.Rename(executable, oldPath); err != nil {
			return fmt.Errorf("failed to move current binary: %w", err)
		}
	}

	// Create the new binary in the original path
	out, err := os.OpenFile(executable, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		// Rollback on Windows if possible
		if runtime.GOOS == "windows" {
			_ = os.Rename(oldPath, executable)
		}
		return fmt.Errorf("failed to create new binary: %w", err)
	}

	_, err = io.Copy(out, resp.Body)
	out.Close() // Explicitly close before Chmod or finishing

	if err != nil {
		return fmt.Errorf("failed to save download: %w", err)
	}

	// For Unix systems, set execution bits
	if runtime.GOOS != "windows" {
		err = os.Chmod(executable, 0755)
		if err != nil {
			return err
		}
	}

	return nil
}

// CleanupOldVersion removes the .old file if it exists (Windows only).
func CleanupOldVersion() {
	if runtime.GOOS == "windows" {
		executable, err := os.Executable()
		if err == nil {
			_ = os.Remove(executable + ".old")
		}
	}
}
