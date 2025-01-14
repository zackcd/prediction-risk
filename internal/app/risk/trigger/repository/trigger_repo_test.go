package trigger_repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"prediction-risk/internal/app/contract"
	trigger_domain "prediction-risk/internal/app/risk/trigger/domain"
	"prediction-risk/internal/app/testutil"
)

func createTestTrigger() *trigger_domain.Trigger {
	contractID := contract.ContractIdentifier{
		Ticker: "FOO",
		Side:   contract.SideYes,
	}

	condition, err := trigger_domain.NewPriceCondition(
		contractID,
		contract.ContractPrice(50),
		trigger_domain.Below,
	)
	if err != nil {
		panic(err)
	}

	action, err := trigger_domain.NewTriggerAction(
		contractID,
		trigger_domain.Sell,
		nil,
		nil,
	)
	if err != nil {
		panic(err)
	}

	now := time.Now()

	trigger := &trigger_domain.Trigger{
		TriggerID: trigger_domain.NewTriggerID(),
		Status:    trigger_domain.StatusActive,
		Condition: *condition,
		Actions:   []trigger_domain.TriggerAction{*action},
		CreatedAt: now,
		UpdatedAt: now,
	}

	return trigger
}

func TestTriggerRepository_Persist(t *testing.T) {
	testDB := testutil.SetupTestDB(t)
	defer testDB.Close(t)

	repo := NewTriggerRepository(testDB.DB())

	t.Run("creates new trigger", func(t *testing.T) {
		defer testDB.Cleanup(t)

		trigger := createTestTrigger()
		err := repo.Persist(context.Background(), trigger)
		require.NoError(t, err)

		// Verify save by retrieving
		saved, err := repo.Get(context.Background(), trigger.TriggerID)
		require.NoError(t, err)
		assert.Equal(t, trigger.TriggerID, saved.TriggerID)
		assert.Equal(t, trigger.Status, saved.Status)
		assert.Equal(t, trigger.Condition.Contract, saved.Condition.Contract)
		assert.Equal(t, trigger.Condition.Price.Threshold, saved.Condition.Price.Threshold)
		assert.Equal(t, trigger.Condition.Price.Direction, saved.Condition.Price.Direction)
		assert.Len(t, saved.Actions, 1)
		assert.Equal(t, trigger.Actions[0].Contract, saved.Actions[0].Contract)
		assert.Equal(t, trigger.Actions[0].Side, saved.Actions[0].Side)
	})

	t.Run("updates existing trigger", func(t *testing.T) {
		defer testDB.Cleanup(t)

		// First create a trigger
		trigger := createTestTrigger()
		err := repo.Persist(context.Background(), trigger)
		require.NoError(t, err)

		// Modify the trigger
		trigger.Status = trigger_domain.StatusCancelled
		newAction, _ := trigger_domain.NewTriggerAction(
			contract.ContractIdentifier{
				Ticker: "BAR",
				Side:   contract.SideYes,
			},
			trigger_domain.Sell,
			nil,
			nil,
		)
		trigger.Actions = append(trigger.Actions, *newAction)
		trigger.UpdatedAt = time.Now()

		// Persist the changes
		err = repo.Persist(context.Background(), trigger)
		require.NoError(t, err)

		// Verify updates
		updated, err := repo.Get(context.Background(), trigger.TriggerID)
		require.NoError(t, err)
		assert.Equal(t, trigger_domain.StatusCancelled, updated.Status)
		assert.Len(t, updated.Actions, 2)
		assert.Equal(t, "BAR", string(updated.Actions[1].Contract.Ticker))
	})
}

func TestTriggerRepository_Get(t *testing.T) {
	testDB := testutil.SetupTestDB(t)
	defer testDB.Close(t)

	repo := NewTriggerRepository(testDB.DB())

	t.Run("get existing trigger", func(t *testing.T) {
		defer testDB.Cleanup(t)

		trigger := createTestTrigger()
		err := repo.Persist(context.Background(), trigger)
		require.NoError(t, err)

		found, err := repo.Get(context.Background(), trigger.TriggerID)
		require.NoError(t, err)
		assert.Equal(t, trigger.TriggerID, found.TriggerID)
	})

	t.Run("trigger not found", func(t *testing.T) {
		defer testDB.Cleanup(t)

		_, err := repo.Get(context.Background(), trigger_domain.NewTriggerID())
		assert.ErrorIs(t, err, ErrTriggerNotFound)
	})
}

func TestTriggerRepository_GetAll(t *testing.T) {
	testDB := testutil.SetupTestDB(t)
	defer testDB.Close(t)

	repo := NewTriggerRepository(testDB.DB())

	t.Run("get all triggers", func(t *testing.T) {
		defer testDB.Cleanup(t)

		// Create multiple triggers
		trigger1 := createTestTrigger()
		err := repo.Persist(context.Background(), trigger1)
		require.NoError(t, err)

		trigger2 := createTestTrigger()
		trigger2.Condition.Contract.Ticker = "ETH-USD"
		err = repo.Persist(context.Background(), trigger2)
		require.NoError(t, err)

		// Get all triggers
		triggers, err := repo.GetAll(context.Background())
		require.NoError(t, err)
		assert.Len(t, triggers, 2)

		// Verify trigger details
		triggerIDs := []trigger_domain.TriggerID{triggers[0].TriggerID, triggers[1].TriggerID}
		assert.Contains(t, triggerIDs, trigger1.TriggerID)
		assert.Contains(t, triggerIDs, trigger2.TriggerID)
	})

	t.Run("get all with no triggers", func(t *testing.T) {
		defer testDB.Cleanup(t)

		triggers, err := repo.GetAll(context.Background())
		require.NoError(t, err)
		assert.Empty(t, triggers)
	})
}
