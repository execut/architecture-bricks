package domain

import "github.com/google/uuid"

// User — бизнес-объект, идентифицирующий пользователя.
type User struct {
	value string
}

func NewUser(id string) (User, error) {
	if _, err := uuid.Parse(id); err != nil {
		return User{}, ErrInvalidUserID
	}

	return User{value: id}, nil
}

func (u User) Value() string {
	return u.value
}
