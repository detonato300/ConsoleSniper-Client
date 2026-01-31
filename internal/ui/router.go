package ui

const (
	TabLicense = iota
	TabDashboard
	TabSearch
	TabSettings
	TabDebug
)

func (m Model) SwitchTab(tab int) Model {
	if m.Unlicensed && tab != TabLicense {
		return m
	}
	m.Tab = tab
	m.Status = "Viewing: " + m.Choices[tab]
	return m
}
