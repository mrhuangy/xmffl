package store

import (
	"context"
	"time"

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
	FindAdminByUsername(ctx context.Context, username string) (domain.AdminUser, error)
	RecordAdminLoginFailure(ctx context.Context, id uint64, lockedUntil *time.Time) error
	RecordAdminLoginSuccess(ctx context.Context, id uint64, ip string) error
	ListAdminPlayers(ctx context.Context, keyword, status string, offset, limit int) ([]domain.AdminPlayerListItem, int, error)
	GetAdminPlayer(ctx context.Context, id uint64) (domain.AdminPlayerDetail, error)
	ListAdmins(ctx context.Context) ([]domain.AdminUser, error)
	GetAdmin(ctx context.Context, id uint64) (domain.AdminUser, error)
	CreateAdmin(ctx context.Context, admin domain.AdminUser) (uint64, error)
	UpdateAdmin(ctx context.Context, admin domain.AdminUser, updatePassword bool) error
	DeleteAdmin(ctx context.Context, id uint64) error
	CountActiveOwners(ctx context.Context) (int, error)
	ListSystemControls(ctx context.Context) ([]domain.SystemControl, error)
	GetSystemControl(ctx context.Context, id uint64) (domain.SystemControl, error)
	CreateSystemControl(ctx context.Context, control domain.SystemControl) (uint64, error)
	UpdateSystemControl(ctx context.Context, control domain.SystemControl) error
	DeleteSystemControl(ctx context.Context, id uint64) error
}
