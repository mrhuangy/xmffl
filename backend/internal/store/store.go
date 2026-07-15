package store

import (
	"context"

	"fpxxl/backend/internal/domain"
)

type Repository interface {
	ListLevels(ctx context.Context, includeDisabled bool) ([]domain.LevelConfig, error)
	UpsertLevel(ctx context.Context, level domain.LevelConfig) error
	GetAdConfig(ctx context.Context) (domain.AdFrequencyConfig, error)
	SaveAdConfig(ctx context.Context, cfg domain.AdFrequencyConfig) error
	GetProgress(ctx context.Context, openID string) (domain.PlayerProgress, error)
	SaveProgress(ctx context.Context, progress domain.PlayerProgress) error
	SaveLevelResult(ctx context.Context, result domain.LevelResult) error
	SubmitLeaderboard(ctx context.Context, entry domain.LeaderboardEntry) error
	ListLeaderboard(ctx context.Context, levelID int, limit int) ([]domain.LeaderboardEntry, error)
	SaveEvents(ctx context.Context, events []domain.GameEvent) error
}
