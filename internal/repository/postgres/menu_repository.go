package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

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
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func buildMenuWhereClause(filter domain.MenuFilter, args []any) (string, []any) {
	conditions := make([]string, 0)

	if filter.Category != "" {
		args = append(args, filter.Category)
		conditions = append(conditions, fmt.Sprintf("category = $%d", len(args)))
	}
	if filter.IsAvailable != nil {
		args = append(args, *filter.IsAvailable)
		conditions = append(conditions, fmt.Sprintf("is_available = $%d", len(args)))
	}
	if filter.Search != "" {
		args = append(args, "%"+filter.Search+"%")
		placeholder := fmt.Sprintf("$%d", len(args))
		conditions = append(conditions, fmt.Sprintf("(name ILIKE %s OR description ILIKE %s)", placeholder, placeholder))
	}

	if len(conditions) == 0 {
		return "", args
	}

	return " WHERE " + strings.Join(conditions, " AND "), args
}

func (r *menuRepository) Fetch(ctx context.Context, filter domain.MenuFilter) (*domain.MenuListResult, error) {
	args := make([]any, 0)
	whereClause, args := buildMenuWhereClause(filter, args)

	countQuery := `SELECT COUNT(*) FROM menu_items` + whereClause
	var total int64
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, err
	}

	pagedArgs := append([]any{}, args...)
	pagedArgs = append(pagedArgs, filter.Limit, filter.Offset)
	limitPlaceholder := fmt.Sprintf("$%d", len(pagedArgs)-1)
	offsetPlaceholder := fmt.Sprintf("$%d", len(pagedArgs))

	query := `SELECT * FROM menu_items` + whereClause +
		` ORDER BY created_at DESC, id DESC LIMIT ` + limitPlaceholder + ` OFFSET ` + offsetPlaceholder

	var items []domain.MenuItem
	err := r.db.SelectContext(ctx, &items, query, pagedArgs...)
	if err != nil {
		return nil, err
	}

	return &domain.MenuListResult{Items: items, Total: total}, nil
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
		return sql.ErrNoRows
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
		return sql.ErrNoRows
	}

	return nil
}
