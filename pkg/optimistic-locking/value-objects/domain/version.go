package domain

import "errors"

var ErrNegativeVersion = errors.New("version must be non-negative")

// Version — VO версии события. Инкапсулирует неотрицательное целое число.
type Version struct {
    value int
}

func NewInitialVersion() Version {
    return Version{value: 0}
}

func NewVersion(v int) (Version, error) {
    if v < 0 {
        return Version{}, ErrNegativeVersion
    }

    return Version{value: v}, nil
}

func (v Version) Value() int {
    return v.value
}

func (v Version) Next() Version {
    return Version{value: v.value + 1}
}
