package domain_test

import (
    "testing"

    "architecture-bricks/app/v6-optimistic-locking/products/domain"
    "architecture-bricks/app/v6-optimistic-locking/products/tests_common"
    biz "architecture-bricks/pkg/optimistic-locking/business-events/value-objects/domain"
    vo "architecture-bricks/pkg/optimistic-locking/value-objects/domain"

    "github.com/google/uuid"
    "github.com/stretchr/testify/require"
)

var _ biz.Entity = (*domain.Product)(nil)

func TestProductLifecycle(t *testing.T) {
    t.Parallel()

    productID, err := domain.NewProductID(uuid.NewString())
    require.NoError(t, err)

    user, err := domain.NewUser(uuid.NewString())
    require.NoError(t, err)

    product := domain.NewProduct(productID)

    t.Run("when_create_then_product_created", func(t *testing.T) {
        err := product.Create(tests_common.TestProductName, user)

        require.NoError(t, err)
        require.Equal(t, productID.Value(), product.ID())
        require.Equal(t, tests_common.TestProductName, product.Name())
    })

    t.Run("when_rename_then_name_updated", func(t *testing.T) {
        err := product.Rename("Tea", user)

        require.NoError(t, err)
        require.Equal(t, productID.Value(), product.ID())
        require.Equal(t, "Tea", product.Name())
    })

    t.Run("when_rename_with_same_name_then_returns_error", func(t *testing.T) {
        err := product.Rename("Tea", user)

        require.ErrorIs(t, err, domain.ErrProductNameNotChanged)
        require.Equal(t, "Tea", product.Name())
    })
}

func TestProductVersion(t *testing.T) {
    t.Parallel()

    t.Run("when_new_product_then_version_is_zero", func(t *testing.T) {
        t.Parallel()

        productID, err := domain.NewProductID(uuid.NewString())
        require.NoError(t, err)

        product := domain.NewProduct(productID)

        v, err := vo.NewVersion(0)
        require.NoError(t, err)
        require.Equal(t, v, product.Version())
    })
}

func TestProductErrors(t *testing.T) {
    t.Parallel()

    productID := func() domain.ProductID {
        id, _ := domain.NewProductID(uuid.NewString())
        return id
    }

    user := func() domain.User {
        u, _ := domain.NewUser(uuid.NewString())
        return u
    }

    t.Run("when_create_with_empty_name_then_returns_error", func(t *testing.T) {
        t.Parallel()

        product := domain.NewProduct(productID())
        err := product.Create("", user())

        require.ErrorIs(t, err, domain.ErrProductNameRequired)
    })

    t.Run("when_double_create_then_returns_error", func(t *testing.T) {
        t.Parallel()

        product := domain.NewProduct(productID())

        err := product.Create(tests_common.TestProductName, user())
        require.NoError(t, err)

        err = product.Create("Another", user())
        require.ErrorIs(t, err, domain.ErrProductAlreadyExists)
    })

    t.Run("when_rename_with_empty_name_then_returns_error", func(t *testing.T) {
        t.Parallel()

        product := domain.NewProduct(productID())

        err := product.Create(tests_common.TestProductName, user())
        require.NoError(t, err)

        err = product.Rename("", user())
        require.ErrorIs(t, err, domain.ErrProductNameRequired)
    })
}
