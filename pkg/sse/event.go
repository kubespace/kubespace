package sse

type EventType string

const EventTypePipeline EventType = "pipeline"

type Event struct {
	Type   EventType
	Key    string
	Object interface{}
}
