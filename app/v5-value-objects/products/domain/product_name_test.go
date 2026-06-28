package domain_test

import (
    "testing"

    "architecture-bricks/app/v5-value-objects/products/domain"

    "github.com/stretchr/testify/require"
)

func TestProductName(t *testing.T) {
    t.Parallel()

    t.Run("when_new_product_name_with_valid_name_then_success", func(t *testing.T) {
        t.Parallel()

        name, err := domain.NewProductName("Coffee")

        require.NoError(t, err)
        require.Equal(t, "Coffee", name.Value())
    })

    t.Run("when_new_product_name_trims_spaces", func(t *testing.T) {
        t.Parallel()

        name, err := domain.NewProductName("  Coffee  ")

        require.NoError(t, err)
        require.Equal(t, "Coffee", name.Value())
    })

    t.Run("when_new_product_name_with_empty_name_then_error", func(t *testing.T) {
        t.Parallel()

        _, err := domain.NewProductName("")

        require.ErrorIs(t, err, domain.ErrProductNameRequired)
    })

    t.Run("when_new_product_name_with_spaces_only_then_error", func(t *testing.T) {
        t.Parallel()

        _, err := domain.NewProductName("   ")

        require.ErrorIs(t, err, domain.ErrProductNameRequired)
    })
}
