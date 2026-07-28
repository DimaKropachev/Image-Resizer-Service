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
	OutPath string
	Status  string
	Err     error
}
