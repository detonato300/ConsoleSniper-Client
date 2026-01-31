package ui

import (
	"agent_client/internal/api"
	"agent_client/internal/license"
	"agent_client/internal/security"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) handleLicenseUpdate(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.RedeemMode = false
		return m, nil
	case "enter":
		m.LicenseInput = strings.TrimSpace(m.LicenseInput)
		obfKey := []byte{0x7b, 0x7b, 0x26, 0x77, 0x70, 0x75, 0x26, 0x73, 0x27, 0x7a, 0x26, 0x72, 0x73, 0x26, 0x7a, 0x71, 0x72, 0x75, 0x73, 0x76, 0x74, 0x74, 0x21, 0x75, 0x27, 0x26, 0x75, 0x71, 0x7b, 0x7a, 0x24, 0x7b, 0x27, 0x21, 0x20, 0x73, 0x73, 0x72, 0x7a, 0x70, 0x71, 0x26, 0x72, 0x7b, 0x72, 0x76, 0x27, 0x26, 0x75, 0x24, 0x7b, 0x74, 0x27, 0x74, 0x77, 0x27, 0x75, 0x71, 0x71, 0x27, 0x7a, 0x20, 0x74, 0x70}
		pubKey := security.Deobfuscate(obfKey, 0x42)

		payload, err := license.VerifyLicense(m.LicenseInput, pubKey)
		if err != nil {
			m.Status = "Verification FAILED: " + err.Error()
		} else {
			m.Status = "Verification SUCCESS! Saving..."
			m.Unlicensed = false
			m.LicenseData = payload
			m.Client = api.NewClient(m.baseURL, m.LicenseInput, m.hwid)
			if m.Worker != nil {
				m.Worker.Client = m.Client
			}
			hwid, _ := license.GetEntropicHWID()
			_ = license.SaveLicenseEncrypted("license.key", m.LicenseInput, hwid)
			m.RedeemMode = false
			return m, m.fetchUserStats()
		}
		m.RedeemMode = false
		return m, nil
	case "backspace":
		if len(m.LicenseInput) > 0 {
			m.LicenseInput = m.LicenseInput[:len(m.LicenseInput)-1]
		}
		return m, nil
	case "ctrl+v":
		str, err := clipboard.ReadAll()
		if err == nil {
			m.LicenseInput = strings.TrimSpace(str)
		}
		return m, nil
	default:
		if len(msg.String()) == 1 {
			m.LicenseInput += msg.String()
		}
		return m, nil
	}
}

func (m Model) handleAISettingsUpdate(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.AISettingsMode = false
		return m, nil
	case "enter":
		m.AIKeyInput = strings.TrimSpace(m.AIKeyInput)
		m.Status = "Validating key..."
		err := api.ValidateGeminiKey(m.AIKeyInput)
		if err != nil {
			m.Status = "Validation FAILED: " + err.Error()
			return m, nil
		}
		_ = api.SaveAIConfig(m.AIKeyInput)
		m.Status = "AI Key validated and saved."
		m.AISettingsMode = false
		return m, nil
	case "backspace":
		if len(m.AIKeyInput) > 0 {
			m.AIKeyInput = m.AIKeyInput[:len(m.AIKeyInput)-1]
		}
		return m, nil
	case "ctrl+v":
		str, err := clipboard.ReadAll()
		if err == nil {
			m.AIKeyInput = strings.TrimSpace(str)
		}
		return m, nil
	default:
		if len(msg.String()) == 1 {
			m.AIKeyInput += msg.String()
		}
		return m, nil
	}
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "q":
		if !m.Searching && !m.RedeemMode && !m.AISettingsMode {
			return m, tea.Quit
		}
	case "1":
		return m.SwitchTab(TabLicense), nil
	case "2":
		return m.SwitchTab(TabDashboard), nil
	case "3":
		return m.SwitchTab(TabSearch), nil
	case "4":
		return m.SwitchTab(TabSettings), nil
	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
		}
	case "down", "j":
		if m.Cursor < len(m.Choices)-1 {
			m.Cursor++
		}
	case "left", "h":
		newTab := m.Tab - 1
		if newTab < 0 {
			newTab = len(m.Choices) - 2
		}
		return m.SwitchTab(newTab), nil
	case "right", "l":
		newTab := m.Tab + 1
		if newTab >= len(m.Choices)-1 {
			newTab = 0
		}
		return m.SwitchTab(newTab), nil
	case "v", "V":
		if m.Tab == TabDashboard {
			m.VolunteerMode = !m.VolunteerMode
			if m.VolunteerMode {
				m.Status = "Volunteer Mode: ON"
			} else {
				m.Status = "Volunteer Mode: OFF"
				m.CurrentTask = "IDLE"
			}
		}
		return m, nil
	case "enter", " ":
		if m.Choices[m.Cursor] == "Exit" {
			return m, tea.Quit
		}
		if m.Choices[m.Cursor] == "License" && m.Tab == TabLicense {
			m.RedeemMode = true
			m.LicenseInput = ""
			str, err := clipboard.ReadAll()
			if err == nil && strings.Contains(str, ".") {
				m.LicenseInput = strings.TrimSpace(str)
				m.Status = "Key detected in clipboard!"
			}
			return m, nil
		}
		if m.Choices[m.Cursor] == "Settings" && m.Tab == TabSettings {
			m.AISettingsMode = true
			m.AIKeyInput = ""
			cfg, err := api.LoadAIConfig()
			if err == nil {
				m.AIKeyInput = cfg.GeminiAPIKey
				m.Status = "Existing AI Key loaded."
			}
			return m, nil
		}
		return m.SwitchTab(m.Cursor), nil
	case "r", "R":
		if m.Tab == TabDashboard {
			m.Status = "Redeeming 500 pts for 24h Premium..."
			return m, m.redeemPoints("premium_24h")
		}
	case "s", "S":
		if m.Tab == TabSearch {
			tier := "free"
			if ld, ok := m.LicenseData.(*license.LicensePayload); ok {
				tier = ld.Tier
			}
			if tier != "free" {
				m.Searching = true
				m.SearchInput = ""
				return m, nil
			}
		}
	}
	return m, nil
}
