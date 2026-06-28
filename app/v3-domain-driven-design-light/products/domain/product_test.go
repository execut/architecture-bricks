package domain_test

import (
    "testing"

    "architecture-bricks/app/v3-domain-driven-design-light/products/domain"
    testhelper "architecture-bricks/app/v3-domain-driven-design-light/products/tests_common"

    "github.com/stretchr/testify/require"
)

func TestProductLifecycle(t *testing.T) {
    t.Parallel()

    id := "test-id"
    var product *domain.Product

    t.Run("create", func(t *testing.T) {
        var err error

        product, err = domain.NewProduct(id, testhelper.TestUserID, testhelper.TestProductName)

        require.NoError(t, err)
        require.Equal(t, id, product.ID())
        require.Equal(t, testhelper.TestUserID, product.UserID())
        require.Equal(t, testhelper.TestProductName, product.Name())
        require.Equal(t, domain.ProductStatusPending, product.Status())
        require.Empty(t, product.ApproveReason())
        require.Empty(t, product.RejectionReason())
    })

    t.Run("get_id", func(t *testing.T) {
        require.Equal(t, id, product.ID())
    })

    t.Run("get_name", func(t *testing.T) {
        require.Equal(t, testhelper.TestProductName, product.Name())
    })

    t.Run("rename", func(t *testing.T) {
        err := product.Rename("Tea")

        require.NoError(t, err)
        require.Equal(t, id, product.ID())
        require.Equal(t, "Tea", product.Name())
    })

    t.Run("rename_same_name", func(t *testing.T) {
        err := product.Rename("Tea")

        require.ErrorIs(t, err, domain.ErrProductNameNotChanged)
    })

    t.Run("rename_empty_name", func(t *testing.T) {
        err := product.Rename("")

        require.ErrorIs(t, err, domain.ErrProductNameRequired)
    })

    t.Run("approve", func(t *testing.T) {
        product.Approve("moderator-1")

        require.Equal(t, domain.ProductStatusApproved, product.Status())
        require.Equal(t, domain.ProductApproveReasonModerator, product.ApproveReason())
        require.Empty(t, product.RejectionReason())
    })

    t.Run("reject", func(t *testing.T) {
        product.Reject("moderator-2")

        require.Equal(t, domain.ProductStatusRejected, product.Status())
        require.Equal(t, domain.ProductRejectionReasonModerator, product.RejectionReason())
        require.Empty(t, product.ApproveReason())
    })

    t.Run("auto_approve", func(t *testing.T) {
        product.AutoApprove()

        require.Equal(t, domain.ProductStatusApproved, product.Status())
        require.Equal(t, domain.ProductApproveReasonAuto, product.ApproveReason())
        require.Empty(t, product.RejectionReason())
    })
}

func TestProduct(t *testing.T) {
    t.Parallel()

    t.Run("create_empty_name", func(t *testing.T) {
        t.Parallel()

        _, err := domain.NewProduct("test-id", testhelper.TestUserID, "")

        require.ErrorIs(t, err, domain.ErrProductNameRequired)
    })
}
