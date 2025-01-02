package inmemory

import (
	"prediction-risk/internal/domain/entities"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStopOrderRepoInMemory(t *testing.T) {
	t.Run("GetByID", func(t *testing.T) {
		t.Run("returns order when found", func(t *testing.T) {
			// Arrange
			repo := NewStopOrderRepoInMemory()
			threshold, err := entities.NewContractPrice(20)
			require.NoError(t, err)
			order := entities.NewStopOrder("FOO", entities.SideYes, threshold, nil, nil)
			require.NoError(t, repo.Persist(order), "Failed to persist order")

			// Act
			foundOrder, err := repo.GetByID(order.ID())

			// Assert
			require.NoError(t, err)
			require.NotNil(t, foundOrder)
			assert.Equal(t, order.ID(), foundOrder.ID())
			assert.Equal(t, order.Ticker(), foundOrder.Ticker())
			assert.Equal(t, order.TriggerPrice(), foundOrder.TriggerPrice())
		})

		t.Run("returns error ErrNotFound when not found", func(t *testing.T) {
			// Arrange
			repo := NewStopOrderRepoInMemory()
			id := entities.NewOrderID()

			// Act
			order, err := repo.GetByID(id)

			expectedErr := &entities.ErrNotFound{
				Entity: "StopOrder",
				ID:     id.String(),
			}

			// Assert
			require.Error(t, err)
			require.Nil(t, order)
			assert.Equal(t, expectedErr, err)
		})
	})

	t.Run("Persist", func(t *testing.T) {
		t.Run("successfully creates new order", func(t *testing.T) {
			// Arrange
			repo := NewStopOrderRepoInMemory()
			threshold, err1 := entities.NewContractPrice(20)
			require.NoError(t, err1)
			order := entities.NewStopOrder("FOO", entities.SideYes, threshold, nil, nil)

			// Act
			err := repo.Persist(order)

			// Assert
			require.NoError(t, err)

			foundOrder, err := repo.GetByID(order.ID())
			require.NoError(t, err)
			require.NotNil(t, foundOrder)
			assert.Equal(t, order.ID(), foundOrder.ID())
			assert.Equal(t, order.Ticker(), foundOrder.Ticker())
			assert.Equal(t, order.TriggerPrice(), foundOrder.TriggerPrice())
		})

		t.Run("successfully updates existing order", func(t *testing.T) {
			// Arrange
			repo := NewStopOrderRepoInMemory()
			threshold, err := entities.NewContractPrice(20)
			require.NoError(t, err)
			order := entities.NewStopOrder("FOO", entities.SideYes, threshold, nil, nil)
			require.NoError(t, repo.Persist(order), "Failed to persist initial order")

			// Act
			newThreshold, err := entities.NewContractPrice(30)
			require.NoError(t, err)
			order.SetTriggerPrice(newThreshold)
			persistErr := repo.Persist(order)

			// Assert
			require.NoError(t, persistErr)

			foundOrder, err := repo.GetByID(order.ID())
			require.NoError(t, err)
			require.NotNil(t, foundOrder)
			assert.Equal(t, order.ID(), foundOrder.ID())
			assert.Equal(t, newThreshold, foundOrder.TriggerPrice())
		})

		t.Run("updates preserve all fields", func(t *testing.T) {
			// Arrange
			repo := NewStopOrderRepoInMemory()
			threshold, err := entities.NewContractPrice(20)
			require.NoError(t, err)
			order := entities.NewStopOrder("FOO", entities.SideYes, threshold, nil, nil)
			require.NoError(t, repo.Persist(order), "Failed to persist initial order")

			// Act
			originalTicker := order.Ticker()
			newThreshold, err := entities.NewContractPrice(30)
			require.NoError(t, err)
			order.SetTriggerPrice(newThreshold)
			persistErr := repo.Persist(order)

			// Assert
			require.NoError(t, persistErr)

			foundOrder, err := repo.GetByID(order.ID())
			require.NoError(t, err)
			require.NotNil(t, foundOrder)
			assert.Equal(t, originalTicker, foundOrder.Ticker(), "Ticker should be preserved during update")
			assert.Equal(t, newThreshold, foundOrder.TriggerPrice(), "Threshold should be updated")
		})
	})
}
