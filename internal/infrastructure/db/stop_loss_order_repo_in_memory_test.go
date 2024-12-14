package repositories

import (
	"prediction-risk/internal/domain/entities"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStopLossRepoInMemory(t *testing.T) {
	t.Run("GetByID", func(t *testing.T) {
		t.Run("returns order when found", func(t *testing.T) {
			// Arrange
			repo := NewStopLossRepoInMemory()
			order := entities.NewStopLossOrder("FOO", entities.SideYes, entities.NewContractPrice(20))
			require.NoError(t, repo.Persist(*order), "Failed to persist order")

			// Act
			foundOrder, err := repo.GetByID(order.ID())

			// Assert
			require.NoError(t, err)
			require.NotNil(t, foundOrder)
			assert.Equal(t, order.ID(), foundOrder.ID())
			assert.Equal(t, order.Ticker(), foundOrder.Ticker())
			assert.Equal(t, order.Threshold(), foundOrder.Threshold())
		})

		t.Run("returns nil when not found", func(t *testing.T) {
			// Arrange
			repo := NewStopLossRepoInMemory()
			id := uuid.New()

			// Act
			order, err := repo.GetByID(id)

			// Assert
			require.NoError(t, err)
			assert.Nil(t, order)
		})
	})

	t.Run("Persist", func(t *testing.T) {
		t.Run("successfully creates new order", func(t *testing.T) {
			// Arrange
			repo := NewStopLossRepoInMemory()
			order := entities.NewStopLossOrder("FOO", entities.SideYes, entities.NewContractPrice(20))

			// Act
			err := repo.Persist(*order)

			// Assert
			require.NoError(t, err)

			foundOrder, err := repo.GetByID(order.ID())
			require.NoError(t, err)
			require.NotNil(t, foundOrder)
			assert.Equal(t, order.ID(), foundOrder.ID())
			assert.Equal(t, order.Ticker(), foundOrder.Ticker())
			assert.Equal(t, order.Threshold(), foundOrder.Threshold())
		})

		t.Run("successfully updates existing order", func(t *testing.T) {
			// Arrange
			repo := NewStopLossRepoInMemory()
			order := entities.NewStopLossOrder("FOO", entities.SideYes, entities.NewContractPrice(20))
			require.NoError(t, repo.Persist(*order), "Failed to persist initial order")

			// Act
			order.SetThreshold(entities.NewContractPrice(30))
			err := repo.Persist(*order)

			// Assert
			require.NoError(t, err)

			foundOrder, err := repo.GetByID(order.ID())
			require.NoError(t, err)
			require.NotNil(t, foundOrder)
			assert.Equal(t, order.ID(), foundOrder.ID())
			assert.Equal(t, entities.NewContractPrice(30), foundOrder.Threshold())
		})

		t.Run("updates preserve all fields", func(t *testing.T) {
			// Arrange
			repo := NewStopLossRepoInMemory()
			order := entities.NewStopLossOrder("FOO", entities.SideYes, entities.NewContractPrice(20))
			require.NoError(t, repo.Persist(*order), "Failed to persist initial order")

			// Act
			originalTicker := order.Ticker()
			order.SetThreshold(entities.NewContractPrice(30))
			err := repo.Persist(*order)

			// Assert
			require.NoError(t, err)

			foundOrder, err := repo.GetByID(order.ID())
			require.NoError(t, err)
			require.NotNil(t, foundOrder)
			assert.Equal(t, originalTicker, foundOrder.Ticker(), "Ticker should be preserved during update")
			assert.Equal(t, entities.NewContractPrice(30), foundOrder.Threshold(), "Threshold should be updated")
		})
	})
}
