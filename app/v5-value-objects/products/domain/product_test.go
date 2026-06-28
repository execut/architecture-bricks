package domain_test

import (
    "testing"

    "architecture-bricks/app/v5-value-objects/products/domain"
    "architecture-bricks/app/v5-value-objects/products/tests_common"
    ddd "architecture-bricks/pkg/domain-events/domain"

    "github.com/google/uuid"
    "github.com/stretchr/testify/require"
)

var _ ddd.Entity = (*domain.Product)(nil)

func TestProductLifecycle(t *testing.T) {
    t.Parallel()

    productID, err := domain.NewProductID(uuid.NewString())
    require.NoError(t, err)

    user, err := domain.NewUser(uuid.NewString())
    require.NoError(t, err)

    product := domain.NewProduct(productID)

    t.Run("create", func(t *testing.T) {
        err := product.Create(tests_common.TestProductName, user)

        require.NoError(t, err)
        require.Equal(t, productID.Value(), product.ID())
        require.Equal(t, tests_common.TestProductName, product.Name())
        require.Equal(t, domain.ProductStatusPending, product.Status())
        require.Empty(t, product.ApproveReason())
        require.Empty(t, product.RejectionReason())
    })

    t.Run("rename", func(t *testing.T) {
        err := product.Rename("Tea", user)

        require.NoError(t, err)
        require.Equal(t, productID.Value(), product.ID())
        require.Equal(t, "Tea", product.Name())
    })

    t.Run("rename_same_name", func(t *testing.T) {
        err := product.Rename("Tea", user)

        require.ErrorIs(t, err, domain.ErrProductNameNotChanged)
        require.Equal(t, "Tea", product.Name())
    })

    t.Run("approve", func(t *testing.T) {
        moderator, err := domain.NewModerator(uuid.NewString())
        require.NoError(t, err)

        product.Approve(moderator)

        require.Equal(t, domain.ProductStatusApproved, product.Status())
        require.Equal(t, domain.ProductApproveReasonModerator, product.ApproveReason())
        require.Empty(t, product.RejectionReason())
    })

    t.Run("reject", func(t *testing.T) {
        moderator, err := domain.NewModerator(uuid.NewString())
        require.NoError(t, err)

        product.Reject(moderator)

        require.Equal(t, domain.ProductStatusRejected, product.Status())
        require.Equal(t, domain.ProductRejectionReasonModerator, product.RejectionReason())
        require.Empty(t, product.ApproveReason())
    })

    t.Run("auto_approve", func(t *testing.T) {
        // Rename to cat triggers auto-approve inside Rename
        err := product.Rename(tests_common.TestProductNameCat, user)

        require.NoError(t, err)
        require.Equal(t, domain.ProductStatusApproved, product.Status())
        require.Equal(t, domain.ProductApproveReasonAuto, product.ApproveReason())
        require.Empty(t, product.RejectionReason())
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

    t.Run("create_empty_name", func(t *testing.T) {
        t.Parallel()

        product := domain.NewProduct(productID())
        err := product.Create("", user())

        require.ErrorIs(t, err, domain.ErrProductNameRequired)
    })

    t.Run("double_create", func(t *testing.T) {
        t.Parallel()

        product := domain.NewProduct(productID())

        err := product.Create(tests_common.TestProductName, user())
        require.NoError(t, err)

        err = product.Create("Another", user())
        require.ErrorIs(t, err, domain.ErrProductAlreadyExists)
    })

    t.Run("rename_empty_name", func(t *testing.T) {
        t.Parallel()

        product := domain.NewProduct(productID())

        err := product.Create(tests_common.TestProductName, user())
        require.NoError(t, err)

        err = product.Rename("", user())
        require.ErrorIs(t, err, domain.ErrProductNameRequired)
    })

    t.Run("auto_approve_on_create", func(t *testing.T) {
        t.Parallel()

        product := domain.NewProduct(productID())
        err := product.Create(tests_common.TestProductNameCat, user())

        require.NoError(t, err)
        require.Equal(t, domain.ProductStatusApproved, product.Status())
        require.Equal(t, domain.ProductApproveReasonAuto, product.ApproveReason())
    })

    t.Run("auto_approve_non_eligible", func(t *testing.T) {
        t.Parallel()

        product := domain.NewProduct(productID())
        err := product.Create("Ordinary", user())

        require.NoError(t, err)
        require.Equal(t, domain.ProductStatusPending, product.Status())
    })
}
