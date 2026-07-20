package models

const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusDone       = "done"
	StatusFailed     = "failed"
)

type Task struct {
	ID      string
	ImgPath string
	Status  string
	Err     error
}
