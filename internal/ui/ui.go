package ui

import (
	"agent_client/internal/api"
	"agent_client/internal/license"
	"agent_client/internal/worker"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	Choices    []string
	Cursor     int
	Status     string
	Tab        int
	Client     *api.Client
	Worker        *worker.Worker
	Unlicensed    bool
	VersionStatus string
	MinVersion    string
	RedeemMode    bool
	AISettingsMode bool
	AIKeyInput    string
	SearchInput   string
	Searching     bool
	SearchResults []interface{}
	XP            int
	Level         int
	Points        int
	Rank          string
	Streak        int
	VolunteerMode bool
	CurrentTask   string
	LevelUpPending bool
	LicenseInput  string
	LicenseData   interface{}
	hwid          string
	baseURL       string
	DebugLog      []string
	LastTaskResult string
	DebugEnabled  bool
	
	// Update states
	UpdateAvailable bool
	UpdateVersion   string
	UpdateURL       string
	Updating        bool
	UpdateDone      bool
	UpdateTimer     int
	
	// Polling state
	NextPollTimer   int
}

func InitialModel(client *api.Client) Model {
	hwid, _ := license.GetEntropicHWID()
	api.CleanupOldVersion()
	return Model{
		Choices:       []string{"License", "Dashboard", "Search", "Settings", "Exit"},
		Status:        "Ready",
		Tab:           0,
		Client:        client,
		Worker:        &worker.Worker{Client: client},
		Unlicensed:    true,
		VersionStatus: "ok",
		RedeemMode:    false,
		XP:            0,
		Level:         1,
		Points:        0,
		Rank:          "Rookie Scouter",
		Streak:        0,
		VolunteerMode: false,
		CurrentTask:   "IDLE",
		hwid:          hwid,
		baseURL:       client.BaseURL,
		DebugLog:      []string{"Client initialized."},
		DebugEnabled:  false,
		UpdateTimer:   5,
		NextPollTimer: 30,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tick(), m.checkUpdate())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Ensure choices match debug mode
	if m.DebugEnabled {
		m.Choices = []string{"License", "Dashboard", "Search", "Settings", "Debug", "Exit"}
	} else {
		m.Choices = []string{"License", "Dashboard", "Search", "Settings", "Exit"}
	}

	switch msg := msg.(type) {
	case updateCheckMsg:
		if msg.err != nil {
			m.logDebug("Update check failed: %v", msg.err)
			// Proceed to stats if licensed
			if !m.Unlicensed {
				return m, m.fetchUserStats()
			}
			return m, nil
		}
		if msg.newVersion != "" {
			m.UpdateAvailable = true
			m.UpdateVersion = msg.newVersion
			m.UpdateURL = msg.downloadURL
			return m, nil
		}
		// No update, fetch stats if licensed
		if !m.Unlicensed {
			return m, m.fetchUserStats()
		}
		return m, nil

	case updateProgressMsg:
		if msg.err != nil {
			m.Status = "Update FAILED: " + msg.err.Error()
			m.Updating = false
			return m, nil
		}
		if msg.done {
			m.Updating = false
			m.UpdateDone = true
			m.Status = "UPDATE COMPLETE! APP WILL CLOSE. RESTART NOW."
			return m, tea.Quit
		}
		return m, nil

	case tickMsg:
		if m.UpdateAvailable && !m.Updating && !m.UpdateDone {
			m.UpdateTimer--
			if m.UpdateTimer <= 0 {
				m.Updating = true
				return m, m.doUpdate()
			}
		}
		
		if m.VolunteerMode && !m.Unlicensed && m.CurrentTask == "IDLE" {
			m.NextPollTimer--
			if m.NextPollTimer <= 0 {
				m.CurrentTask = "POLLING..."
				m.NextPollTimer = 30 // Reset for next cycle
				return m, m.pollTask()
			}
		} else {
			// If not in volunteer mode or busy, keep timer at 30
			if m.CurrentTask == "IDLE" {
				m.NextPollTimer = 30
			}
		}
		return m, tick()

	case taskClaimMsg:
		if msg.err != nil {
			m.Status = "Poll Error: " + msg.err.Error()
			m.logDebug("ReadyForTask ERROR: %v", msg.err)
			m.CurrentTask = "IDLE"
			m.NextPollTimer = 30
			return m, tick()
		}
		if msg.task == nil {
			m.CurrentTask = "IDLE"
			m.NextPollTimer = 30
			return m, tick()
		}
		
		taskLabel := msg.task.Type
		if msg.task.Payload["is_bounty"] == true {
			taskLabel = "🎯 BOUNTY: " + msg.task.Type
		}
		m.logDebug("Claimed Task %d (%s)", msg.task.ID, msg.task.Type)
		m.CurrentTask = fmt.Sprintf("WORKING: %s (%d)", taskLabel, msg.task.ID)
		return m, m.processTask(msg.task)

	case taskResultMsg:
		m.CurrentTask = "IDLE"
		m.NextPollTimer = 30
		m.LastTaskResult = msg.lastResult
		if msg.err != nil {
			m.Status = "Task Error: " + msg.err.Error()
			m.logDebug("SubmitTask ERROR: %v", msg.err)
		} else {
			m.Status = "Task Completed!"
			m.logDebug("Task Submitted Successfully")
			if msg.stats != nil {
				if lvl, ok := msg.stats["level"].(float64); ok {
					if int(lvl) > m.Level {
						m.LevelUpPending = true
						m.Status = fmt.Sprintf("LEVEL UP! Reached Level %d", int(lvl))
						m.logDebug("LEVEL UP to %d", int(lvl))
					}
				}
			}
			// Always sync stats after task
			return m, tea.Batch(tick(), m.fetchUserStats())
		}

	case searchResultMsg:
		if msg.err != nil {
			m.Status = "Search Error: " + msg.err.Error()
		} else {
			m.SearchResults = msg.items
			m.Status = fmt.Sprintf("Found %d items locally.", len(msg.items))
		}
		return m, nil

	case userStatsMsg:
		if msg.err != nil {
			m.logDebug("FetchStats ERROR: %v", msg.err)
		} else if msg.stats != nil {
			m.logDebug("Stats Synchronized")
			if xp, ok := msg.stats["xp"].(float64); ok {
				m.XP = int(xp)
			}
			if lvl, ok := msg.stats["level"].(float64); ok {
				m.Level = int(lvl)
			}
			if pts, ok := msg.stats["points"].(float64); ok {
				m.Points = int(pts)
			}
			if streak, ok := msg.stats["streak"].(float64); ok {
				m.Streak = int(streak)
			}
		}
		return m, nil

	case tea.KeyMsg:
		if m.UpdateAvailable && !m.Updating && !m.UpdateDone {
			switch msg.String() {
			case "u", "U":
				m.Updating = true
				return m, m.doUpdate()
			case "esc":
				m.UpdateAvailable = false // Hide banner
				return m, nil
			}
		}

		if m.UpdateDone {
			return m, tea.Quit // Force restart
		}

		m.LevelUpPending = false
		if m.RedeemMode {
			return m.handleLicenseUpdate(msg)
		}

		if m.AISettingsMode {
			return m.handleAISettingsUpdate(msg)
		}

		if m.Searching {
			switch msg.String() {
			case "esc":
				m.Searching = false
				return m, nil
			case "backspace":
				if len(m.SearchInput) > 0 {
					m.SearchInput = m.SearchInput[:len(m.SearchInput)-1]
				}
				return m, nil
			case "enter":
				m.Searching = false
				m.Status = "Searching locally..."
				return m, m.performLocalSearch(m.SearchInput)
			default:
				if len(msg.String()) == 1 {
					m.SearchInput += msg.String()
				}
				return m, nil
			}
		}

		return m.handleKeyMsg(msg)
	}
	return m, nil
}

