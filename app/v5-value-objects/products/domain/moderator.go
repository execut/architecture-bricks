package domain

import "github.com/google/uuid"

// Moderator — бизнес-объект, идентифицирующий модератора.
type Moderator struct {
	value string
}

func NewModerator(id string) (Moderator, error) {
	if _, err := uuid.Parse(id); err != nil {
		return Moderator{}, ErrInvalidUserID
	}

	return Moderator{value: id}, nil
}

func (m Moderator) Value() string {
	return m.value
}
