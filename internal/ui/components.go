package ui

import (
	"agent_client/internal/api"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

var (
	// Colors
	primaryColor = lipgloss.Color("#7D56F4")
	accentColor  = lipgloss.Color("#F4DBD6")
	successColor = lipgloss.Color("#A6DA95")
	errorColor   = lipgloss.Color("#ED8796")
	dimColor     = lipgloss.Color("#5B6078")

	// Styles
	headerStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Background(primaryColor).
			Padding(0, 1).
			Bold(true)

	tabStyle = lipgloss.NewStyle().
			Foreground(dimColor).
			Padding(0, 2).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(dimColor)

	activeTabStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Padding(0, 2).
			Border(lipgloss.ThickBorder(), false, false, true, false).
			BorderForeground(primaryColor)

	statusStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Italic(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			MarginBottom(1)
)

func (m Model) HeaderView() string {
	aiStatus := "AI: OFF"
	aiColor := errorColor
	cfg, err := api.LoadAIConfig()
	if err == nil && cfg.GeminiAPIKey != "" {
		rem := api.GetRemainingQuota()
		aiStatus = fmt.Sprintf("AI: ON (%d)", rem)
		aiColor = successColor
		if api.IsLowQuota(100) {
			aiStatus = fmt.Sprintf("AI: LOW (%d)", rem)
			aiColor = lipgloss.Color("#EED49F")
		}
	}

	title := headerStyle.Render(" CONSOLE SNIPER ")
	nodeStatus := lipgloss.NewStyle().Foreground(aiColor).Render(" ● " + aiStatus)
	
	version := lipgloss.NewStyle().Foreground(dimColor).Render("v3.4.3")
	
	debugTag := ""
	if m.DebugEnabled {
		debugTag = " " + lipgloss.NewStyle().
			Foreground(accentColor).
			Background(errorColor).
			Padding(0, 1).
			Bold(true).
			Render("DEBUG")
	}
	
	return lipgloss.JoinHorizontal(lipgloss.Center, title, debugTag, " ", nodeStatus, " ", version) + "\n"
}

func (m Model) RenderTabs() string {
	var tabs []string
	for i, choice := range m.Choices {
		if choice == "Exit" {
			continue
		}
		style := tabStyle
		if m.Tab == i {
			style = activeTabStyle
		}
		tabs = append(tabs, style.Render(choice))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...) + "\n"
}

func (m Model) FooterView() string {
	footer := "\n" + lipgloss.NewStyle().Foreground(dimColor).Render("----------------------------------------") + "\n"
	
	if m.UpdateAvailable && !m.Updating && !m.UpdateDone {
		updateBanner := lipgloss.NewStyle().
			Foreground(accentColor).
			Background(lipgloss.Color("#EED49F")).
			Padding(0, 1).
			Bold(true).
			Render(fmt.Sprintf(" UPDATE AVAILABLE: %s (Auto-update in %ds) [Press 'U' to start / 'Esc' to hide] ", m.UpdateVersion, m.UpdateTimer))
		footer += updateBanner + "\n"
	}

	status := statusStyle.Render(m.Status)
	if strings.Contains(m.Status, "FAILED") || strings.Contains(m.Status, "Error") {
		status = lipgloss.NewStyle().Foreground(errorColor).Bold(true).Render(m.Status)
	}
	
	return footer + lipgloss.JoinHorizontal(lipgloss.Top, "Status: ", status)
}