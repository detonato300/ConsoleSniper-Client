package ui

import (
	"agent_client/internal/api"
	"regexp"
	"strings"
	"testing"
)

func stripANSI(str string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	return re.ReplaceAllString(str, "")
}

func TestHeaderFooter(t *testing.T) {
	client := api.NewClient("http://localhost", "token", "hwid")
	m := InitialModel(client)
	m.Status = "Test Status"
	
	header := stripANSI(m.HeaderView())
	if !strings.Contains(strings.ToUpper(header), "CONSOLE SNIPER") {
		t.Errorf("Header missing app title, got: %s", header)
	}

	footer := stripANSI(m.FooterView())
	if !strings.Contains(footer, "Test Status") {
		t.Errorf("Footer missing status, got: %s", footer)
	}
}
