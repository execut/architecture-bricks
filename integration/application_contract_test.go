package integration_test

import (
    "context"
    "testing"

    "architecture-bricks/contract"
    testhelper "architecture-bricks/integration/tests"

    "github.com/google/uuid"
    "github.com/stretchr/testify/require"
)

func TestApplication_ProductLifecycle(t *testing.T) {
    t.Parallel()

    for _, variant := range testhelper.TestApplicationVariantList(t) {
        t.Run("when_use_"+variant+"_variant_then_product_lifecycle_completed", func(t *testing.T) {
            t.Parallel()

            testApplicationProductLifecycle(t, variant)
        })
    }
}

func testApplicationProductLifecycle(t *testing.T, variant string) {
    t.Helper()

    databaseURL := testhelper.TestDatabaseURL(t)
    ctx := context.Background()
    testhelper.WaitDB(t, ctx, databaseURL)

    application := testhelper.NewTestApplication(t, ctx, variant)
    productID := uuid.NewString()
    userID := uuid.NewString()
    moderatorID := uuid.NewString()
    newName := "Tea"
    autoApprovedName := testhelper.TestProductNameCat
    t.Cleanup(func() {
        testhelper.CleanupProduct(t, ctx, databaseURL, productID)
    })

    t.Run("get_empty_history", func(t *testing.T) {
        history, err := application.ProductHistory(ctx, contract.ProductHistory{ProductID: productID})

        require.NoError(t, err)
        require.Empty(t, history)
    })

    t.Run("create", func(t *testing.T) {
        err := application.CreateProduct(ctx, contract.CreateProduct{
            ProductID: productID,
            UserID:    userID,
            Name:      testhelper.TestProductName,
        })

        require.NoError(t, err)
    })

    t.Run("get_created", func(t *testing.T) {
        product, err := application.GetProduct(ctx, contract.GetProduct{ProductID: productID})

        require.NoError(t, err)
        require.Equal(t, contract.Product{
            ID:     productID,
            Name:   testhelper.TestProductName,
            UserID: userID,
            Status: contract.ProductStatusPending,
        }, product)
    })

    t.Run("update_name", func(t *testing.T) {
        err := application.UpdateProduct(ctx, contract.UpdateProduct{
            ProductID: productID,
            UserID:    userID,
            Name:      newName,
        })

        require.NoError(t, err)
    })

    t.Run("get_updated", func(t *testing.T) {
        product, err := application.GetProduct(ctx, contract.GetProduct{ProductID: productID})

        require.NoError(t, err)
        require.Equal(t, contract.Product{
            ID:     productID,
            Name:   newName,
            UserID: userID,
            Status: contract.ProductStatusPending,
        }, product)
    })

    t.Run("approve", func(t *testing.T) {
        err := application.ApproveProduct(ctx, contract.ApproveProduct{
            ProductID:   productID,
            ModeratorID: moderatorID,
        })

        require.NoError(t, err)
    })

    t.Run("reject", func(t *testing.T) {
        err := application.RejectProduct(ctx, contract.RejectProduct{
            ProductID:   productID,
            ModeratorID: moderatorID,
        })

        require.NoError(t, err)
    })

    t.Run("update_to_auto_approvable", func(t *testing.T) {
        err := application.UpdateProduct(ctx, contract.UpdateProduct{
            ProductID: productID,
            UserID:    userID,
            Name:      autoApprovedName,
        })

        require.NoError(t, err)
    })

    t.Run("get_auto_approved", func(t *testing.T) {
        product, err := application.GetProduct(ctx, contract.GetProduct{ProductID: productID})

        require.NoError(t, err)
        require.Equal(t, contract.Product{
            ID:            productID,
            Name:          autoApprovedName,
            UserID:        userID,
            Status:        contract.ProductStatusApproved,
            ApproveReason: contract.ProductApproveReasonAuto,
        }, product)
    })

    t.Run("get_history", func(t *testing.T) {
        history, err := application.ProductHistory(ctx, contract.ProductHistory{ProductID: productID})

        require.NoError(t, err)
        require.Len(t, history, 6)

        createEvent := history[0]
        require.Equal(t, contract.EventProductCreated, createEvent.Event)
        require.Equal(t, testhelper.TestProductName, createEvent.Name)

        updateEvent := history[1]
        require.Equal(t, contract.EventProductUpdated, updateEvent.Event)
        require.Equal(t, newName, updateEvent.Name)

        approveEvent := history[2]
        require.Equal(t, contract.EventProductApproved, approveEvent.Event)
        require.Equal(t, moderatorID, approveEvent.ModeratorID)
        require.Equal(t, newName, approveEvent.Name)
        require.Equal(t, contract.ProductApproveReasonModerator, approveEvent.ApproveReason)

        rejectEvent := history[3]
        require.Equal(t, contract.EventProductRejected, rejectEvent.Event)
        require.Equal(t, moderatorID, rejectEvent.ModeratorID)
        require.Equal(t, newName, rejectEvent.Name)
        require.Equal(t, contract.ProductRejectionReasonModerator, rejectEvent.RejectionReason)

        autoApproveUpdateEvent := history[4]
        require.Equal(t, contract.EventProductUpdated, autoApproveUpdateEvent.Event)
        require.Equal(t, autoApprovedName, autoApproveUpdateEvent.Name)

        autoApproveEvent := history[5]
        require.Equal(t, contract.EventProductAutoApproved, autoApproveEvent.Event)
        require.Equal(t, autoApprovedName, autoApproveEvent.Name)
        require.Equal(t, contract.ProductApproveReasonAuto, autoApproveEvent.ApproveReason)
    })

    t.Run("update_same_name_error", func(t *testing.T) {
        err := application.UpdateProduct(ctx, contract.UpdateProduct{
            ProductID: productID,
            UserID:    userID,
            Name:      autoApprovedName,
        })

        require.ErrorIs(t, err, contract.ErrProductNameNotChanged)
    })
}

