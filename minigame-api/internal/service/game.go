package service

import (
	"context"

	"fpxxl/minigame-api/internal/domain"
	"fpxxl/minigame-api/internal/repository"
)

type GameService struct {
	repo repository.Repository
}

func NewGameService(repo repository.Repository) *GameService {
	return &GameService{repo: repo}
}

func (s *GameService) StartLevel(ctx context.Context, playerID uint64, levelID int) (domain.PlayerProgress, error) {
	return s.repo.StartLevel(ctx, playerID, levelID)
}

func (s *GameService) SubmitLevelResult(ctx context.Context, player domain.Player, result domain.LevelResult) (domain.PlayerProgress, error) {
	return s.repo.SubmitLevelResult(ctx, player, result)
}
