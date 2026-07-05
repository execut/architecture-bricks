package domain_test

import (
	"testing"

	"architecture-bricks/app/v5-value-objects/products/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUser(t *testing.T) {
	t.Parallel()

	t.Run("when_new_user_with_valid_uuid_then_success", func(t *testing.T) {
		t.Parallel()

		user, err := domain.NewUser(uuid.NewString())

		require.NoError(t, err)
		require.NotEmpty(t, user.Value())
	})

	t.Run("when_new_user_with_invalid_uuid_then_error", func(t *testing.T) {
		t.Parallel()

		_, err := domain.NewUser("not-a-uuid")

		require.ErrorIs(t, err, domain.ErrInvalidUserID)
	})

	t.Run("when_new_user_with_empty_string_then_error", func(t *testing.T) {
		t.Parallel()

		_, err := domain.NewUser("")

		require.ErrorIs(t, err, domain.ErrInvalidUserID)
	})
}
