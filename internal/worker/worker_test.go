package worker

import (
	"agent_client/internal/api"
	"testing"
)

type MockScraper struct {
	Called bool
}

func (m *MockScraper) Search(query string, params map[string]string) ([]interface{}, error) {
	m.Called = true
	return []interface{}{"item1"}, nil
}

func TestWorker_ProcessTask(t *testing.T) {
	w := &Worker{}
	scraper := &MockScraper{}
	
	task := &api.Task{
		ID:   1,
		Type: "mercari_search",
		Payload: map[string]interface{}{
			"keyword": "3ds",
		},
	}
	
	result, err := w.ProcessTask(task, scraper)
	if err != nil {
		t.Fatalf("ProcessTask failed: %v", err)
	}
	
	if !scraper.Called {
		t.Error("Scraper was not called")
	}
	
	items := result.Data.(map[string]interface{})["items"].([]interface{})
	if len(items) != 1 {
		t.Errorf("Expected 1 result, got %v", len(items))
	}
}
