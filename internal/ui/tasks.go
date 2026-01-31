package ui

import (
	"agent_client/internal/api"
	"context"
	"encoding/json"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) pollTask() tea.Cmd {
	return func() tea.Msg {
		caps := []string{"scraping", "browser"}
		aiStatus := "none"
		quota := 0

		cfg, err := api.LoadAIConfig()
		if err == nil && cfg.GeminiAPIKey != "" {
			caps = append(caps, "ai_gemini")
			aiStatus = "active"
			if api.IsLowQuota(100) {
				aiStatus = "low_quota"
			}
			quota = api.GetRemainingQuota()
		}

		task, err := m.Client.ReadyForTask(context.Background(), caps, aiStatus, quota, "3.4.6")
		return taskClaimMsg{task: task, err: err}
	}
}

func (m Model) processTask(task *api.Task) tea.Cmd {
	return func() tea.Msg {
		res, err := m.Worker.ProcessTask(task, nil)
		if err != nil {
			return taskResultMsg{err: err}
		}

		resJSON, _ := json.MarshalIndent(res.Data, "", "  ")
		resPreview := string(resJSON)
		if len(resPreview) > 1000 {
			resPreview = resPreview[:1000] + "\n... (truncated)"
		}

		stats, err := m.Client.SubmitTask(context.Background(), task.ID, res.Data, "completed", res.Metadata)
		if err != nil {
			return taskResultMsg{err: err, lastResult: resPreview}
		}

		return taskResultMsg{success: true, stats: stats, lastResult: resPreview}
	}
}

func (m Model) performLocalSearch(query string) tea.Cmd {
	return func() tea.Msg {
		task := &api.Task{
			Type: "search",
			Payload: map[string]interface{}{
				"query": query,
			},
		}
		res, err := m.Worker.ProcessTask(task, nil)
		if err != nil {
			return searchResultMsg{err: err}
		}
		
		itemsData := res.Data.(map[string]interface{})["items"].([]interface{})
		return searchResultMsg{items: itemsData}
	}
}
