package ui

import (
	"agent_client/internal/license"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) RenderDashboard() string {
	tier := "FREE"
	if ld, ok := m.LicenseData.(*license.LicensePayload); ok {
		tier = strings.ToUpper(ld.Tier)
	}

	stats := fmt.Sprintf("Rank:   %s\nLevel:  %d\nXP:     %d\nPoints: %d\nTier:   %s",
		lipgloss.NewStyle().Foreground(lipgloss.Color("#FAB387")).Render(m.Rank),
		m.Level, m.XP, m.Points,
		lipgloss.NewStyle().Foreground(primaryColor).Bold(true).Render(tier),
	)

	volunteerStatus := lipgloss.NewStyle().Foreground(errorColor).Render("OFF")
	if m.VolunteerMode {
		volunteerStatus = lipgloss.NewStyle().Foreground(successColor).Bold(true).Render("ON")
	}

	taskColor := dimColor
	if strings.Contains(m.CurrentTask, "WORKING") {
		taskColor = successColor
	}
	if strings.Contains(m.CurrentTask, "BOUNTY") {
		taskColor = lipgloss.Color("#EED49F")
	}

	taskInfo := lipgloss.NewStyle().Foreground(taskColor).Render(m.CurrentTask)
	if m.CurrentTask == "IDLE" && m.VolunteerMode {
		taskInfo += lipgloss.NewStyle().Foreground(dimColor).Render(fmt.Sprintf(" (Next poll in: %ds)", m.NextPollTimer))
	}

	dashboardContent := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(30).Render(stats),
		lipgloss.NewStyle().PaddingLeft(4).Render(
			"Volunteer Mode: ["+volunteerStatus+"]\n"+
			"(Press 'v' to toggle)\n\n"+
			"Current Task:\n"+taskInfo),
	)

	if m.LevelUpPending {
		dashboardContent += lipgloss.NewStyle().Foreground(successColor).Bold(true).PaddingTop(1).Render("\n🎉 LEVEL UP! YOU ARE GETTING STRONGER! 🎉")
	}

	return boxStyle.Render(titleStyle.Render("---"+" OPERATIONAL DASHBOARD "+"---") + "\n" + dashboardContent)
}
