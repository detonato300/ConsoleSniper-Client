package main

import (
	"os"
	"testing"
)

func TestProjectStructure(t *testing.T) {
	dirs := []string{
		"cmd/client",
		"internal/ui",
		"internal/scraper",
		"internal/license",
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Directory %s does not exist", dir)
		}
	}

	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		t.Error("go.mod file does not exist")
	}
}
