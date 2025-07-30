package models

const (
	StatusIdle       = "Idle"
	StatusProcessing = "Processing"
	StatusDone       = "Done"
)

type Task struct {
	TaskId        string            `json:"task_id"`
	Links         []string          `json:"links,omitempty"`
	LinksStatuses map[string]string `json:"links_statuses,omitempty"`
	LinksError    map[string]string `json:"links_error,omitempty"`
	Status        string            `json:"status"`
	ZipPath       string            `json:"zip_path,omitempty"`
}

type Status struct {
	TaskId        string            `json:"task_id"`
	Status        string            `json:"status"`
	LinksStatuses map[string]string `json:"links_statuses,omitempty"`
	LinksError    map[string]string `json:"links_error,omitempty"`
	ZipPath       string            `json:"url,omitempty"`
}

type AddLinksRequest struct {
	TaskId string   `json:"task_id"`
	Links  []string `json:"links"`
}

type DownloadJob struct {
	TaskId string `json:"task_id"`
	Url    string `json:"url"`
	Err    error  `json:"error,omitempty"`
}

type ZipJob struct {
	TaskId         string  `json:"task_id"`
	Url            string  `json:"url"`
	Data           *[]byte `json:"data"`
	ResponseStatus string  `json:"response_status"`
	FileName       string  `json:"file_name"`
	Err            error   `json:"error"`
}

type ValueAndError struct {
	Value interface{}
	Err   error
}
