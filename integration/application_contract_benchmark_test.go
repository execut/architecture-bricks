package integration_test

import (
	"context"
	"strconv"
	"testing"

	"architecture-bricks/contract"
	testhelper "architecture-bricks/integration/tests"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const benchmarkUpdatedProductName = "Tea"

type benchmarkApplicationProductLifecycleFixture struct {
	ctx         context.Context
	databaseURL string
	application contract.Application
	userID      string
	moderatorID string
}

func newBenchmarkApplicationProductLifecycleFixture(
	b *testing.B,
	variant string,
) benchmarkApplicationProductLifecycleFixture {
	b.Helper()

	databaseURL := testhelper.TestDatabaseURL(b)
	ctx := context.Background()
	testhelper.WaitDB(b, ctx, databaseURL)

	return benchmarkApplicationProductLifecycleFixture{
		ctx:         ctx,
		databaseURL: databaseURL,
		application: testhelper.NewTestApplication(b, ctx, variant),
		userID:      uuid.NewString(),
		moderatorID: uuid.NewString(),
	}
}

func BenchmarkApplication_ProductLifecycle(b *testing.B) {
	for _, variant := range testhelper.TestApplicationVariantList(b) {
		b.Run("when_use_"+variant+"_variant", func(b *testing.B) {
			fixture := newBenchmarkApplicationProductLifecycleFixture(b, variant)

			b.Run("when_get_product_history_without_events", func(b *testing.B) {
				b.ReportAllocs()

				productID := uuid.NewString()

				b.ResetTimer()
				for range b.N {
					history, err := fixture.application.ProductHistory(
						fixture.ctx,
						contract.ProductHistory{ProductID: productID},
					)

					require.NoError(b, err)
					require.Empty(b, history)
				}
			})

			b.Run("when_create_product", func(b *testing.B) {
				b.ReportAllocs()

				productIDList := make([]string, 0, b.N)
				b.Cleanup(func() {
					testhelper.CleanupProducts(b, fixture.ctx, fixture.databaseURL, productIDList)
				})

				b.ResetTimer()
				for range b.N {
					b.StopTimer()
					productID := uuid.NewString()
					productIDList = append(productIDList, productID)
					b.StartTimer()

					err := fixture.application.CreateProduct(fixture.ctx, contract.CreateProduct{
						ProductID: productID,
						UserID:    fixture.userID,
						Name:      testhelper.TestProductName,
					})

					require.NoError(b, err)
				}
			})

			b.Run("when_get_product", func(b *testing.B) {
				b.ReportAllocs()

				productID := createBenchmarkProduct(b, fixture, testhelper.TestProductName)
				b.Cleanup(func() {
					testhelper.CleanupProduct(b, fixture.ctx, fixture.databaseURL, productID)
				})

				b.ResetTimer()
				for range b.N {
					product, err := fixture.application.GetProduct(
						fixture.ctx,
						contract.GetProduct{ProductID: productID},
					)

					require.NoError(b, err)
					require.Equal(b, contract.Product{
						ID:     productID,
						Name:   testhelper.TestProductName,
						UserID: fixture.userID,
						Status: contract.ProductStatusPending,
					}, product)
				}
			})

			b.Run("when_get_product_history_with_create_event", func(b *testing.B) {
				b.ReportAllocs()

				productID := createBenchmarkProduct(b, fixture, testhelper.TestProductName)
				b.Cleanup(func() {
					testhelper.CleanupProduct(b, fixture.ctx, fixture.databaseURL, productID)
				})

				b.ResetTimer()
				for range b.N {
					history, err := fixture.application.ProductHistory(
						fixture.ctx,
						contract.ProductHistory{ProductID: productID},
					)

					require.NoError(b, err)
					require.Len(b, history, 1)
				}
			})

			b.Run("when_update_product_name", func(b *testing.B) {
				b.ReportAllocs()

				productID := createBenchmarkProduct(b, fixture, testhelper.TestProductName)
				b.Cleanup(func() {
					testhelper.CleanupProduct(b, fixture.ctx, fixture.databaseURL, productID)
				})

				b.ResetTimer()
				for i := range b.N {
					b.StopTimer()
					name := benchmarkUpdatedProductName + strconv.Itoa(i)
					b.StartTimer()

					err := fixture.application.UpdateProduct(fixture.ctx, contract.UpdateProduct{
						ProductID: productID,
						UserID:    fixture.userID,
						Name:      name,
					})

					require.NoError(b, err)
				}
			})

			b.Run("when_get_product_after_update", func(b *testing.B) {
				b.ReportAllocs()

				productID := createBenchmarkUpdatedProduct(b, fixture)
				b.Cleanup(func() {
					testhelper.CleanupProduct(b, fixture.ctx, fixture.databaseURL, productID)
				})

				b.ResetTimer()
				for range b.N {
					product, err := fixture.application.GetProduct(
						fixture.ctx,
						contract.GetProduct{ProductID: productID},
					)

					require.NoError(b, err)
					require.Equal(b, contract.Product{
						ID:     productID,
						Name:   benchmarkUpdatedProductName,
						UserID: fixture.userID,
						Status: contract.ProductStatusPending,
					}, product)
				}
			})

			b.Run("when_approve_product", func(b *testing.B) {
				b.ReportAllocs()

				productID := createBenchmarkUpdatedProduct(b, fixture)
				b.Cleanup(func() {
					testhelper.CleanupProduct(b, fixture.ctx, fixture.databaseURL, productID)
				})

				b.ResetTimer()
				for range b.N {
					err := fixture.application.ApproveProduct(fixture.ctx, contract.ApproveProduct{
						ProductID:   productID,
						ModeratorID: fixture.moderatorID,
					})

					require.NoError(b, err)
				}
			})

			b.Run("when_reject_product", func(b *testing.B) {
				b.ReportAllocs()

				productID := createBenchmarkUpdatedProduct(b, fixture)
				b.Cleanup(func() {
					testhelper.CleanupProduct(b, fixture.ctx, fixture.databaseURL, productID)
				})

				b.ResetTimer()
				for range b.N {
					err := fixture.application.RejectProduct(fixture.ctx, contract.RejectProduct{
						ProductID:   productID,
						ModeratorID: fixture.moderatorID,
					})

					require.NoError(b, err)
				}
			})

			b.Run("when_update_product_to_auto_approvable", func(b *testing.B) {
				b.ReportAllocs()

				productIDList := make([]string, 0, b.N)
				b.Cleanup(func() {
					testhelper.CleanupProducts(b, fixture.ctx, fixture.databaseURL, productIDList)
				})

				b.ResetTimer()
				for range b.N {
					b.StopTimer()
					productID := createBenchmarkProduct(b, fixture, testhelper.TestProductName)
					productIDList = append(productIDList, productID)
					b.StartTimer()

					err := fixture.application.UpdateProduct(fixture.ctx, contract.UpdateProduct{
						ProductID: productID,
						UserID:    fixture.userID,
						Name:      testhelper.TestProductNameCat,
					})

					require.NoError(b, err)
				}
			})

			b.Run("when_get_product_after_auto_approve", func(b *testing.B) {
				b.ReportAllocs()

				productID := createBenchmarkAutoApprovedProduct(b, fixture)
				b.Cleanup(func() {
					testhelper.CleanupProduct(b, fixture.ctx, fixture.databaseURL, productID)
				})

				b.ResetTimer()
				for range b.N {
					product, err := fixture.application.GetProduct(
						fixture.ctx,
						contract.GetProduct{ProductID: productID},
					)

					require.NoError(b, err)
					require.Equal(b, contract.Product{
						ID:            productID,
						Name:          testhelper.TestProductNameCat,
						UserID:        fixture.userID,
						Status:        contract.ProductStatusApproved,
						ApproveReason: contract.ProductApproveReasonAuto,
					}, product)
				}
			})

			b.Run("when_get_product_history_with_create_and_update_events", func(b *testing.B) {
				b.ReportAllocs()

				productID := createBenchmarkUpdatedProduct(b, fixture)
				b.Cleanup(func() {
					testhelper.CleanupProduct(b, fixture.ctx, fixture.databaseURL, productID)
				})

				b.ResetTimer()
				for range b.N {
					history, err := fixture.application.ProductHistory(
						fixture.ctx,
						contract.ProductHistory{ProductID: productID},
					)

					require.NoError(b, err)
					require.Len(b, history, 2)
				}
			})

			b.Run("when_get_product_history_with_full_lifecycle_events", func(b *testing.B) {
				b.ReportAllocs()

				productID := createBenchmarkLifecycleProduct(b, fixture)
				b.Cleanup(func() {
					testhelper.CleanupProduct(b, fixture.ctx, fixture.databaseURL, productID)
				})

				b.ResetTimer()
				for range b.N {
					history, err := fixture.application.ProductHistory(
						fixture.ctx,
						contract.ProductHistory{ProductID: productID},
					)

					require.NoError(b, err)
					require.Len(b, history, 6)
				}
			})

			b.Run("when_update_product_with_same_name", func(b *testing.B) {
				b.ReportAllocs()

				productID := createBenchmarkAutoApprovedProduct(b, fixture)
				b.Cleanup(func() {
					testhelper.CleanupProduct(b, fixture.ctx, fixture.databaseURL, productID)
				})

				b.ResetTimer()
				for range b.N {
					err := fixture.application.UpdateProduct(fixture.ctx, contract.UpdateProduct{
						ProductID: productID,
						UserID:    fixture.userID,
						Name:      testhelper.TestProductNameCat,
					})

					require.ErrorIs(b, err, contract.ErrProductNameNotChanged)
				}
			})
		})
	}
}

func createBenchmarkProduct(
	b *testing.B,
	fixture benchmarkApplicationProductLifecycleFixture,
	name string,
) string {
	b.Helper()

	productID := uuid.NewString()
	err := fixture.application.CreateProduct(fixture.ctx, contract.CreateProduct{
		ProductID: productID,
		UserID:    fixture.userID,
		Name:      name,
	})
	require.NoError(b, err)

	return productID
}

func createBenchmarkUpdatedProduct(
	b *testing.B,
	fixture benchmarkApplicationProductLifecycleFixture,
) string {
	b.Helper()

	productID := createBenchmarkProduct(b, fixture, testhelper.TestProductName)
	err := fixture.application.UpdateProduct(fixture.ctx, contract.UpdateProduct{
		ProductID: productID,
		UserID:    fixture.userID,
		Name:      benchmarkUpdatedProductName,
	})
	require.NoError(b, err)

	return productID
}

func createBenchmarkAutoApprovedProduct(
	b *testing.B,
	fixture benchmarkApplicationProductLifecycleFixture,
) string {
	b.Helper()

	productID := createBenchmarkUpdatedProduct(b, fixture)
	err := fixture.application.UpdateProduct(fixture.ctx, contract.UpdateProduct{
		ProductID: productID,
		UserID:    fixture.userID,
		Name:      testhelper.TestProductNameCat,
	})
	require.NoError(b, err)

	return productID
}

func createBenchmarkLifecycleProduct(
	b *testing.B,
	fixture benchmarkApplicationProductLifecycleFixture,
) string {
	b.Helper()

	productID := createBenchmarkUpdatedProduct(b, fixture)

	err := fixture.application.ApproveProduct(fixture.ctx, contract.ApproveProduct{
		ProductID:   productID,
		ModeratorID: fixture.moderatorID,
	})
	require.NoError(b, err)

	err = fixture.application.RejectProduct(fixture.ctx, contract.RejectProduct{
		ProductID:   productID,
		ModeratorID: fixture.moderatorID,
	})
	require.NoError(b, err)

	err = fixture.application.UpdateProduct(fixture.ctx, contract.UpdateProduct{
		ProductID: productID,
		UserID:    fixture.userID,
		Name:      testhelper.TestProductNameCat,
	})
	require.NoError(b, err)

	return productID
}
