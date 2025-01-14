package trigger_repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"prediction-risk/internal/app/contract"
	trigger_domain "prediction-risk/internal/app/risk/trigger/domain"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // postgres driver
)

var (
	ErrTriggerNotFound      = errors.New("trigger not found")
	ErrTriggerAlreadyExists = errors.New("active trigger already exists for this contract and type")
)

// Database models for scanning
type TriggerDB struct {
	TriggerID uuid.UUID `db:"trigger_id"`
	Type      string    `db:"trigger_type"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type PriceConditionDB struct {
	TriggerID      uuid.UUID `db:"trigger_id"`
	ContractTicker string    `db:"contract_ticker"`
	ContractSide   string    `db:"contract_side"`
	ThresholdPrice int       `db:"threshold_price"`
	Direction      string    `db:"direction"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

type TriggerActionDB struct {
	ActionID       uuid.UUID     `db:"action_id"`
	TriggerID      uuid.UUID     `db:"trigger_id"`
	ContractTicker string        `db:"contract_ticker"`
	ContractSide   string        `db:"contract_side"`
	OrderSide      string        `db:"order_side"`
	OrderSize      sql.NullInt64 `db:"order_size"`
	LimitPrice     sql.NullInt64 `db:"limit_price"`
	CreatedAt      time.Time     `db:"created_at"`
	UpdatedAt      time.Time     `db:"updated_at"`
}

type TriggerRepository struct {
	db *sqlx.DB
}

func NewTriggerRepository(db *sqlx.DB) *TriggerRepository {
	return &TriggerRepository{db: db}
}

// Save stores a new trigger in the database
func (r *TriggerRepository) Persist(ctx context.Context, trigger *trigger_domain.Trigger) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Upsert main trigger record
	triggerQuery := `
			INSERT INTO event_contract.trigger (
				trigger_id, trigger_type, status, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (trigger_id) DO UPDATE SET
				status = EXCLUDED.status,
				updated_at = EXCLUDED.updated_at
		`
	_, err = tx.ExecContext(ctx, triggerQuery,
		uuid.UUID(trigger.TriggerID),
		trigger_domain.TriggerTypeStop,
		trigger.Status,
		trigger.CreatedAt,
		trigger.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert trigger: %w", err)
	}

	// Upsert price condition
	conditionQuery := `
			INSERT INTO event_contract.price_trigger_condition (
				trigger_id, contract_ticker, contract_side,
				threshold_price, direction, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (trigger_id) DO UPDATE SET
				contract_ticker = EXCLUDED.contract_ticker,
				contract_side = EXCLUDED.contract_side,
				threshold_price = EXCLUDED.threshold_price,
				direction = EXCLUDED.direction,
				updated_at = EXCLUDED.updated_at
		`
	_, err = tx.ExecContext(ctx, conditionQuery,
		uuid.UUID(trigger.TriggerID),
		trigger.Condition.Contract.Ticker,
		trigger.Condition.Contract.Side,
		int(trigger.Condition.Price.Threshold),
		trigger.Condition.Price.Direction,
		trigger.CreatedAt,
		trigger.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert condition: %w", err)
	}

	// For actions, still need to delete and reinsert since they're a collection
	_, err = tx.ExecContext(ctx,
		"DELETE FROM event_contract.trigger_action WHERE trigger_id = $1",
		uuid.UUID(trigger.TriggerID),
	)
	if err != nil {
		return fmt.Errorf("delete actions: %w", err)
	}

	// Insert actions
	actionQuery := `
			INSERT INTO event_contract.trigger_action (
				trigger_id, contract_ticker, contract_side,
				order_side, order_size, limit_price,
				created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`
	for _, action := range trigger.Actions {
		var size *int64
		if action.Size != nil {
			i64 := int64(*action.Size)
			size = &i64
		}
		var limitPrice *int
		if action.LimitPrice != nil {
			intPrice := int(*action.LimitPrice)
			limitPrice = &intPrice
		}

		_, err = tx.ExecContext(ctx, actionQuery,
			uuid.UUID(trigger.TriggerID),
			action.Contract.Ticker,
			action.Contract.Side,
			action.Side,
			size,
			limitPrice,
			trigger.CreatedAt,
			trigger.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("insert action: %w", err)
		}
	}

	return tx.Commit()
}

