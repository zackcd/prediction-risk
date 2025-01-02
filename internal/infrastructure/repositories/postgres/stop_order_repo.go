package postgres

import (
	"database/sql"
	"prediction-risk/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type StopOrderRepoPostgres struct {
	db *sqlx.DB
}

func NewStopOrderRepoPostgres(db *sqlx.DB) *StopOrderRepoPostgres {
	return &StopOrderRepoPostgres{db: db}
}

func (r *StopOrderRepoPostgres) GetByID(id entities.OrderID) (*entities.StopOrder, error) {
	const query = `
		SELECT
			o.order_id,
			o.ticker,
			o.side,
			o.status,
			o.created_at,
			o.updated_at,
			so.trigger_price,
			so.limit_price
		FROM event_contract.order o
		JOIN event_contract.stop_order so ON o.order_id = so.order_id
		WHERE o.order_id = $1 AND o.order_type = 'STOP'
	`

	var result struct {
		OrderID      uuid.UUID     `db:"order_id"`
		Ticker       string        `db:"ticker"`
		Side         string        `db:"side"`
		Status       string        `db:"status"`
		CreatedAt    string        `db:"created_at"`
		UpdatedAt    string        `db:"updated_at"`
		TriggerPrice int           `db:"trigger_price"`
		LimitPrice   sql.NullInt64 `db:"limit_price"`
	}

	if err := r.db.Get(&result, query, id.String()); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	side, err := entities.NewSide(result.Side)
	if err != nil {
		return nil, err
	}

	// Not checking error because database value is guaranteed to be valid
	triggerPrice, _ := entities.NewContractPrice(result.TriggerPrice)
	var limitPrice *entities.ContractPrice
	if result.LimitPrice.Valid {
		price, _ := entities.NewContractPrice(int(result.LimitPrice.Int64))
		limitPrice = &price
	}

	stopOrder := entities.NewStopOrder(
		result.Ticker,
		side,
		triggerPrice,
		limitPrice,
		&id,
	)

	status, err := entities.ParseOrderStatus(result.Status)
	if err != nil {
		return nil, err
	}

	if err := stopOrder.UpdateStatus(status); err != nil {
		return nil, err
	}

	return stopOrder, nil
}

func (r *StopOrderRepoPostgres) GetAll() ([]*entities.StopOrder, error) {
	const query = `
		SELECT
			o.order_id,
			o.ticker,
			o.side,
			o.status,
			o.created_at,
			o.updated_at,
			so.trigger_price,
			so.limit_price
		FROM event_contract.order o
		JOIN event_contract.stop_order so ON o.order_id = so.order_id
		WHERE o.order_type = 'STOP'
	`

	var results []struct {
		OrderID      uuid.UUID     `db:"order_id"`
		Ticker       string        `db:"ticker"`
		Side         string        `db:"side"`
		Status       string        `db:"status"`
		CreatedAt    string        `db:"created_at"`
		UpdatedAt    string        `db:"updated_at"`
		TriggerPrice int           `db:"trigger_price"`
		LimitPrice   sql.NullInt64 `db:"limit_price"`
	}

	if err := r.db.Select(&results, query); err != nil {
		return nil, err
	}

	stopOrders := make([]*entities.StopOrder, 0, len(results))
	for _, result := range results {
		side, err := entities.NewSide(result.Side)
		if err != nil {
			return nil, err
		}

		triggerPrice, _ := entities.NewContractPrice(result.TriggerPrice)
		var limitPrice *entities.ContractPrice
		if result.LimitPrice.Valid {
			price, _ := entities.NewContractPrice(int(result.LimitPrice.Int64))
			limitPrice = &price
		}

		id := entities.OrderID(result.OrderID)

		stopOrder := entities.NewStopOrder(
			result.Ticker,
			side,
			triggerPrice,
			limitPrice,
			&id,
		)

		status, err := entities.ParseOrderStatus(result.Status)
		if err != nil {
			return nil, err
		}

		if err := stopOrder.UpdateStatus(status); err != nil {
			return nil, err
		}

		stopOrders = append(stopOrders, stopOrder)
	}

	return stopOrders, nil
}

func (r *StopOrderRepoPostgres) Persist(stopOrder *entities.StopOrder) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert/update order
	orderQuery := `
		INSERT INTO event_contract.order (
			order_id,
			order_type,
			ticker,
			side,
			status,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (order_id) DO UPDATE
		SET
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`

	_, err = tx.Exec(
		orderQuery,
		stopOrder.ID().String(),
		stopOrder.OrderType(),
		stopOrder.Ticker(),
		stopOrder.Side(),
		stopOrder.Status(),
		stopOrder.CreatedAt(),
		stopOrder.UpdatedAt(),
	)
	if err != nil {
		return err
	}

	// Insert/update stop order details
	stopOrderQuery := `
		INSERT INTO event_contract.stop_order (
			order_id,
			trigger_price,
			limit_price
		)
		VALUES ($1, $2, $3)
		ON CONFLICT (order_id) DO UPDATE
		SET
			trigger_price = EXCLUDED.trigger_price,
			limit_price = EXCLUDED.limit_price
	`

	var limitPrice sql.NullInt64
	if stopOrder.LimitPrice() != nil {
		limitPrice = sql.NullInt64{
			Int64: int64((*stopOrder.LimitPrice()).Value()),
			Valid: true,
		}
	}

	_, err = tx.Exec(
		stopOrderQuery,
		stopOrder.ID().String(),
		stopOrder.TriggerPrice().Value(),
		limitPrice,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}
