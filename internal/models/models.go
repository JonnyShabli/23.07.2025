package models

import (
	"github.com/google/uuid"
)

const (
	StatusIdle       = "Idle"
	StatusProcessing = "Processing"
	StatusDone       = "Done"
)

type Task struct {
	TaskId     uuid.UUID `json:"task_id"`
	Links      []string  `json:"links"`
	LinksDone  []string  `json:"links_done"`
	LinksError []string  `json:"links_error"`
	Count      int       `json:"count"`
	Status     string    `json:"status"`
	Zip        *[]byte   `json:"zip"`
}

type Status struct {
	TaskId     uuid.UUID `json:"task_id"`
	Status     string    `json:"status"`
	LinksDone  []string  `json:"links_done"`
	LinksError []string  `json:"links_error"`
	Zip        *[]byte   `json:"zip"`
}

type AddLinksRequest struct {
	TaskId uuid.UUID `json:"task_id"`
	Links  []string  `json:"links"`
}
