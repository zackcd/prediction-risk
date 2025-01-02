package postgres

import (
	"context"
	"fmt"
	"prediction-risk/internal/domain/entities"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type testDB struct {
	container testcontainers.Container
	db        *sqlx.DB
}

func setupTestDB(t *testing.T) *testDB {
	ctx := context.Background()

	// Container request
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections"),
		).WithDeadline(30 * time.Second),
	}

	// Start container
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err, "Failed to start container")

	// Get the mapped port
	mappedPort, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err, "Failed to get mapped port")

	hostIP, err := container.Host(ctx)
	require.NoError(t, err, "Failed to get host IP")

	t.Logf("Container Host IP: %s", hostIP)
	t.Logf("Mapped Port: %s", mappedPort.Port())

	// Connection string with mapped port
	connStr := fmt.Sprintf(
		"postgres://postgres:postgres@%s:%s/testdb?sslmode=disable",
		hostIP,
		mappedPort.Port(),
	)

	t.Logf("Connection string: %s", connStr)

	// Connect to the database
	db, err := sqlx.Open("postgres", connStr)
	require.NoError(t, err, "Failed to create database instance")

	maxAttempts := 5
	for i := 0; i < maxAttempts; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		t.Logf("Failed to ping database, attempt %d/%d: %v", i+1, maxAttempts, err)
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	require.NoError(t, err, "Failed to connect to database after retries")

	// Apply migrations
	_, err = db.Exec(`
		CREATE SCHEMA IF NOT EXISTS event_contract;

		CREATE DOMAIN event_contract.contract_price_cents AS INTEGER CHECK (
			value >= 0
			AND value <= 100
		);

		CREATE TYPE event_contract.order_status AS ENUM ('ACTIVE', 'TRIGGERED', 'CANCELLED', 'EXPIRED');
		CREATE TYPE event_contract.order_type AS ENUM ('STOP');
		CREATE TYPE event_contract.order_side AS ENUM ('YES', 'NO');

		CREATE TABLE event_contract.order (
			order_id UUID PRIMARY KEY,
			order_type event_contract.order_type NOT NULL,
			ticker VARCHAR NOT NULL,
			side event_contract.order_side NOT NULL,
			status event_contract.order_status NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE event_contract.stop_order (
			order_id UUID PRIMARY KEY REFERENCES event_contract.order (order_id) ON DELETE CASCADE,
			trigger_price event_contract.contract_price_cents,
			limit_price event_contract.contract_price_cents
		);

		CREATE UNIQUE INDEX idx_unique_active_stop_order ON event_contract.order (ticker, side)
		WHERE
			status = 'ACTIVE'
			AND order_type = 'STOP';
	`)
	require.NoError(t, err)

	return &testDB{
		container: container,
		db:        db,
	}
}

func (tdb *testDB) Close(t *testing.T) {
	ctx := context.Background()
	require.NoError(t, tdb.db.Close())
	require.NoError(t, tdb.container.Terminate(ctx))
}

func createTestStopOrder(t *testing.T, ticker string) *entities.StopOrder {
	side, err := entities.NewSide("YES")
	require.NoError(t, err)

	triggerPrice, err := entities.NewContractPrice(50)
	require.NoError(t, err)

	limitPrice, err := entities.NewContractPrice(45)
	require.NoError(t, err)

	return entities.NewStopOrder(
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

	order, err := repo.GetByID(entities.NewOrderID())
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
	orderMap := make(map[entities.OrderID]*entities.StopOrder)
	for _, order := range orders {
		orderMap[order.ID()] = order
	}

	// Assert both orders were retrieved correctly
	for _, original := range []*entities.StopOrder{order1, order2} {
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

// func TestStopOrderRepo_Persist_Update(t *testing.T) {
// 	testdb := setupTestDB(t)
// 	defer testdb.Close(t)

// 	repo := NewStopOrderRepoPostgres(testdb.db)

// 	// Create and persist initial order
// 	stopOrder := createTestStopOrder(t, "FOO")
// 	require.NoError(t, repo.Persist(stopOrder))

// 	// Update status
// 	status, err := entities.ParseOrderStatus("TRIGGERED")
// 	require.NoError(t, err)
// 	require.NoError(t, stopOrder.UpdateStatus(status))

// 	// Persist update
// 	require.NoError(t, repo.Persist(stopOrder))

// 	// Retrieve and verify update
// 	retrieved, err := repo.GetByID(stopOrder.ID())
// 	require.NoError(t, err)
// 	require.NotNil(t, retrieved)
// 	assert.Equal(t, entities.OrderStatusTriggered, retrieved.Status())
// }

// func TestStopOrderRepo_Persist_NullLimitPrice(t *testing.T) {
// 	testdb := setupTestDB(t)
// 	defer testdb.Close(t)

// 	repo := NewStopOrderRepoPostgres(testdb.db)

// 	side, err := entities.NewSide("YES")
// 	require.NoError(t, err)

// 	triggerPrice, err := entities.NewContractPrice(50)
// 	require.NoError(t, err)

// 	// Create order with nil limit price
// 	stopOrder := entities.NewStopOrder(
// 		"TEST-2024",
// 		side,
// 		triggerPrice,
// 		nil,
// 	)

// 	// Persist and retrieve
// 	require.NoError(t, repo.Persist(stopOrder))

// 	retrieved, err := repo.GetByID(stopOrder.ID())
// 	require.NoError(t, err)
// 	require.NotNil(t, retrieved)
// 	assert.Nil(t, retrieved.LimitPrice())
// }

// func TestStopOrderRepo_UniqueConstraint(t *testing.T) {
// 	testdb := setupTestDB(t)
// 	defer testdb.Close(t)

// 	repo := NewStopOrderRepoPostgres(testdb.db)
// 	ticker := "FOO"

// 	// Create first order
// 	order1 := createTestStopOrder(t, ticker)
// 	require.NoError(t, repo.Persist(order1))

// 	// Create second order with same ticker and side
// 	order2 := createTestStopOrder(t, ticker)
// 	err := repo.Persist(order2)
// 	assert.Error(t, err) // Should fail due to unique constraint
// }
