package task_models

type TaskStatus string

const (
	StatusNew        TaskStatus = "New"
	StatusInProgress TaskStatus = "In Progress"
	StatusCompleted  TaskStatus = "Done"
)

func (s TaskStatus) IsValid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusCompleted:
		return true
	default:
		return false
	}
}

type Task struct {
	ID         string         `json:"id,omitempty" validate:"required"`
	UserID     string         `json:"user_uid,omitempty" validate:"required"`
	Attributes TaskAttributes `json:"attributes,omitempty" validate:"required"`
	Deleted    bool           `json:"-"`
}

type TaskAttributes struct {
	Status      TaskStatus `json:"status" validate:"required"`
	Title       string     `json:"title" validate:"required,min=1"`
	Description string     `json:"description" validate:"required,min=1"`
}