func TestApplication_ProductAutoApproveLifecycle(t *testing.T) {
    t.Parallel()

    for _, variant := range testhelper.TestApplicationVariantList(t) {
        t.Run("when_use_"+variant+"_variant_then_auto_approve_lifecycle_completed", func(t *testing.T) {
            t.Parallel()

            testApplicationProductAutoApproveLifecycle(t, variant)
        })
    }
}

func testApplicationProductAutoApproveLifecycle(t *testing.T, variant string) {
    t.Helper()

    databaseURL := testhelper.TestDatabaseURL(t)
    ctx := context.Background()
    testhelper.WaitDB(t, ctx, databaseURL)

    application := testhelper.NewTestApplication(t, ctx, variant)
    userID := uuid.NewString()

    testCaseList := []struct {
        name        string
        productName string
    }{
        {name: "cat", productName: testhelper.TestProductNameCat},
        {name: "dog", productName: testhelper.TestProductNameDog},
    }

    for _, testCase := range testCaseList {
        t.Run(testCase.name, func(t *testing.T) {
            productID := uuid.NewString()
            t.Cleanup(func() {
                testhelper.CleanupProduct(t, ctx, databaseURL, productID)
            })

            t.Run("create", func(t *testing.T) {
                err := application.CreateProduct(ctx, contract.CreateProduct{
                    ProductID: productID,
                    UserID:    userID,
                    Name:      testCase.productName,
                })

                require.NoError(t, err)
            })

            t.Run("get_approved", func(t *testing.T) {
                product, err := application.GetProduct(ctx, contract.GetProduct{ProductID: productID})

                require.NoError(t, err)
                require.Equal(t, contract.Product{
                    ID:            productID,
                    Name:          testCase.productName,
                    UserID:        userID,
                    Status:        contract.ProductStatusApproved,
                    ApproveReason: contract.ProductApproveReasonAuto,
                }, product)
            })

            t.Run("get_history", func(t *testing.T) {
                history, err := application.ProductHistory(ctx, contract.ProductHistory{ProductID: productID})

                require.NoError(t, err)
                require.Len(t, history, 2)

                createEvent := history[0]
                require.Equal(t, contract.EventProductCreated, createEvent.Event)
                require.Equal(t, testCase.productName, createEvent.Name)

                autoApproveEvent := history[1]
                require.Equal(t, contract.EventProductAutoApproved, autoApproveEvent.Event)
                require.Equal(t, testCase.productName, autoApproveEvent.Name)
                require.Equal(t, contract.ProductApproveReasonAuto, autoApproveEvent.ApproveReason)
            })
        })
    }
}

