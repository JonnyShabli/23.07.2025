package models

import (
	"github.com/google/uuid"
)

const (
	StatusIdle       = "Idle"
	StatusProcessing = "Processing"
	StatudDone       = "Done"
)

type Task struct {
	TaskId uuid.UUID `json:"task_id"`
	Links  []string  `json:"links"`
	Count  int       `json:"count"`
	Status string    `json:"status"`
	Zip    []byte    `json:"zip"`
}
