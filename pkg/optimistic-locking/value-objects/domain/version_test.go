package domain_test

import (
    "testing"

    "architecture-bricks/pkg/optimistic-locking/value-objects/domain"

    "github.com/stretchr/testify/require"
)

func TestVersion(t *testing.T) {
    t.Parallel()

    t.Run("when_new_version_with_zero_then_success", func(t *testing.T) {
        t.Parallel()

        v, err := domain.NewVersion(0)

        require.NoError(t, err)
        require.Equal(t, 0, v.Value())
    })

    t.Run("when_new_version_with_positive_number_then_success", func(t *testing.T) {
        t.Parallel()

        v, err := domain.NewVersion(42)

        require.NoError(t, err)
        require.Equal(t, 42, v.Value())
    })

    t.Run("when_new_version_with_negative_number_then_error", func(t *testing.T) {
        t.Parallel()

        _, err := domain.NewVersion(-1)

        require.ErrorIs(t, err, domain.ErrNegativeVersion)
    })

    t.Run("when_next_then_version_incremented", func(t *testing.T) {
        t.Parallel()

        v, err := domain.NewVersion(5)
        require.NoError(t, err)

        next := v.Next()

        require.Equal(t, 6, next.Value())
    })
}
