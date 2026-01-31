package ui

import (
	"agent_client/internal/api"
	"time"
)

type tickMsg time.Time

type taskClaimMsg struct {
	task *api.Task
	err  error
}

type taskResultMsg struct {
	success    bool
	stats      map[string]interface{}
	lastResult string
	err        error
}

type userStatsMsg struct {
	stats map[string]interface{}
	err   error
}

type updateCheckMsg struct {
	newVersion string
	downloadURL string
	err        error
}

type updateProgressMsg struct {
	status string
	done   bool
	err    error
}

type searchResultMsg struct {
	items []interface{}
	err   error
}