// Get retrieves a trigger by its ID
func (r *TriggerRepository) Get(ctx context.Context, id trigger_domain.TriggerID) (*trigger_domain.Trigger, error) {
	// Get main trigger record
	var triggerDB TriggerDB
	err := r.db.GetContext(ctx, &triggerDB, `
		SELECT trigger_id, trigger_type, status, created_at, updated_at
		FROM event_contract.trigger
		WHERE trigger_id = $1
	`, uuid.UUID(id))
	if err == sql.ErrNoRows {
		return nil, ErrTriggerNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query trigger: %w", err)
	}

	// Get condition
	var conditionDB PriceConditionDB
	err = r.db.GetContext(ctx, &conditionDB, `
		SELECT contract_ticker, contract_side, threshold_price, direction
		FROM event_contract.price_trigger_condition
		WHERE trigger_id = $1
	`, uuid.UUID(id))
	if err != nil {
		return nil, fmt.Errorf("query condition: %w", err)
	}

	side, err := contract.NewSide(conditionDB.ContractSide)
	if err != nil {
		return nil, fmt.Errorf("create side: %w", err)
	}

	condition, err := trigger_domain.NewPriceCondition(
		contract.ContractIdentifier{
			Ticker: contract.Ticker(conditionDB.ContractTicker),
			Side:   side,
		},
		contract.ContractPrice(conditionDB.ThresholdPrice),
		trigger_domain.Direction(conditionDB.Direction),
	)
	if err != nil {
		return nil, fmt.Errorf("create price condition: %w", err)
	}

	// Get actions
	var actionsDB []TriggerActionDB
	err = r.db.SelectContext(ctx, &actionsDB, `
		SELECT contract_ticker, contract_side, order_side, order_size, limit_price
		FROM event_contract.trigger_action
		WHERE trigger_id = $1
		ORDER BY created_at
	`, uuid.UUID(id))
	if err != nil {
		return nil, fmt.Errorf("query actions: %w", err)
	}

	var actions []trigger_domain.TriggerAction
	for _, actionDB := range actionsDB {
		var size *uint
		if actionDB.OrderSize.Valid {
			uintSize := uint(actionDB.OrderSize.Int64)
			size = &uintSize
		}

		var limitPrice *contract.ContractPrice
		if actionDB.LimitPrice.Valid {
			price := contract.ContractPrice(actionDB.LimitPrice.Int64)
			limitPrice = &price
		}

		contractSide, err := contract.NewSide(conditionDB.ContractSide)
		if err != nil {
			return nil, fmt.Errorf("create side: %w", err)
		}

		orderSide, err := trigger_domain.NewOrderSide(actionDB.OrderSide)
		if err != nil {
			return nil, fmt.Errorf("create order side: %w", err)
		}

		action, err := trigger_domain.NewTriggerAction(
			contract.ContractIdentifier{
				Ticker: contract.Ticker(actionDB.ContractTicker),
				Side:   contractSide,
			},
			orderSide,
			size,
			limitPrice,
		)
		if err != nil {
			return nil, fmt.Errorf("create trigger action: %w", err)
		}
		actions = append(actions, *action)
	}

	status, err := trigger_domain.NewTriggerStatus(triggerDB.Status)
	if err != nil {
		return nil, fmt.Errorf("create trigger status: %w", err)
	}

	return &trigger_domain.Trigger{
		TriggerID: trigger_domain.TriggerID(triggerDB.TriggerID),
		Status:    status,
		Condition: *condition,
		Actions:   actions,
		CreatedAt: triggerDB.CreatedAt,
		UpdatedAt: triggerDB.UpdatedAt,
	}, nil
}

// GetAll retrieves all triggers
func (r *TriggerRepository) GetAll(ctx context.Context) ([]*trigger_domain.Trigger, error) {
	var triggerDBs []TriggerDB
	err := r.db.SelectContext(ctx, &triggerDBs, `
		SELECT DISTINCT t.trigger_id, t.trigger_type, t.status, t.created_at, t.updated_at
		FROM event_contract.trigger t
		JOIN event_contract.price_trigger_condition pc ON t.trigger_id = pc.trigger_id
	`)
	if err != nil {
		return nil, fmt.Errorf("query triggers: %w", err)
	}

	var triggers []*trigger_domain.Trigger
	for _, t := range triggerDBs {
		trigger, err := r.Get(ctx, trigger_domain.TriggerID(t.TriggerID))
		if err != nil {
			return nil, fmt.Errorf("get trigger %s: %w", t.TriggerID, err)
		}
		triggers = append(triggers, trigger)
	}

	return triggers, nil
}

// Helper method
func (r *TriggerRepository) checkExists(ctx context.Context, trigger *trigger_domain.Trigger) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
        SELECT EXISTS (
            SELECT 1 FROM event_contract.trigger t
            JOIN event_contract.price_trigger_condition pc ON t.trigger_id = pc.trigger_id
            WHERE t.trigger_id = $1
        )
    `, uuid.UUID(trigger.TriggerID)).Scan(&exists)
	return exists, err
}