func (m Model) View() string {
	if m.UpdateDone {
		return m.HeaderView() + "\n" + boxStyle.BorderForeground(successColor).Render(
			titleStyle.Render("--- UPDATE COMPLETE ---") + "\n\n" +
				"The new version has been installed successfully.\n" +
				"Please restart the application to apply changes.\n\n" +
				lipgloss.NewStyle().Foreground(dimColor).Render("Press any key to exit"),
		) + m.FooterView()
	}

	if m.Updating {
		return m.HeaderView() + "\n" + boxStyle.Render(
			titleStyle.Render("--- UPDATING ---") + "\n\n" +
				"Downloading " + m.UpdateVersion + "...\n" +
				"Please do not close the application.\n\n" +
				"Status: " + lipgloss.NewStyle().Foreground(primaryColor).Render("In Progress"),
		) + m.FooterView()
	}

	if m.VersionStatus == "deprecated" {
		return m.HeaderView() + fmt.Sprintf("\n\n  [!!!] VERSION DEPRECATED [!!!]\n  Minimum required version: %s\n  Please update your client.\n\n", m.MinVersion) + m.FooterView()
	}

	// 1. Header & Navigation
	s := m.HeaderView() + "\n"
	s += m.RenderTabs() + "\n"

	var content string

	if m.RedeemMode {
		input := lipgloss.NewStyle().Foreground(successColor).Render(m.LicenseInput + "_")
		content = boxStyle.Render(
			titleStyle.Render("--- REDEEM LICENSE ---") + "\n\n" +
				"Paste your license key here (Ctrl+V works):\n" +
				"> " + input + "\n\n" +
				lipgloss.NewStyle().Foreground(dimColor).Render("[Enter] Verify  [Esc] Cancel"),
		)
	} else if m.AISettingsMode {
		input := lipgloss.NewStyle().Foreground(successColor).Render(m.AIKeyInput + "_")
		content = boxStyle.Render(
			titleStyle.Render("--- CONFIGURE AI (GEMINI) ---") + "\n\n" +
				"Enter your Google Gemini API Key:\n" +
				"> " + input + "\n\n" +
				lipgloss.NewStyle().Foreground(dimColor).Render("[Enter] Save  [Esc] Cancel"),
		)
	} else {
		// Render Content based on Tab
		if m.Unlicensed && m.Tab != TabLicense {
			content = boxStyle.BorderForeground(errorColor).Render(
				lipgloss.NewStyle().Foreground(errorColor).Bold(true).Render("[!] ACCESS DENIED") + "\n\n" +
					"License REQUIRED to access this view.\n" +
					"Please go to the License tab to activate.",
			)
		} else {
			switch m.Tab {
			case TabLicense:
				content = m.RenderLicense()
			case TabDashboard:
				content = m.RenderDashboard()
			case TabSearch:
				content = m.RenderSearch()
			case TabSettings:
				content = m.RenderSettings()
			case TabDebug:
				content = m.RenderDebug()
			}
		}
	}

	s += content
	s += m.FooterView() + "\n"
	return s
}
