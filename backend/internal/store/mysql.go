package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"fpxxl/backend/internal/domain"
)

type MySQLStore struct {
	db *sql.DB
}

func NewMySQLStore(db *sql.DB) *MySQLStore {
	return &MySQLStore{db: db}
}

func (s *MySQLStore) ListLevels(ctx context.Context, includeDisabled bool) ([]domain.LevelConfig, error) {
	query := `SELECT level_id, rows_count, cols_count, pair_count, mode, theme_id, initial_preview_ms,
		flip_back_delay_ms, level_time_limit_seconds, max_mismatch_count, excellent_step_threshold,
		normal_step_threshold, excellent_time_threshold, normal_time_threshold, time_limit_seconds,
		step_limit, enabled, updated_at FROM level_configs`
	if !includeDisabled {
		query += " WHERE enabled = 1"
	}
	query += " ORDER BY level_id"

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	levels := make([]domain.LevelConfig, 0)
	for rows.Next() {
		level, err := scanLevel(rows)
		if err != nil {
			return nil, err
		}
		levels = append(levels, level)
	}
	return levels, rows.Err()
}

func (s *MySQLStore) UpsertLevel(ctx context.Context, level domain.LevelConfig) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO level_configs (
		level_id, rows_count, cols_count, pair_count, mode, theme_id, initial_preview_ms,
		flip_back_delay_ms, level_time_limit_seconds, max_mismatch_count, excellent_step_threshold,
		normal_step_threshold, excellent_time_threshold, normal_time_threshold, time_limit_seconds,
		step_limit, enabled
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE
		rows_count = VALUES(rows_count),
		cols_count = VALUES(cols_count),
		pair_count = VALUES(pair_count),
		mode = VALUES(mode),
		theme_id = VALUES(theme_id),
		initial_preview_ms = VALUES(initial_preview_ms),
		flip_back_delay_ms = VALUES(flip_back_delay_ms),
		level_time_limit_seconds = VALUES(level_time_limit_seconds),
		max_mismatch_count = VALUES(max_mismatch_count),
		excellent_step_threshold = VALUES(excellent_step_threshold),
		normal_step_threshold = VALUES(normal_step_threshold),
		excellent_time_threshold = VALUES(excellent_time_threshold),
		normal_time_threshold = VALUES(normal_time_threshold),
		time_limit_seconds = VALUES(time_limit_seconds),
		step_limit = VALUES(step_limit),
		enabled = VALUES(enabled)`,
		level.LevelID, level.Rows, level.Cols, level.PairCount, level.Mode, level.ThemeID,
		level.InitialPreviewMs, level.FlipBackDelayMs, level.LevelTimeLimitSeconds,
		level.MaxMismatchCount, level.ExcellentStepThreshold, level.NormalStepThreshold,
		level.ExcellentTimeThreshold, level.NormalTimeThreshold, level.TimeLimitSeconds,
		level.StepLimit, level.Enabled,
	)
	return err
}

func (s *MySQLStore) GetAdConfig(ctx context.Context) (domain.AdFrequencyConfig, error) {
	var cfg domain.AdFrequencyConfig
	var scenesJSON string
	var updatedAt time.Time
	err := s.db.QueryRowContext(ctx, `SELECT no_interstitial_before_level, interstitial_every_levels,
		max_interstitial_per_day, max_revive_per_level, banner_enabled_scenes, updated_at
		FROM ad_frequency_configs WHERE id = 1`).Scan(
		&cfg.NoInterstitialBeforeLevel,
		&cfg.InterstitialEveryLevels,
		&cfg.MaxInterstitialPerDay,
		&cfg.MaxRevivePerLevel,
		&scenesJSON,
		&updatedAt,
	)
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal([]byte(scenesJSON), &cfg.BannerEnabledScenes); err != nil {
		return cfg, err
	}
	cfg.UpdatedAt = updatedAt.Format(time.RFC3339)
	return cfg, nil
}

func (s *MySQLStore) SaveAdConfig(ctx context.Context, cfg domain.AdFrequencyConfig) error {
	scenesJSON, err := json.Marshal(cfg.BannerEnabledScenes)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO ad_frequency_configs (
		id, no_interstitial_before_level, interstitial_every_levels, max_interstitial_per_day,
		max_revive_per_level, banner_enabled_scenes
	) VALUES (1, ?, ?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE
		no_interstitial_before_level = VALUES(no_interstitial_before_level),
		interstitial_every_levels = VALUES(interstitial_every_levels),
		max_interstitial_per_day = VALUES(max_interstitial_per_day),
		max_revive_per_level = VALUES(max_revive_per_level),
		banner_enabled_scenes = VALUES(banner_enabled_scenes)`,
		cfg.NoInterstitialBeforeLevel, cfg.InterstitialEveryLevels, cfg.MaxInterstitialPerDay,
		cfg.MaxRevivePerLevel, string(scenesJSON),
	)
	return err
}

func (s *MySQLStore) GetProgress(ctx context.Context, openID string) (domain.PlayerProgress, error) {
	var progress domain.PlayerProgress
	var levelStarsJSON, completedLevelsJSON string
	err := s.db.QueryRowContext(ctx, `SELECT open_id, current_level, coins, hints, level_stars,
		completed_levels, updated_at FROM player_progress WHERE open_id = ?`, openID).Scan(
		&progress.OpenID, &progress.CurrentLevel, &progress.Coins, &progress.Hints,
		&levelStarsJSON, &completedLevelsJSON, &progress.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.PlayerProgress{
			OpenID:          openID,
			CurrentLevel:    1,
			Coins:           0,
			Hints:           0,
			LevelStars:      map[string]int{},
			CompletedLevels: []int{},
			UpdatedAt:       time.Now(),
		}, nil
	}
	if err != nil {
		return progress, err
	}
	if err := json.Unmarshal([]byte(levelStarsJSON), &progress.LevelStars); err != nil {
		return progress, err
	}
	if err := json.Unmarshal([]byte(completedLevelsJSON), &progress.CompletedLevels); err != nil {
		return progress, err
	}
	return progress, nil
}

