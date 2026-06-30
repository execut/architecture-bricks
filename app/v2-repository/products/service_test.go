//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=repository.go -destination=tests/mocks/repository.go -package=mocks

package products_test

import (
	"context"
	"testing"

	"architecture-bricks/app/v2-repository/products"
	testdata "architecture-bricks/app/v2-repository/products/tests"
	"architecture-bricks/app/v2-repository/products/tests/mocks"
	"architecture-bricks/contract"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestService_ProductLifecycle(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	controller := gomock.NewController(t)
	mockRepository := mocks.NewMockRepository(controller)
	service := products.NewService(mockRepository)
	expectedProduct := contract.Product{
		ID:     testdata.TestProductID,
		Name:   testdata.TestProductName,
		UserID: testdata.TestUserID,
		Status: contract.ProductStatusPending,
	}
	createdEvent := contract.ProductHistoryRow{}
	moderatorID := uuid.NewString()

	t.Run("when_get_product_history_without_events_then_returns_empty_history", func(t *testing.T) {
		mockRepository.EXPECT().
			ProductHistory(ctx, testdata.TestProductID).
			Return([]contract.ProductHistoryRow{}, nil)

		history, err := service.ProductHistory(ctx, contract.ProductHistory{ProductID: testdata.TestProductID})

		require.NoError(t, err)
		require.Empty(t, history)
	})

	t.Run("when_create_product_then_product_is_created", func(t *testing.T) {
		mockRepository.EXPECT().
			CreateProduct(ctx, expectedProduct, gomock.Any()).
			DoAndReturn(func(
				ctx context.Context,
				product contract.Product,
				events ...contract.ProductHistoryRow,
			) error {
				if len(events) > 0 {
					createdEvent = events[0]
				}

				return nil
			})

		err := service.CreateProduct(ctx, contract.CreateProduct{
			ProductID: testdata.TestProductID,
			UserID:    testdata.TestUserID,
			Name:      testdata.TestProductName,
		})

		require.NoError(t, err)
		require.NotEmpty(t, createdEvent.ID)
		require.Equal(t, contract.EventProductCreated, createdEvent.Event)
		require.Equal(t, testdata.TestProductName, createdEvent.Name)
	})

	t.Run("when_get_product_then_returns_created_product", func(t *testing.T) {
		mockRepository.EXPECT().
			GetProduct(ctx, testdata.TestProductID).
			Return(expectedProduct, 1, nil)

		product, err := service.GetProduct(ctx, contract.GetProduct{ProductID: testdata.TestProductID})

		require.NoError(t, err)
		require.Equal(t, expectedProduct, product)
	})

	t.Run("when_get_product_history_then_returns_created_event", func(t *testing.T) {
		mockRepository.EXPECT().
			ProductHistory(ctx, testdata.TestProductID).
			Return([]contract.ProductHistoryRow{createdEvent}, nil)

		history, err := service.ProductHistory(ctx, contract.ProductHistory{ProductID: testdata.TestProductID})

		require.NoError(t, err)
		require.Len(t, history, 1)
		row := history[0]
		require.NotEmpty(t, row.ID)
		require.Equal(t, contract.EventProductCreated, row.Event)
		require.Equal(t, testdata.TestProductName, row.Name)
	})

	t.Run("when_approve_product_then_product_is_approved", func(t *testing.T) {
		mockRepository.EXPECT().
			GetProduct(ctx, testdata.TestProductID).
			Return(expectedProduct, 1, nil)

		mockRepository.EXPECT().
			ApproveProduct(ctx, testdata.TestProductID, moderatorID, 1, gomock.Any()).
			Return(nil)

		err := service.ApproveProduct(ctx, contract.ApproveProduct{
			ProductID:   testdata.TestProductID,
			ModeratorID: moderatorID,
		})

		require.NoError(t, err)
	})

	t.Run("when_reject_product_then_product_is_rejected", func(t *testing.T) {
		mockRepository.EXPECT().
			GetProduct(ctx, testdata.TestProductID).
			Return(expectedProduct, 1, nil)

		mockRepository.EXPECT().
			RejectProduct(ctx, testdata.TestProductID, moderatorID, 1, gomock.Any()).
			Return(nil)

		err := service.RejectProduct(ctx, contract.RejectProduct{
			ProductID:   testdata.TestProductID,
			ModeratorID: moderatorID,
		})

		require.NoError(t, err)
	})

	t.Run("when_create_auto_approvable_product_then_creates_with_auto_approve", func(t *testing.T) {
		autoApproveProductID := uuid.NewString()
		expectedAutoApproveProduct := contract.Product{
			ID:            autoApproveProductID,
			Name:          testdata.TestProductNameCat,
			UserID:        testdata.TestUserID,
			Status:        contract.ProductStatusApproved,
			ApproveReason: contract.ProductApproveReasonAuto,
		}

		mockRepository.EXPECT().
			CreateProduct(ctx, expectedAutoApproveProduct, gomock.Any(), gomock.Any()).
			Return(nil)

		err := service.CreateProduct(ctx, contract.CreateProduct{
			ProductID: autoApproveProductID,
			UserID:    testdata.TestUserID,
			Name:      testdata.TestProductNameCat,
		})

		require.NoError(t, err)
	})
}
