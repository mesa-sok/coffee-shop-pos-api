package postgres

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"coffee-shop-pos/internal/domain"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestOrderRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewOrderRepository(sqlxDB)

	orderID := uuid.New()
	order := &domain.Order{
		ID:          orderID,
		OrderNumber: "ORD-123",
		Status:      domain.OrderStatusPending,
		Subtotal:    decimal.NewFromFloat(10),
		Tax:         decimal.NewFromFloat(1),
		Total:       decimal.NewFromFloat(11),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Items: []domain.OrderItem{{
			ID:         uuid.New(),
			OrderID:    orderID,
			MenuItemID: uuid.New(),
			Quantity:   2,
			UnitPrice:  decimal.NewFromFloat(5),
			LineTotal:  decimal.NewFromFloat(10),
		}},
	}

	mock.ExpectBegin()
	orderQuery := `INSERT INTO orders (id, order_number, status, subtotal, tax, total, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	mock.ExpectExec(regexp.QuoteMeta(orderQuery)).
		WithArgs(order.ID, order.OrderNumber, order.Status, order.Subtotal, order.Tax, order.Total, order.CreatedAt, order.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	itemQuery := `INSERT INTO order_items (id, order_id, menu_item_id, quantity, unit_price, line_total)
		VALUES (?, ?, ?, ?, ?, ?)`
	item := order.Items[0]
	mock.ExpectExec(regexp.QuoteMeta(itemQuery)).
		WithArgs(item.ID, item.OrderID, item.MenuItemID, item.Quantity, item.UnitPrice, item.LineTotal).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = repo.Create(context.Background(), order)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewOrderRepository(sqlxDB)
	orderID := uuid.New()

	joinRows := sqlmock.NewRows([]string{"id", "order_number", "status", "subtotal", "tax", "total", "created_at", "updated_at", "item_id", "order_id", "menu_item_id", "quantity", "unit_price", "line_total"}).
		AddRow(orderID, "ORD-1", domain.OrderStatusPending, decimal.NewFromFloat(10), decimal.NewFromFloat(1), decimal.NewFromFloat(11), time.Now(), time.Now(), uuid.New(), orderID, uuid.New(), 2, decimal.NewFromFloat(5), decimal.NewFromFloat(10))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT o.id, o.order_number, o.status, o.subtotal, o.tax, o.total, o.created_at, o.updated_at,
		oi.id AS item_id, oi.order_id, oi.menu_item_id, oi.quantity, oi.unit_price, oi.line_total
		FROM orders o
		LEFT JOIN order_items oi ON oi.order_id = o.id
		WHERE o.id = $1
		ORDER BY oi.id`)).
		WithArgs(orderID).
		WillReturnRows(joinRows)

	order, err := repo.GetByID(context.Background(), orderID)
	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.Len(t, order.Items, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewOrderRepository(sqlxDB)
	orderID := uuid.New()

	rows := sqlmock.NewRows([]string{"id", "order_number", "status", "subtotal", "tax", "total", "created_at", "updated_at"}).
		AddRow(orderID, "ORD-1", domain.OrderStatusPending, decimal.NewFromFloat(10), decimal.NewFromFloat(1), decimal.NewFromFloat(11), time.Now(), time.Now())
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, order_number, status, subtotal, tax, total, created_at, updated_at FROM orders ORDER BY created_at DESC`)).WillReturnRows(rows)

	itemRows := sqlmock.NewRows([]string{"id", "order_id", "menu_item_id", "quantity", "unit_price", "line_total"}).
		AddRow(uuid.New(), orderID, uuid.New(), 1, decimal.NewFromFloat(10), decimal.NewFromFloat(10))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, order_id, menu_item_id, quantity, unit_price, line_total
		FROM order_items WHERE order_id IN (?) ORDER BY order_id, id`)).
		WithArgs(orderID).
		WillReturnRows(itemRows)

	orders, err := repo.List(context.Background())
	assert.NoError(t, err)
	assert.Len(t, orders, 1)
	assert.Len(t, orders[0].Items, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_UpdateStatus_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewOrderRepository(sqlxDB)
	id := uuid.New()

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3`)).
		WithArgs(domain.OrderStatusPaid, sqlmock.AnyArg(), id).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.UpdateStatus(context.Background(), id, domain.OrderStatusPaid, time.Now())
	assert.ErrorIs(t, err, sql.ErrNoRows)
}
