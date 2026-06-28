package domain

type Event struct {
    version int64
    payload interface{}
}

func NewEvent(version int64, payload interface{}) Event {
    return Event{version: version, payload: payload}
}

func (e Event) Version() int64 {
    return e.version
}

func (e Event) Payload() interface{} {
    return e.payload
}
