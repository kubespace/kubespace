package sse

const (
	EventTypePipeline    string = "pipeline"
	EventTypePipelineRun string = "pipeline_run"
)

const EventLabelType = "__event_type"

const (
	CatalogDatabase = "database"
	CatalogCluster  = "cluster"
)

type Event struct {
	Labels map[string]string `json:"labels"`
	Object interface{}       `json:"object"`
}
