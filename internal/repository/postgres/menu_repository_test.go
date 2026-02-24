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

	item := &domain.MenuItem{
		ID:          uuid.New(),
		Name:        "Espresso",
		Description: "Strong coffee",
		Price:       decimal.NewFromFloat(2.50),
		Category:    "Coffee",
		IsAvailable: true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

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
	item := &domain.MenuItem{
		ID:          id,
		Name:        "Latte",
		Price:       decimal.NewFromFloat(4.00),
		Category:    "Coffee",
	}

	rows := sqlmock.NewRows([]string{"id", "name", "price", "category"}).
		AddRow(item.ID, item.Name, item.Price, item.Category)

	query := `SELECT * FROM menu_items WHERE id = $1`
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(id).
		WillReturnRows(rows)

	result, err := repo.GetByID(context.Background(), id)
	assert.NoError(t, err)
	assert.Equal(t, item.ID, result.ID)
	assert.Equal(t, item.Name, result.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMenuRepository_Fetch(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewMenuItemRepository(sqlxDB)

	rows := sqlmock.NewRows([]string{"id", "name", "price"}).
		AddRow(uuid.New(), "Tea", decimal.NewFromFloat(2.00)).
		AddRow(uuid.New(), "Cake", decimal.NewFromFloat(3.50))

	query := `SELECT * FROM menu_items`
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnRows(rows)

	results, err := repo.Fetch(context.Background())
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMenuRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewMenuItemRepository(sqlxDB)

	item := &domain.MenuItem{
		ID:          uuid.New(),
		Name:        "Updated Latte",
		Description: "Better Latte",
		Price:       decimal.NewFromFloat(4.50),
		Category:    "Coffee",
		IsAvailable: true,
		UpdatedAt:   time.Now(),
	}

	query := `UPDATE menu_items SET name=?, description=?, price=?, category=?,
              is_available=?, updated_at=? WHERE id=?`

	mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(item.Name, item.Description, item.Price, item.Category, item.IsAvailable, item.UpdatedAt, item.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

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

	item := &domain.MenuItem{
		ID:          uuid.New(),
		Name:        "Updated Latte",
		UpdatedAt:   time.Now(),
	}

	query := `UPDATE menu_items SET name=?, description=?, price=?, category=?,
              is_available=?, updated_at=? WHERE id=?`

	mock.ExpectExec(regexp.QuoteMeta(query)).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected

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

	mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Delete(context.Background(), id)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
