package ui

import (
	"agent_client/internal/license"
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) RenderSearch() string {
	tier := "free"
	if ld, ok := m.LicenseData.(*license.LicensePayload); ok {
		tier = ld.Tier
	}

	if tier == "free" {
		return boxStyle.BorderForeground(lipgloss.Color("#EED49F")).Render(
			lipgloss.NewStyle().Foreground(lipgloss.Color("#EED49F")).Bold(true).Render("[!] PERSONAL SEARCH LOCKED") + "\n\n" +
				"This feature is available for Premium tiers.\n" +
				"Earn points via Volunteer Mode to unlock!",
		)
	}

	searchUI := "Search for consoles locally (Results NOT shared):\n\n" +
		lipgloss.NewStyle().Foreground(primaryColor).Render("[Press 'S' to start local search]")

	if m.Searching {
		searchUI = "Enter keyword: " + lipgloss.NewStyle().Foreground(successColor).Render(m.SearchInput+"_") + "\n\n" +
			lipgloss.NewStyle().Foreground(dimColor).Render("[Enter] Search  [Esc] Cancel")
	}

	results := ""
	if len(m.SearchResults) > 0 {
		results = "\n\nRecent Results:\n"
		for i, item := range m.SearchResults {
			if i > 4 {
				break
			}
			results += fmt.Sprintf(" ● %v\n", item)
		}
	}
	return boxStyle.Render(titleStyle.Render("---" + " PRIVATE SEARCH " + "---") + "\n" + searchUI + results)
}
