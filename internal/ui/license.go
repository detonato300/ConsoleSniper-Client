package ui

import (
	"agent_client/internal/license"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) RenderLicense() string {
	status := lipgloss.NewStyle().Foreground(errorColor).Render("UNLICENSED (Locked)")
	details := "Discord: Not linked\n\n[Press 'Enter' to start activation]"

	if !m.Unlicensed {
		if ld, ok := m.LicenseData.(*license.LicensePayload); ok {
			status = lipgloss.NewStyle().Foreground(successColor).Bold(true).Render("ACTIVATED")
			details = fmt.Sprintf("User ID: %d\nTier:    %s\nExpires: %s\nHWID:    Locked", ld.UserID, strings.ToUpper(ld.Tier), ld.Expires)
		}
		details += "\n\n[Press 'Enter' to change/re-activate]"
	}

	return boxStyle.Render(
		titleStyle.Render("--- LICENSE & ACTIVATION ---") + "\n" +
			"Status: " + status + "\n\n" + details,
	)
}
