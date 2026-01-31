package worker

import "runtime"

type ResultMetadata struct {
	ExecutionTimeMS int64  `json:"execution_time_ms"`
	ScouterVersion  string `json:"scouter_version"`
	OS              string `json:"os"`
	Arch            string `json:"arch"`
	AIModel         string `json:"ai_model,omitempty"`
}

type WorkerResult struct {
	Status   string          `json:"status"`
	Data     interface{}     `json:"data"`
	Metadata ResultMetadata `json:"metadata"`
}

func NewWorkerResult(status string, data interface{}, execTimeMS int64, version string, aiModel string) *WorkerResult {
	return &WorkerResult{
		Status: status,
		Data:   data,
		Metadata: ResultMetadata{
			ExecutionTimeMS: execTimeMS,
			ScouterVersion:  version,
			OS:              runtime.GOOS,
			Arch:            runtime.GOARCH,
			AIModel:         aiModel,
		},
	}
}
