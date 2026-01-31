package ui

import (
	"agent_client/internal/api"
	"fmt"
	"runtime"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) logDebug(format string, a ...interface{}) {
	if !m.DebugEnabled {
		return
	}
	msg := fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), fmt.Sprintf(format, a...))
	m.DebugLog = append(m.DebugLog, msg)
	if len(m.DebugLog) > 50 {
		m.DebugLog = m.DebugLog[1:]
	}
}

func (m Model) fetchUserStats() tea.Cmd {
	return func() tea.Msg {
		stats, err := m.Client.GetStats()
		return userStatsMsg{stats: stats, err: err}
	}
}

func (m Model) checkUpdate() tea.Cmd {
	return func() tea.Msg {
		release, err := api.GetLatestRelease()
		if err != nil {
			return updateCheckMsg{err: err}
		}

		current := "v3.4.3"
		if release.TagName != current {
			target := "consolesniper_linux_amd64"
			if runtime.GOOS == "windows" {
				target = "consolesniper.exe"
			} else if runtime.GOOS == "darwin" {
				if runtime.GOARCH == "arm64" {
					target = "consolesniper_darwin_arm64"
				} else {
					target = "consolesniper_darwin_amd64"
				}
			}

			var downloadURL string
			for _, asset := range release.Assets {
				if asset.Name == target {
					downloadURL = asset.BrowserDownloadURL
					break
				}
			}

			if downloadURL != "" {
				return updateCheckMsg{newVersion: release.TagName, downloadURL: downloadURL}
			}
		}
		return updateCheckMsg{}
	}
}

func (m Model) doUpdate() tea.Cmd {
	return func() tea.Msg {
		err := api.DownloadAndReplace(m.UpdateURL)
		if err != nil {
			return updateProgressMsg{err: err}
		}
		return updateProgressMsg{done: true}
	}
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) redeemPoints(rewardType string) tea.Cmd {
	return func() tea.Msg {
		err := m.Client.RedeemPoints(rewardType)
		if err != nil {
			return taskResultMsg{err: err}
		}
		return taskResultMsg{success: true}
	}
}
