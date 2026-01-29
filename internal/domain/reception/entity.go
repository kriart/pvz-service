package reception

import "time"

const (
	StatusInProgress = "in_progress"
	StatusClosed     = "closed"
)

type Reception struct {
	ID        string
	StartedAt time.Time
	PVZID     string
	Status    string
}