func TestApplication_ProductErrors(t *testing.T) {
    t.Parallel()

    for _, variant := range testhelper.TestApplicationVariantList(t) {
        t.Run("when_use_"+variant+"_variant_then_errors_handled", func(t *testing.T) {
            t.Parallel()

            testApplicationProductErrors(t, variant)
        })
    }
}

func testApplicationProductErrors(t *testing.T, variant string) {
    t.Helper()

    databaseURL := testhelper.TestDatabaseURL(t)
    ctx := context.Background()
    testhelper.WaitDB(t, ctx, databaseURL)

    application := testhelper.NewTestApplication(t, ctx, variant)
    productID := uuid.NewString()
    userID := uuid.NewString()
    t.Cleanup(func() {
        testhelper.CleanupProduct(t, ctx, databaseURL, productID)
    })

    t.Run("when_create_product_with_invalid_product_id_then_returns_error", func(t *testing.T) {
        err := application.CreateProduct(ctx, contract.CreateProduct{
            ProductID: "not-a-uuid",
            UserID:    userID,
            Name:      testhelper.TestProductName,
        })

        require.ErrorIs(t, err, contract.ErrInvalidProductID)
    })

    t.Run("when_create_product_with_invalid_user_id_then_returns_error", func(t *testing.T) {
        err := application.CreateProduct(ctx, contract.CreateProduct{
            ProductID: productID,
            UserID:    "not-a-uuid",
            Name:      testhelper.TestProductName,
        })

        require.ErrorIs(t, err, contract.ErrInvalidUserID)
    })

    t.Run("when_create_product_with_empty_name_then_returns_error", func(t *testing.T) {
        err := application.CreateProduct(ctx, contract.CreateProduct{
            ProductID: uuid.NewString(),
            UserID:    userID,
            Name:      "",
        })

        require.ErrorIs(t, err, contract.ErrProductNameRequired)
    })

    t.Run("when_create_product_twice_then_returns_error", func(t *testing.T) {
        err := application.CreateProduct(ctx, contract.CreateProduct{
            ProductID: productID,
            UserID:    userID,
            Name:      testhelper.TestProductName,
        })

        require.NoError(t, err)

        err = application.CreateProduct(ctx, contract.CreateProduct{
            ProductID: productID,
            UserID:    userID,
            Name:      "Another",
        })

        require.ErrorIs(t, err, contract.ErrProductAlreadyExists)
    })

    t.Run("when_get_nonexistent_product_then_returns_error", func(t *testing.T) {
        _, err := application.GetProduct(ctx, contract.GetProduct{
            ProductID: uuid.NewString(),
        })

        require.ErrorIs(t, err, contract.ErrProductNotFound)
    })

    t.Run("when_update_nonexistent_product_then_returns_error", func(t *testing.T) {
        err := application.UpdateProduct(ctx, contract.UpdateProduct{
            ProductID: uuid.NewString(),
            UserID:    userID,
            Name:      "NewName",
        })

        require.ErrorIs(t, err, contract.ErrProductNotFound)
    })

    t.Run("when_get_history_with_invalid_product_id_then_returns_error", func(t *testing.T) {
        _, err := application.ProductHistory(ctx, contract.ProductHistory{
            ProductID: "not-a-uuid",
        })

        require.ErrorIs(t, err, contract.ErrInvalidProductID)
    })
}
