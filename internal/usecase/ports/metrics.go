package ports

import "time"

type Metrics interface {
	IncRequest()
	ObserveRequestDuration(duration time.Duration)
	IncPVZCreated()
	IncReceptionCreated()
	IncProductAdded()
}
