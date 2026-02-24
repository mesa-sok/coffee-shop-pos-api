package postgres

import (
	"context"
	"database/sql"
	"errors"

	"coffee-shop-pos/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type menuRepository struct {
	db *sqlx.DB
}

func NewMenuItemRepository(db *sqlx.DB) domain.MenuItemRepository {
	return &menuRepository{db: db}
}

func (r *menuRepository) Create(ctx context.Context, item *domain.MenuItem) error {
	query := `INSERT INTO menu_items (id, name, description, price, category, is_available, created_at, updated_at)
              VALUES (:id, :name, :description, :price, :category, :is_available, :created_at, :updated_at)`
	_, err := r.db.NamedExecContext(ctx, query, item)
	return err
}

func (r *menuRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.MenuItem, error) {
	var item domain.MenuItem
	query := `SELECT * FROM menu_items WHERE id = $1`
	err := r.db.GetContext(ctx, &item, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *menuRepository) Fetch(ctx context.Context) ([]domain.MenuItem, error) {
	var items []domain.MenuItem
	query := `SELECT * FROM menu_items`
	err := r.db.SelectContext(ctx, &items, query)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *menuRepository) Update(ctx context.Context, item *domain.MenuItem) error {
	query := `UPDATE menu_items SET name=:name, description=:description, price=:price, category=:category,
              is_available=:is_available, updated_at=:updated_at WHERE id=:id`
	result, err := r.db.NamedExecContext(ctx, query, item)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *menuRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM menu_items WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}
