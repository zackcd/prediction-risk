package trigger_repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"prediction-risk/internal/app/contract"
	trigger_domain "prediction-risk/internal/app/risk/trigger/domain"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var (
	ErrTriggerNotFound      = errors.New("trigger not found")
	ErrTriggerAlreadyExists = errors.New("active trigger already exists for this contract and type")
)

type TriggerRepository struct {
	db *sqlx.DB
}

func NewTriggerRepository(db *sqlx.DB) *TriggerRepository {
	return &TriggerRepository{db: db}
}

// Save stores a new trigger in the database
func (r *TriggerRepository) Save(ctx context.Context, trigger *trigger_domain.Trigger) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert main trigger record
	triggerQuery := `
		INSERT INTO triggers (
			trigger_id, trigger_type, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5)
	`
	_, err = tx.ExecContext(ctx, triggerQuery,
		trigger.TriggerID,
		trigger_domain.TriggerTypeStop,
		trigger.Status,
		trigger.CreatedAt,
		trigger.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert trigger: %w", err)
	}

	// Insert price condition
	conditionQuery := `
		INSERT INTO price_trigger_conditions (
			trigger_id, contract_ticker, contract_side,
			threshold, direction, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = tx.ExecContext(ctx, conditionQuery,
		trigger.TriggerID,
		trigger.Condition.Contract.Ticker,
		trigger.Condition.Contract.Side,
		trigger.Condition.Price.Threshold,
		trigger.Condition.Price.Direction,
		trigger.CreatedAt,
		trigger.UpdatedAt,
	)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" && pqErr.Constraint == "idx_unique_active_trigger" {
				return ErrTriggerAlreadyExists
			}
		}
		return fmt.Errorf("insert price condition: %w", err)
	}

	// Insert actions
	actionQuery := `
		INSERT INTO trigger_actions (
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

		_, err = tx.Exec(actionQuery,
			trigger.TriggerID,
			action.Contract.Ticker,
			action.Contract.Side,
			action.Side,
			size,
			action.LimitPrice,
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
	// First get the main trigger record
	triggerQuery := `
		SELECT trigger_id, trigger_type, status, created_at, updated_at
		FROM triggers
		WHERE trigger_id = $1
	`
	var trigger trigger_domain.Trigger
	var triggerType string

	err := r.db.QueryRowContext(ctx, triggerQuery, id).Scan(
		&trigger.TriggerID,
		&triggerType,
		&trigger.Status,
		&trigger.CreatedAt,
		&trigger.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrTriggerNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query trigger: %w", err)
	}

	// Get the condition
	conditionQuery := `
		SELECT contract_ticker, contract_side, threshold, direction
		FROM price_trigger_conditions
		WHERE trigger_id = $1
	`
	var (
		contractTicker string
		contractSide   string
		threshold      float64
		direction      string
	)

	err = r.db.QueryRowContext(ctx, conditionQuery, id).Scan(
		&contractTicker,
		&contractSide,
		&threshold,
		&direction,
	)
	if err != nil {
		return nil, fmt.Errorf("query condition: %w", err)
	}

	side, err := contract.NewSide(contractSide)
	if err != nil {
		return nil, fmt.Errorf("create side: %w", err)
	}
	// Create condition
	condition, err := trigger_domain.NewPriceCondition(
		contract.ContractIdentifier{
			Ticker: contract.Ticker(contractTicker),
			Side:   side,
		},
		contract.ContractPrice(threshold),
		trigger_domain.Direction(direction),
	)
	if err != nil {
		return nil, fmt.Errorf("create price condition: %w", err)
	}
	trigger.Condition = *condition

	// Get all actions
	actionsQuery := `
		SELECT contract_ticker, contract_side, order_side, order_size, limit_price
		FROM trigger_actions
		WHERE trigger_id = $1
		ORDER BY created_at
	`
	rows, err := r.db.QueryContext(ctx, actionsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("query actions: %w", err)
	}
	defer rows.Close()

	var actions []trigger_domain.TriggerAction
	for rows.Next() {
		var (
			actionContractTicker string
			actionContractSide   string
			orderSide            string
			orderSize            sql.NullInt64
			limitPrice           sql.NullFloat64
		)

		err := rows.Scan(
			&actionContractTicker,
			&actionContractSide,
			&orderSide,
			&orderSize,
			&limitPrice,
		)
		if err != nil {
			return nil, fmt.Errorf("scan action: %w", err)
		}

		var size *uint
		if orderSize.Valid {
			uintSize := uint(orderSize.Int64)
			size = &uintSize
		}

		var price *contract.ContractPrice
		if limitPrice.Valid {
			p := contract.ContractPrice(limitPrice.Float64)
			price = &p
		}

		side, err := contract.NewSide(contractSide)
		if err != nil {
			return nil, fmt.Errorf("create side: %w", err)
		}
		action, err := trigger_domain.NewTriggerAction(
			contract.ContractIdentifier{
				Ticker: contract.Ticker(actionContractTicker),
				Side:   side,
			},
			trigger_domain.OrderSide(orderSide),
			size,
			price,
		)
		if err != nil {
			return nil, fmt.Errorf("create trigger action: %w", err)
		}

		actions = append(actions, *action)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("action rows: %w", err)
	}

	trigger.Actions = actions
	return &trigger, nil
}

// Update updates an existing trigger
func (r *TriggerRepository) Update(ctx context.Context, trigger *trigger_domain.Trigger) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update main trigger
	triggerQuery := `
		UPDATE triggers
		SET status = $2,
			updated_at = $3
		WHERE trigger_id = $1
	`
	result, err := tx.ExecContext(ctx, triggerQuery,
		trigger.TriggerID,
		trigger.Status,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("update trigger: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrTriggerNotFound
	}

	// Update condition
	conditionQuery := `
		UPDATE price_trigger_conditions
		SET contract_ticker = $2,
			contract_side = $3,
			threshold = $4,
			direction = $5,
			updated_at = $6
		WHERE trigger_id = $1
	`
	_, err = tx.ExecContext(ctx, conditionQuery,
		trigger.TriggerID,
		trigger.Condition.Contract.Ticker,
		trigger.Condition.Contract.Side,
		trigger.Condition.Price.Threshold,
		trigger.Condition.Price.Direction,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("update condition: %w", err)
	}

	// Delete existing actions and insert new ones
	_, err = tx.ExecContext(ctx, "DELETE FROM trigger_actions WHERE trigger_id = $1", trigger.TriggerID)
	if err != nil {
		return fmt.Errorf("delete actions: %w", err)
	}

	actionQuery := `
		INSERT INTO trigger_actions (
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

		_, err = tx.ExecContext(ctx, actionQuery,
			trigger.TriggerID,
			action.Contract.Ticker,
			action.Contract.Side,
			action.Side,
			size,
			action.LimitPrice,
			time.Now(),
			time.Now(),
		)
		if err != nil {
			return fmt.Errorf("insert action: %w", err)
		}
	}

	return tx.Commit()
}

// GetActiveByContract retrieves all active triggers for a specific contract
func (r *TriggerRepository) GetAll(ctx context.Context) ([]*trigger_domain.Trigger, error) {
	query := `
		SELECT DISTINCT t.trigger_id
		FROM triggers t
		JOIN price_trigger_conditions pc ON t.trigger_id = pc.trigger_id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query triggers: %w", err)
	}
	defer rows.Close()

	var triggers []*trigger_domain.Trigger
	for rows.Next() {
		var triggerID trigger_domain.TriggerID
		if err := rows.Scan(&triggerID); err != nil {
			return nil, fmt.Errorf("scan trigger id: %w", err)
		}

		trigger, err := r.Get(ctx, triggerID)
		if err != nil {
			return nil, fmt.Errorf("get trigger %s: %w", triggerID, err)
		}

		triggers = append(triggers, trigger)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("trigger rows: %w", err)
	}

	return triggers, nil
}
