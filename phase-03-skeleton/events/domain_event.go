package events

import "time"

type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

type baseEvent struct {
	occurredAt time.Time
}

func occurredAt(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now().UTC()
	}
	return t.UTC()
}

func (e baseEvent) OccurredAt() time.Time {
	return e.occurredAt
}
