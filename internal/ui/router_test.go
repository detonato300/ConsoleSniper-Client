package ui

import (
	"agent_client/internal/api"
	"strings"
	"testing"
)

func TestNavigation(t *testing.T) {
	client := api.NewClient("http://localhost", "token", "hwid")
	m := InitialModel(client)
	m.Unlicensed = true // Start locked

	// Try to switch to Dashboard while locked
	m = m.SwitchTab(1)
	if m.Tab != 0 {
		t.Errorf("Expected Tab 0 (License) while locked, got %d", m.Tab)
	}

	// Unlock
	m.Unlicensed = false
	
	// Switch to Search
	m = m.SwitchTab(2)
	if m.Tab != 2 {
		t.Errorf("Expected Tab 2 (Search) after unlocking, got %d", m.Tab)
	}
	
	// Switch back to Dashboard
	m = m.SwitchTab(1)
	if m.Tab != 1 {
		t.Errorf("Expected Tab 1 (Dashboard), got %d", m.Tab)
	}
}

func TestKillSwitchRendering(t *testing.T) {
	client := api.NewClient("http://localhost", "token", "hwid")
	m := InitialModel(client)
	m.VersionStatus = "deprecated"
	m.MinVersion = "2.1.0"
	
	view := m.View()
	if !strings.Contains(view, "VERSION DEPRECATED") {
		t.Errorf("View should show version deprecation message, got: %s", view)
	}
	if !strings.Contains(view, "2.1.0") {
		t.Errorf("View should show minimum version, got: %s", view)
	}
}
