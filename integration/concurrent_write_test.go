package integration_test

import (
    "context"
    "errors"
    "fmt"
    "sync"
    "testing"

    "architecture-bricks/contract"
    testhelper "architecture-bricks/integration/tests"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestApplication_ConcurrentCreate(t *testing.T) {
    t.Parallel()

    for _, variant := range testhelper.TestApplicationVariantList(t) {
        t.Run("when_use_"+variant+"_variant_then_concurrent_create_handled", func(t *testing.T) {
            t.Parallel()

            testConcurrentCreate(t, variant)
        })
    }
}

func testConcurrentCreate(t *testing.T, variant string) {
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

    const goroutines = 10
    var wg sync.WaitGroup
    var mu sync.Mutex
    successCount := 0
    alreadyExistsCount := 0
    failed := false

    for range goroutines {
        wg.Add(1)
        go func() {
            defer wg.Done()

            err := application.CreateProduct(ctx, contract.CreateProduct{
                ProductID: productID,
                UserID:    userID,
                Name:      testhelper.TestProductName,
            })

            mu.Lock()
            switch {
            case err == nil:
                successCount++
            case errors.Is(err, contract.ErrProductAlreadyExists):
                alreadyExistsCount++
            default:
                failed = true
                t.Error(err)
            }
            mu.Unlock()
        }()
    }

    wg.Wait()

    if failed {
        t.FailNow()
    }
    assert.Equal(t, 1, successCount, "ровно одна горутина должна создать продукт при %s", variant)
    assert.Equal(t, goroutines-1, alreadyExistsCount, "остальные горутины должны получить ErrProductAlreadyExists при %s", variant)
}

func TestApplication_ConcurrentUpdate(t *testing.T) {
    t.Parallel()

    for _, variant := range testhelper.TestApplicationVariantList(t) {
        t.Run("when_use_"+variant+"_variant_then_concurrent_update_handled", func(t *testing.T) {
            t.Parallel()

            testConcurrentUpdate(t, variant)
        })
    }
}

func testConcurrentUpdate(t *testing.T, variant string) {
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

    err := application.CreateProduct(ctx, contract.CreateProduct{
        ProductID: productID,
        UserID:    userID,
        Name:      testhelper.TestProductName,
    })
    require.NoError(t, err, "создание продукта должно пройти перед конкурентным обновлением")

    const goroutines = 10
    var wg sync.WaitGroup
    var mu sync.Mutex
    successCount := 0
    conflictCount := 0
    failed := false

    for i := range goroutines {
        wg.Add(1)
        go func(idx int) {
            defer wg.Done()

            err := application.UpdateProduct(ctx, contract.UpdateProduct{
                ProductID: productID,
                UserID:    userID,
                Name:      fmt.Sprintf("Name-%d", idx),
            })

            mu.Lock()
            switch {
            case err == nil:
                successCount++
            case errors.Is(err, contract.ErrProductAlreadyChanged):
                conflictCount++
            default:
                failed = true
                t.Error(err)
            }
            mu.Unlock()
        }(i)
    }

    wg.Wait()

    if failed {
        t.FailNow()
    }
    assert.GreaterOrEqual(t, successCount, 1, "хотя бы одна горутина должна успешно обновить продукт при %s", variant)
    if testhelper.SupportsConflict(variant) {
        assert.GreaterOrEqual(t, conflictCount, 1, "хотя бы одна горутина должна получить ErrProductAlreadyChanged при %s", variant)
    }
    assert.Equal(t, goroutines, successCount+conflictCount, "все горутины должны либо обновить продукт, либо получить конфликт при %s", variant)
}
