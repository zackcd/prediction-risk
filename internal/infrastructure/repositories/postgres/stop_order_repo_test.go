package postgres

import (
	"prediction-risk/internal/domain/contract"
	"prediction-risk/internal/domain/order"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestStopOrder(t *testing.T, ticker string) *order.StopOrder {
	side, err := contract.NewSide("YES")
	require.NoError(t, err)

	triggerPrice, err := contract.NewContractPrice(50)
	require.NoError(t, err)

	limitPrice, err := contract.NewContractPrice(45)
	require.NoError(t, err)

	return order.NewStopOrder(
		ticker,
		side,
		triggerPrice,
		&limitPrice,
		nil,
	)
}

func TestStopOrderRepo_GetByID_NotFound(t *testing.T) {
	testdb := setupTestDB(t)
	defer testdb.Close(t)

	repo := NewStopOrderRepoPostgres(testdb.db)

	order, err := repo.GetByID(order.NewOrderID())
	require.NoError(t, err)
	assert.Nil(t, order)
}

func TestStopOrderRepo_Persist_And_GetByID(t *testing.T) {
	testdb := setupTestDB(t)
	defer testdb.Close(t)

	repo := NewStopOrderRepoPostgres(testdb.db)

	stopOrder := createTestStopOrder(t, "TEST-2024")
	originalID := stopOrder.ID()
	t.Logf("Original Order ID: %s", originalID)

	// Persist the order
	err := repo.Persist(stopOrder)
	require.NoError(t, err)

	// Double check the ID hasn't changed after persist
	t.Logf("Order ID after persist: %s", stopOrder.ID())
	assert.Equal(t, originalID, stopOrder.ID(), "ID should not change after persist")

	// Now retrieve it
	retrieved, err := repo.GetByID(originalID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	t.Logf("Retrieved Order ID: %s", retrieved.ID())

	// Assert equality
	assert.Equal(t, originalID, retrieved.ID(), "Retrieved ID should match original")
	assert.Equal(t, stopOrder.Ticker(), retrieved.Ticker())
	assert.Equal(t, stopOrder.Side(), retrieved.Side())
	assert.Equal(t, stopOrder.Status(), retrieved.Status())
	assert.Equal(t, stopOrder.TriggerPrice().Value(), retrieved.TriggerPrice().Value())
	assert.NotNil(t, retrieved.LimitPrice())
	assert.Equal(t, (*stopOrder.LimitPrice()).Value(), (*retrieved.LimitPrice()).Value())
}

func TestStopOrderRepo_GetAll_Empty(t *testing.T) {
	testdb := setupTestDB(t)
	defer testdb.Close(t)

	repo := NewStopOrderRepoPostgres(testdb.db)

	orders, err := repo.GetAll()
	require.NoError(t, err)
	assert.Empty(t, orders)
}

func TestStopOrderRepo_GetAll_MultipleOrders(t *testing.T) {
	testdb := setupTestDB(t)
	defer testdb.Close(t)

	repo := NewStopOrderRepoPostgres(testdb.db)

	// Create and persist multiple orders
	order1 := createTestStopOrder(t, "FOO-1")
	order2 := createTestStopOrder(t, "FOO-2")

	require.NoError(t, repo.Persist(order1))
	require.NoError(t, repo.Persist(order2))

	// Retrieve all
	orders, err := repo.GetAll()
	require.NoError(t, err)
	assert.Len(t, orders, 2)

	// Map orders by ID for easier comparison
	orderMap := make(map[order.OrderID]*order.StopOrder)
	for _, order := range orders {
		orderMap[order.ID()] = order
	}

	// Assert both orders were retrieved correctly
	for _, original := range []*order.StopOrder{order1, order2} {
		retrieved, exists := orderMap[original.ID()]
		assert.True(t, exists)
		assert.Equal(t, original.Ticker(), retrieved.Ticker())
		assert.Equal(t, original.Side(), retrieved.Side())
		assert.Equal(t, original.Status(), retrieved.Status())
		assert.Equal(t, original.TriggerPrice().Value(), retrieved.TriggerPrice().Value())
		assert.NotNil(t, retrieved.LimitPrice())
		assert.Equal(t, (*original.LimitPrice()).Value(), (*retrieved.LimitPrice()).Value())
	}
}

func TestStopOrderRepo_Persist_Update(t *testing.T) {
	testdb := setupTestDB(t)
	defer testdb.Close(t)

	repo := NewStopOrderRepoPostgres(testdb.db)

	// Create and persist initial order
	stopOrder := createTestStopOrder(t, "FOO")
	require.NoError(t, repo.Persist(stopOrder))

	// Update status
	status := order.OrderStatusTriggered
	require.NoError(t, stopOrder.UpdateStatus(status))

	// Persist update
	require.NoError(t, repo.Persist(stopOrder))

	// Retrieve and verify update
	retrieved, err := repo.GetByID(stopOrder.ID())
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, order.OrderStatusTriggered, retrieved.Status())
}

func TestStopOrderRepo_Persist_NullLimitPrice(t *testing.T) {
	testdb := setupTestDB(t)
	defer testdb.Close(t)

	repo := NewStopOrderRepoPostgres(testdb.db)

	side, err := contract.NewSide("YES")
	require.NoError(t, err)

	triggerPrice, err := contract.NewContractPrice(50)
	require.NoError(t, err)

	// Create order with nil limit price
	stopOrder := order.NewStopOrder(
		"TEST-2024",
		side,
		triggerPrice,
		nil,
		nil,
	)

	// Persist and retrieve
	require.NoError(t, repo.Persist(stopOrder))

	retrieved, err := repo.GetByID(stopOrder.ID())
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Nil(t, retrieved.LimitPrice())
}

func TestStopOrderRepo_UniqueConstraint(t *testing.T) {
	testdb := setupTestDB(t)
	defer testdb.Close(t)

	repo := NewStopOrderRepoPostgres(testdb.db)
	ticker := "FOO"

	// Create first order
	order1 := createTestStopOrder(t, ticker)
	require.NoError(t, repo.Persist(order1))

	// Create second order with same ticker and side
	order2 := createTestStopOrder(t, ticker)
	err := repo.Persist(order2)
	assert.Error(t, err) // Should fail due to unique constraint
}
