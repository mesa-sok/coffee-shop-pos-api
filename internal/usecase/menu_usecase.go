package usecase

import (
	"context"
	"time"

	"coffee-shop-pos/internal/domain"
	"github.com/google/uuid"
)

type menuUsecase struct {
	menuRepo domain.MenuItemRepository
}

func NewMenuUsecase(repo domain.MenuItemRepository) domain.MenuItemUsecase {
	return &menuUsecase{
		menuRepo: repo,
	}
}

func (u *menuUsecase) Create(ctx context.Context, item *domain.MenuItem) error {
	item.ID = uuid.New()
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()
	return u.menuRepo.Create(ctx, item)
}

func (u *menuUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.MenuItem, error) {
	return u.menuRepo.GetByID(ctx, id)
}

func (u *menuUsecase) Fetch(ctx context.Context) ([]domain.MenuItem, error) {
	return u.menuRepo.Fetch(ctx)
}

func (u *menuUsecase) Update(ctx context.Context, item *domain.MenuItem) error {
	item.UpdatedAt = time.Now()
	return u.menuRepo.Update(ctx, item)
}

func (u *menuUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.menuRepo.Delete(ctx, id)
}
