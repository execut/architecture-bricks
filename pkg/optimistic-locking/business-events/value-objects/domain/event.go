package domain

import "architecture-bricks/pkg/optimistic-locking/value-objects/domain"

type Event struct {
    version domain.Version
    payload interface{}
}

func NewEvent(version domain.Version, payload interface{}) Event {
    return Event{version: version, payload: payload}
}

func (e Event) Version() domain.Version {
    return e.version
}

func (e Event) Payload() interface{} {
    return e.payload
}