func (s *MySQLStore) SaveProgress(ctx context.Context, progress domain.PlayerProgress) error {
	levelStarsJSON, err := json.Marshal(progress.LevelStars)
	if err != nil {
		return err
	}
	completedLevelsJSON, err := json.Marshal(progress.CompletedLevels)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO player_progress (
		open_id, current_level, coins, hints, level_stars, completed_levels
	) VALUES (?, ?, ?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE
		current_level = VALUES(current_level),
		coins = VALUES(coins),
		hints = VALUES(hints),
		level_stars = VALUES(level_stars),
		completed_levels = VALUES(completed_levels)`,
		progress.OpenID, progress.CurrentLevel, progress.Coins, progress.Hints,
		string(levelStarsJSON), string(completedLevelsJSON),
	)
	return err
}

func (s *MySQLStore) SaveLevelResult(ctx context.Context, result domain.LevelResult) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO level_results (
		open_id, level_id, success, reason, steps, mismatch_count, elapsed_ms, stars, coins_earned, used_hints
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		result.OpenID, result.LevelID, result.Success, result.Reason, result.Steps, result.MismatchCount,
		result.ElapsedMs, result.Stars, result.CoinsEarned, result.UsedHints,
	)
	return err
}

func (s *MySQLStore) SubmitLeaderboard(ctx context.Context, entry domain.LeaderboardEntry) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO leaderboard_entries (
		open_id, nickname, level_id, stars, steps, elapsed_ms
	) VALUES (?, ?, ?, ?, ?, ?)`,
		entry.OpenID, entry.Nickname, entry.LevelID, entry.Stars, entry.Steps, entry.ElapsedMs,
	)
	return err
}

func (s *MySQLStore) ListLeaderboard(ctx context.Context, levelID int, limit int) ([]domain.LeaderboardEntry, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT open_id, nickname, level_id, stars, steps, elapsed_ms, submitted_at
		FROM leaderboard_entries
		WHERE (? = 0 OR level_id = ?)
		ORDER BY stars DESC, steps ASC, elapsed_ms ASC, submitted_at ASC
		LIMIT ?`, levelID, levelID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]domain.LeaderboardEntry, 0)
	for rows.Next() {
		var entry domain.LeaderboardEntry
		if err := rows.Scan(&entry.OpenID, &entry.Nickname, &entry.LevelID, &entry.Stars, &entry.Steps, &entry.ElapsedMs, &entry.SubmittedAt); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

func (s *MySQLStore) SaveEvents(ctx context.Context, events []domain.GameEvent) error {
	if len(events) == 0 {
		return nil
	}

	parts := make([]string, 0, len(events))
	args := make([]any, 0, len(events)*5)
	for _, event := range events {
		payload, err := json.Marshal(event.Payload)
		if err != nil {
			return err
		}
		parts = append(parts, "(?, ?, ?, ?, ?)")
		args = append(args, event.EventID, event.OpenID, event.EventType, string(payload), event.CreatedAt)
	}
	_, err := s.db.ExecContext(ctx, "INSERT INTO game_events (event_id, open_id, event_type, payload, created_at) VALUES "+strings.Join(parts, ","), args...)
	return err
}

type levelScanner interface {
	Scan(dest ...any) error
}

func scanLevel(scanner levelScanner) (domain.LevelConfig, error) {
	var level domain.LevelConfig
	var excellentTime, normalTime, timeLimit, stepLimit sql.NullInt64
	var enabled bool
	if err := scanner.Scan(
		&level.LevelID, &level.Rows, &level.Cols, &level.PairCount, &level.Mode, &level.ThemeID,
		&level.InitialPreviewMs, &level.FlipBackDelayMs, &level.LevelTimeLimitSeconds,
		&level.MaxMismatchCount, &level.ExcellentStepThreshold, &level.NormalStepThreshold,
		&excellentTime, &normalTime, &timeLimit, &stepLimit, &enabled, &level.UpdatedAt,
	); err != nil {
		return level, err
	}
	level.Enabled = enabled
	level.ExcellentTimeThreshold = nullableInt(excellentTime)
	level.NormalTimeThreshold = nullableInt(normalTime)
	level.TimeLimitSeconds = nullableInt(timeLimit)
	level.StepLimit = nullableInt(stepLimit)
	return level, validateLevel(level)
}

func nullableInt(value sql.NullInt64) *int {
	if !value.Valid {
		return nil
	}
	v := int(value.Int64)
	return &v
}

func validateLevel(level domain.LevelConfig) error {
	if level.LevelID <= 0 {
		return fmt.Errorf("levelId must be positive")
	}
	if level.Rows <= 0 || level.Cols <= 0 || level.PairCount <= 0 {
		return fmt.Errorf("rows, cols and pairCount must be positive")
	}
	if level.Rows*level.Cols != level.PairCount*2 {
		return fmt.Errorf("rows * cols must equal pairCount * 2")
	}
	return nil
}
