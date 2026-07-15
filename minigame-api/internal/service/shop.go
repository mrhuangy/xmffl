package service

import (
	"context"

	"fpxxl/minigame-api/internal/domain"
	"fpxxl/minigame-api/internal/repository"
)

type ShopService struct {
	repo repository.Repository
}

func NewShopService(repo repository.Repository) *ShopService {
	return &ShopService{repo: repo}
}

func (s *ShopService) Products(ctx context.Context) ([]domain.ShopProduct, error) {
	return s.repo.ListShopProducts(ctx)
}

func (s *ShopService) Purchase(ctx context.Context, playerID uint64, productKey string) (domain.PlayerProgress, string, error) {
	return s.repo.PurchaseProduct(ctx, playerID, productKey)
}
