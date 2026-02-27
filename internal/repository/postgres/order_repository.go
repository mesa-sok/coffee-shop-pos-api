package postgres

import (
	"context"
	"database/sql"
	"time"

	"coffee-shop-pos/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

type orderRepository struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) domain.OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(ctx context.Context, order *domain.Order) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	orderQuery := `INSERT INTO orders (id, order_number, status, subtotal, tax, total, created_at, updated_at)
		VALUES (:id, :order_number, :status, :subtotal, :tax, :total, :created_at, :updated_at)`
	if _, err := tx.NamedExecContext(ctx, orderQuery, order); err != nil {
		return err
	}

	itemQuery := `INSERT INTO order_items (id, order_id, menu_item_id, quantity, unit_price, line_total)
		VALUES (:id, :order_id, :menu_item_id, :quantity, :unit_price, :line_total)`
	for i := range order.Items {
		if _, err := tx.NamedExecContext(ctx, itemQuery, &order.Items[i]); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *orderRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	query := `SELECT o.id, o.order_number, o.status, o.subtotal, o.tax, o.total, o.created_at, o.updated_at,
		oi.id AS item_id, oi.order_id, oi.menu_item_id, oi.quantity, oi.unit_price, oi.line_total
		FROM orders o
		LEFT JOIN order_items oi ON oi.order_id = o.id
		WHERE o.id = $1
		ORDER BY oi.id`

	type orderJoinRow struct {
		ID          uuid.UUID        `db:"id"`
		OrderNumber string           `db:"order_number"`
		Status      string           `db:"status"`
		Subtotal    decimal.Decimal  `db:"subtotal"`
		Tax         decimal.Decimal  `db:"tax"`
		Total       decimal.Decimal  `db:"total"`
		CreatedAt   time.Time        `db:"created_at"`
		UpdatedAt   time.Time        `db:"updated_at"`
		ItemID      *uuid.UUID       `db:"item_id"`
		OrderID     *uuid.UUID       `db:"order_id"`
		MenuItemID  *uuid.UUID       `db:"menu_item_id"`
		Quantity    *int             `db:"quantity"`
		UnitPrice   *decimal.Decimal `db:"unit_price"`
		LineTotal   *decimal.Decimal `db:"line_total"`
	}

	var rows []orderJoinRow
	if err := r.db.SelectContext(ctx, &rows, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}

	order := &domain.Order{
		ID:          rows[0].ID,
		OrderNumber: rows[0].OrderNumber,
		Status:      rows[0].Status,
		Subtotal:    rows[0].Subtotal,
		Tax:         rows[0].Tax,
		Total:       rows[0].Total,
		CreatedAt:   rows[0].CreatedAt,
		UpdatedAt:   rows[0].UpdatedAt,
		Items:       []domain.OrderItem{},
	}

	for _, row := range rows {
		if row.ItemID == nil {
			continue
		}
		item := domain.OrderItem{
			ID:         *row.ItemID,
			OrderID:    *row.OrderID,
			MenuItemID: *row.MenuItemID,
			Quantity:   *row.Quantity,
			UnitPrice:  *row.UnitPrice,
			LineTotal:  *row.LineTotal,
		}
		order.Items = append(order.Items, item)
	}

	return order, nil
}

func (r *orderRepository) List(ctx context.Context) ([]domain.Order, error) {
	query := `SELECT id, order_number, status, subtotal, tax, total, created_at, updated_at FROM orders ORDER BY created_at DESC`
	var orders []domain.Order
	if err := r.db.SelectContext(ctx, &orders, query); err != nil {
		return nil, err
	}

	if len(orders) == 0 {
		return orders, nil
	}

	orderIDs := make([]uuid.UUID, 0, len(orders))
	for _, order := range orders {
		orderIDs = append(orderIDs, order.ID)
	}

	itemsByOrder, err := r.getOrderItems(ctx, orderIDs)
	if err != nil {
		return nil, err
	}

	for i := range orders {
		orders[i].Items = itemsByOrder[orders[i].ID]
	}

	return orders, nil
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, updatedAt time.Time) error {
	query := `UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3`
	result, err := r.db.ExecContext(ctx, query, status, updatedAt, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *orderRepository) getOrderItems(ctx context.Context, orderIDs []uuid.UUID) (map[uuid.UUID][]domain.OrderItem, error) {
	itemsByOrder := make(map[uuid.UUID][]domain.OrderItem)
	query, args, err := sqlx.In(`SELECT id, order_id, menu_item_id, quantity, unit_price, line_total
		FROM order_items WHERE order_id IN (?) ORDER BY order_id, id`, orderIDs)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)

	var items []domain.OrderItem
	if err := r.db.SelectContext(ctx, &items, query, args...); err != nil {
		return nil, err
	}

	for _, item := range items {
		itemsByOrder[item.OrderID] = append(itemsByOrder[item.OrderID], item)
	}

	for _, id := range orderIDs {
		if _, ok := itemsByOrder[id]; !ok {
			itemsByOrder[id] = []domain.OrderItem{}
		}
	}

	return itemsByOrder, nil
}
