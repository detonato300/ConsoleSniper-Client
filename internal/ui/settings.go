package ui

import (
	"github.com/charmbracelet/lipgloss"
)

func (m Model) RenderSettings() string {
	return boxStyle.Render(
		titleStyle.Render("-- SETTINGS --") + "\n" +
			"App Version: v3.4.6\n" +
			"Engine:      C2-Orchestrator v3.4.6\n\n" +
			lipgloss.NewStyle().Foreground(primaryColor).Render("[Press 'Enter'] to configure Gemini API Key"),
	)
}

func (m Model) RenderDebug() string {
	logContent := ""
	start := 0
	if len(m.DebugLog) > 8 {
		start = len(m.DebugLog) - 8
	}
	for _, log := range m.DebugLog[start:] {
		logContent += " " + log + "\n"
	}

	resultPreview := m.LastTaskResult
	if len(resultPreview) > 300 {
		resultPreview = resultPreview[:300] + "..."
	}

	return boxStyle.Render(
		titleStyle.Render("-- DEBUG CONSOLE --") + "\n" +
			lipgloss.NewStyle().Underline(true).Render("Last Task Data:") + "\n" +
			lipgloss.NewStyle().Foreground(dimColor).Render(resultPreview) + "\n\n" +
			lipgloss.NewStyle().Underline(true).Render("System Logs:") + "\n" +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#CAD3F5")).Render(logContent),
	)
}
