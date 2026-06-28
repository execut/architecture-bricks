package domain_test

import (
    "testing"

    "architecture-bricks/app/v4-domain-driven-design-light-with-events/products/domain"
    "architecture-bricks/app/v4-domain-driven-design-light-with-events/products/tests_common"
    ddd "architecture-bricks/pkg/domain-events/domain"

    "github.com/stretchr/testify/require"
)

var _ ddd.Entity = (*domain.Product)(nil)

func TestProductLifecycle(t *testing.T) {
    t.Parallel()

    product := domain.NewProduct("test-id")

    t.Run("create", func(t *testing.T) {
        err := product.Create(tests_common.TestProductName, tests_common.TestUserID)

        require.NoError(t, err)
        require.Equal(t, "test-id", product.ID())
        require.Equal(t, tests_common.TestUserID, product.UserID())
        require.Equal(t, tests_common.TestProductName, product.Name())
        require.Equal(t, domain.ProductStatusPending, product.Status())
        require.Empty(t, product.ApproveReason())
        require.Empty(t, product.RejectionReason())
    })

    t.Run("rename", func(t *testing.T) {
        err := product.Rename("Tea", tests_common.TestUserID)

        require.NoError(t, err)
        require.Equal(t, "test-id", product.ID())
        require.Equal(t, "Tea", product.Name())
    })

    t.Run("rename_same_name", func(t *testing.T) {
        err := product.Rename("Tea", tests_common.TestUserID)

        require.ErrorIs(t, err, domain.ErrProductNameNotChanged)
        require.Equal(t, "Tea", product.Name())
    })

    t.Run("approve", func(t *testing.T) {
        product.Approve("moderator-1")

        require.Equal(t, domain.ProductStatusApproved, product.Status())
        require.Equal(t, domain.ProductApproveReasonModerator, product.ApproveReason())
        require.Empty(t, product.RejectionReason())
    })

    t.Run("reject", func(t *testing.T) {
        // Need to advance from approved to rejected state
        product.Reject("moderator-2")

        require.Equal(t, domain.ProductStatusRejected, product.Status())
        require.Equal(t, domain.ProductRejectionReasonModerator, product.RejectionReason())
        require.Empty(t, product.ApproveReason())
    })

    t.Run("auto_approve", func(t *testing.T) {
        // Rename to cat triggers auto-approve inside Rename
        err := product.Rename(tests_common.TestProductNameCat, tests_common.TestUserID)

        require.NoError(t, err)
        require.Equal(t, domain.ProductStatusApproved, product.Status())
        require.Equal(t, domain.ProductApproveReasonAuto, product.ApproveReason())
        require.Empty(t, product.RejectionReason())
    })
}

func TestProductErrors(t *testing.T) {
    t.Parallel()

    t.Run("create_empty_name", func(t *testing.T) {
        t.Parallel()

        product := domain.NewProduct("test-id")
        err := product.Create("", "user-1")

        require.ErrorIs(t, err, domain.ErrProductNameRequired)
    })

    t.Run("double_create", func(t *testing.T) {
        t.Parallel()

        product := domain.NewProduct("test-id")

        err := product.Create(tests_common.TestProductName, tests_common.TestUserID)
        require.NoError(t, err)

        err = product.Create("Another", tests_common.TestUserID)
        require.ErrorIs(t, err, domain.ErrProductAlreadyExists)
    })

    t.Run("rename_empty_name", func(t *testing.T) {
        t.Parallel()

        product := domain.NewProduct("test-id")

        err := product.Create(tests_common.TestProductName, tests_common.TestUserID)
        require.NoError(t, err)

        err = product.Rename("", tests_common.TestUserID)
        require.ErrorIs(t, err, domain.ErrProductNameRequired)
    })

    t.Run("auto_approve_if_eligible_non_auto", func(t *testing.T) {
        t.Parallel()

        product := domain.NewProduct("test-id")
        err := product.Create("Ordinary", "user-1")
        require.NoError(t, err)

        // Product was not auto-approved because name is ordinary
        require.Equal(t, domain.ProductStatusPending, product.Status())
    })

    t.Run("auto_approve_if_eligible_on_rename", func(t *testing.T) {
        t.Parallel()

        product := domain.NewProduct("test-id")
        err := product.Create("Ordinary", "user-1")
        require.NoError(t, err)
        require.Equal(t, domain.ProductStatusPending, product.Status())

        // Rename triggers auto-approve
        err = product.Rename(tests_common.TestProductNameCat, "user-1")
        require.NoError(t, err)

        require.Equal(t, domain.ProductStatusApproved, product.Status())
        require.Equal(t, domain.ProductApproveReasonAuto, product.ApproveReason())
    })

    t.Run("auto_approve_if_eligible_on_create", func(t *testing.T) {
        t.Parallel()

        product := domain.NewProduct("test-id")
        err := product.Create(tests_common.TestProductNameCat, "user-1")
        require.NoError(t, err)

        require.Equal(t, domain.ProductStatusApproved, product.Status())
        require.Equal(t, domain.ProductApproveReasonAuto, product.ApproveReason())
    })
}
