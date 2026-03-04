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

func TestMenuRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewMenuItemRepository(sqlxDB)

	item := &domain.MenuItem{ID: uuid.New(), Name: "Espresso", Description: "Strong coffee", Price: decimal.NewFromFloat(2.50), Category: "Coffee", IsAvailable: true, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	query := `INSERT INTO menu_items (id, name, description, price, category, is_available, created_at, updated_at)
              VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(item.ID, item.Name, item.Description, item.Price, item.Category, item.IsAvailable, item.CreatedAt, item.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Create(context.Background(), item)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMenuRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewMenuItemRepository(sqlxDB)

	id := uuid.New()
	item := &domain.MenuItem{ID: id, Name: "Latte", Price: decimal.NewFromFloat(4.00), Category: "Coffee"}
	rows := sqlmock.NewRows([]string{"id", "name", "price", "category"}).AddRow(item.ID, item.Name, item.Price, item.Category)

	query := `SELECT * FROM menu_items WHERE id = $1`
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(id).WillReturnRows(rows)

	result, err := repo.GetByID(context.Background(), id)
	assert.NoError(t, err)
	assert.Equal(t, item.ID, result.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMenuRepository_Fetch_WithFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewMenuItemRepository(sqlxDB)

	available := true
	filter := domain.MenuFilter{Category: "Coffee", IsAvailable: &available, Search: "latte", Limit: 5, Offset: 10}

	countQuery := `SELECT COUNT(*) FROM menu_items WHERE category = $1 AND is_available = $2 AND (name ILIKE $3 OR description ILIKE $3)`
	mock.ExpectQuery(regexp.QuoteMeta(countQuery)).WithArgs("Coffee", true, "%latte%").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	dataQuery := `SELECT * FROM menu_items WHERE category = $1 AND is_available = $2 AND (name ILIKE $3 OR description ILIKE $3) ORDER BY created_at DESC, id DESC LIMIT $4 OFFSET $5`
	rows := sqlmock.NewRows([]string{"id", "name", "price", "category"}).
		AddRow(uuid.New(), "Latte", decimal.NewFromFloat(4.20), "Coffee")
	mock.ExpectQuery(regexp.QuoteMeta(dataQuery)).WithArgs("Coffee", true, "%latte%", 5, 10).WillReturnRows(rows)

	result, err := repo.Fetch(context.Background(), filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), result.Total)
	assert.Len(t, result.Items, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMenuRepository_Fetch_WithoutFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewMenuItemRepository(sqlxDB)

	filter := domain.MenuFilter{Limit: 20, Offset: 0}

	countQuery := `SELECT COUNT(*) FROM menu_items`
	mock.ExpectQuery(regexp.QuoteMeta(countQuery)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	dataQuery := `SELECT * FROM menu_items ORDER BY created_at DESC, id DESC LIMIT $1 OFFSET $2`
	rows := sqlmock.NewRows([]string{"id", "name", "price", "category"}).
		AddRow(uuid.New(), "Americano", decimal.NewFromFloat(3.20), "Coffee").
		AddRow(uuid.New(), "Tea", decimal.NewFromFloat(2.00), "Tea")
	mock.ExpectQuery(regexp.QuoteMeta(dataQuery)).WithArgs(20, 0).WillReturnRows(rows)

	result, err := repo.Fetch(context.Background(), filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), result.Total)
	assert.Len(t, result.Items, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMenuRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewMenuItemRepository(sqlxDB)

	item := &domain.MenuItem{ID: uuid.New(), Name: "Updated Latte", Description: "Better Latte", Price: decimal.NewFromFloat(4.50), Category: "Coffee", IsAvailable: true, UpdatedAt: time.Now()}
	query := `UPDATE menu_items SET name=?, description=?, price=?, category=?,
              is_available=?, updated_at=? WHERE id=?`

	mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs(item.Name, item.Description, item.Price, item.Category, item.IsAvailable, item.UpdatedAt, item.ID).WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Update(context.Background(), item)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMenuRepository_Update_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewMenuItemRepository(sqlxDB)

	item := &domain.MenuItem{ID: uuid.New(), Name: "Updated Latte", UpdatedAt: time.Now()}
	query := `UPDATE menu_items SET name=?, description=?, price=?, category=?,
              is_available=?, updated_at=? WHERE id=?`

	mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Update(context.Background(), item)
	assert.ErrorIs(t, err, sql.ErrNoRows)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMenuRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewMenuItemRepository(sqlxDB)

	id := uuid.New()
	query := `DELETE FROM menu_items WHERE id = $1`

	mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs(id).WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Delete(context.Background(), id)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
