package ui

import (
	"agent_client/internal/api"
	"testing"
)

func TestInitialModel(t *testing.T) {
	client := api.NewClient("http://localhost", "token", "test-hwid")
	m := InitialModel(client)
	
	expectedChoices := []string{"License", "Dashboard", "Search", "Settings", "Exit"}
	if len(m.Choices) != len(expectedChoices) {
		t.Errorf("Expected %d choices, got %d", len(expectedChoices), len(m.Choices))
	}

	for i, choice := range m.Choices {
		if choice != expectedChoices[i] {
			t.Errorf("Expected choice %d to be %s, got %s", i, expectedChoices[i], choice)
		}
	}

	if m.Status != "Ready" {
		t.Errorf("Expected status to be 'Ready', got %s", m.Status)
	}
}
