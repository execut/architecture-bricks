package domain_test

import (
	"testing"

	"architecture-bricks/app/v5-value-objects/products/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestProductID(t *testing.T) {
	t.Parallel()

	t.Run("when_new_product_id_with_valid_uuid_then_success", func(t *testing.T) {
		t.Parallel()

		id, err := domain.NewProductID(uuid.NewString())

		require.NoError(t, err)
		require.NotEmpty(t, id.Value())
	})

	t.Run("when_new_product_id_with_invalid_uuid_then_error", func(t *testing.T) {
		t.Parallel()

		_, err := domain.NewProductID("not-a-uuid")

		require.ErrorIs(t, err, domain.ErrInvalidProductID)
	})

	t.Run("when_new_product_id_with_empty_string_then_error", func(t *testing.T) {
		t.Parallel()

		_, err := domain.NewProductID("")

		require.ErrorIs(t, err, domain.ErrInvalidProductID)
	})
}
