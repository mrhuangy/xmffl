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
		step_limit, show_steps, show_timer, show_mismatch, hint_highlight_ms, coin_reward_base, stamina_cost,
		enabled, version, updated_at FROM level_configs`
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
		step_limit, show_steps, show_timer, show_mismatch, hint_highlight_ms, coin_reward_base, stamina_cost, enabled, version
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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
		show_steps = VALUES(show_steps),
		show_timer = VALUES(show_timer),
		show_mismatch = VALUES(show_mismatch),
		hint_highlight_ms = VALUES(hint_highlight_ms),
		coin_reward_base = VALUES(coin_reward_base),
		stamina_cost = VALUES(stamina_cost),
		enabled = VALUES(enabled),
		version = version + 1`,
		level.LevelID, level.Rows, level.Cols, level.PairCount, level.Mode, level.ThemeID,
		level.InitialPreviewMs, level.FlipBackDelayMs, level.LevelTimeLimitSeconds,
		level.MaxMismatchCount, level.ExcellentStepThreshold, level.NormalStepThreshold,
		level.ExcellentTimeThreshold, level.NormalTimeThreshold, level.TimeLimitSeconds,
		level.StepLimit, level.ShowSteps, level.ShowTimer, level.ShowMismatch, level.HintHighlightMs,
		level.CoinRewardBase, level.StaminaCost, level.Enabled, max(level.Version, 1),
	)
	return err
}

func (s *MySQLStore) GetAdConfig(ctx context.Context) (domain.AdFrequencyConfig, error) {
	var cfg domain.AdFrequencyConfig
	var scenesJSON string
	var updatedAt time.Time
	err := s.db.QueryRowContext(ctx, `SELECT no_interstitial_before_level, interstitial_every_levels,
		max_interstitial_per_day, max_revive_per_level, banner_enabled_scenes, version, updated_at
		FROM ad_frequency_configs WHERE id = 1`).Scan(
		&cfg.NoInterstitialBeforeLevel,
		&cfg.InterstitialEveryLevels,
		&cfg.MaxInterstitialPerDay,
		&cfg.MaxRevivePerLevel,
		&scenesJSON,
		&cfg.Version,
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
		max_revive_per_level, banner_enabled_scenes, version
	) VALUES (1, ?, ?, ?, ?, ?, 1)
	ON DUPLICATE KEY UPDATE
		no_interstitial_before_level = VALUES(no_interstitial_before_level),
		interstitial_every_levels = VALUES(interstitial_every_levels),
		max_interstitial_per_day = VALUES(max_interstitial_per_day),
		max_revive_per_level = VALUES(max_revive_per_level),
		banner_enabled_scenes = VALUES(banner_enabled_scenes),
		version = version + 1`,
		cfg.NoInterstitialBeforeLevel, cfg.InterstitialEveryLevels, cfg.MaxInterstitialPerDay,
		cfg.MaxRevivePerLevel, string(scenesJSON),
	)
	return err
}

func (s *MySQLStore) GetProgress(ctx context.Context, openID string) (domain.PlayerProgress, error) {
	var progress domain.PlayerProgress
	var levelStarsJSON, completedLevelsJSON string
	err := s.db.QueryRowContext(ctx, `SELECT p.id, p.open_id, pp.current_level, pp.coins, pp.stamina,
		pp.max_stamina, pp.hints, pp.preview_again_count, pp.remove_pair_count, pp.level_stars,
		pp.completed_levels, pp.next_stamina_recover_at, pp.updated_at
		FROM players p JOIN player_progress pp ON pp.player_id = p.id WHERE p.open_id = ?`, openID).Scan(
		&progress.PlayerID, &progress.OpenID, &progress.CurrentLevel, &progress.Coins, &progress.Stamina,
		&progress.MaxStamina, &progress.Hints, &progress.PreviewAgain, &progress.RemovePair,
		&levelStarsJSON, &completedLevelsJSON, &progress.NextRecoverAt, &progress.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.PlayerProgress{
			OpenID:          openID,
			CurrentLevel:    1,
			Coins:           0,
			Stamina:         5,
			MaxStamina:      5,
			Hints:           3,
			PreviewAgain:    3,
			RemovePair:      3,
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
	playerID, err := s.ensurePlayer(ctx, progress.OpenID)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO player_progress (
		player_id, current_level, coins, stamina, max_stamina, hints, preview_again_count,
		remove_pair_count, level_stars, completed_levels, next_stamina_recover_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE
		current_level = VALUES(current_level),
		coins = VALUES(coins),
		stamina = VALUES(stamina),
		max_stamina = VALUES(max_stamina),
		hints = VALUES(hints),
		preview_again_count = VALUES(preview_again_count),
		remove_pair_count = VALUES(remove_pair_count),
		level_stars = VALUES(level_stars),
		completed_levels = VALUES(completed_levels),
		next_stamina_recover_at = VALUES(next_stamina_recover_at)`,
		playerID, progress.CurrentLevel, progress.Coins, progress.Stamina, progress.MaxStamina,
		progress.Hints, progress.PreviewAgain, progress.RemovePair, string(levelStarsJSON),
		string(completedLevelsJSON), progress.NextRecoverAt,
	)
	return err
}

func (s *MySQLStore) SaveLevelResult(ctx context.Context, result domain.LevelResult) error {
	playerID, err := s.ensurePlayer(ctx, result.OpenID)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO level_results (
		player_id, level_id, success, reason, steps, mismatch_count, elapsed_ms, stars, coins_earned, used_hints
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		playerID, result.LevelID, result.Success, result.Reason, result.Steps, result.MismatchCount,
		result.ElapsedMs, result.Stars, result.CoinsEarned, result.UsedHints,
	)
	return err
}

func (s *MySQLStore) SubmitLeaderboard(ctx context.Context, entry domain.LeaderboardEntry) error {
	playerID, err := s.ensurePlayer(ctx, entry.OpenID)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO leaderboard_entries (
		player_id, nickname, level_id, stars, steps, elapsed_ms
	) VALUES (?, ?, ?, ?, ?, ?)`,
		playerID, entry.Nickname, entry.LevelID, entry.Stars, entry.Steps, entry.ElapsedMs,
	)
	return err
}

func (s *MySQLStore) ListLeaderboard(ctx context.Context, levelID int, limit int) ([]domain.LeaderboardEntry, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT p.open_id, le.nickname, le.level_id, le.stars, le.steps, le.elapsed_ms, le.submitted_at
		FROM leaderboard_entries le JOIN players p ON p.id = le.player_id
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
	args := make([]any, 0, len(events)*6)
	for _, event := range events {
		payload, err := json.Marshal(event.Payload)
		if err != nil {
			return err
		}
		var playerID any
		if event.OpenID != "" {
			id, err := s.ensurePlayer(ctx, event.OpenID)
			if err != nil {
				return err
			}
			playerID = id
		}
		parts = append(parts, "(?, ?, ?, ?, ?, ?)")
		args = append(args, event.EventID, playerID, event.EventType, event.LevelID, string(payload), event.CreatedAt)
	}
	_, err := s.db.ExecContext(ctx, "INSERT INTO game_events (event_id, player_id, event_type, level_id, payload, created_at) VALUES "+strings.Join(parts, ","), args...)
	return err
}

func (s *MySQLStore) ensurePlayer(ctx context.Context, openID string) (uint64, error) {
	if openID == "" {
		return 0, errors.New("openId is required")
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO players (open_id, last_login_at)
		VALUES (?, CURRENT_TIMESTAMP)
		ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id)`, openID)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return uint64(id), nil
}

func (s *MySQLStore) FindAdminByUsername(ctx context.Context, username string) (domain.AdminUser, error) {
	return scanAdmin(s.db.QueryRowContext(ctx, `SELECT id, username, email, password_hash, display_name, role,
		permissions, status, failed_login_attempts, locked_until, password_changed_at, last_login_at,
		last_login_ip, created_by, created_at, updated_at FROM admin_users WHERE username = ?`, username))
}

func (s *MySQLStore) RecordAdminLoginFailure(ctx context.Context, id uint64, lockedUntil *time.Time) error {
	status := "active"
	if lockedUntil != nil {
		status = "locked"
	}
	_, err := s.db.ExecContext(ctx, `UPDATE admin_users SET failed_login_attempts = failed_login_attempts + 1,
		locked_until = ?, status = ? WHERE id = ?`, lockedUntil, status, id)
	return err
}

func (s *MySQLStore) RecordAdminLoginSuccess(ctx context.Context, id uint64, ip string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE admin_users SET failed_login_attempts = 0, locked_until = NULL,
		status = 'active', last_login_at = CURRENT_TIMESTAMP, last_login_ip = ? WHERE id = ?`, ip, id)
	return err
}

func (s *MySQLStore) ListAdminPlayers(ctx context.Context, keyword, status string, offset, limit int) ([]domain.AdminPlayerListItem, int, error) {
	where := " WHERE 1 = 1"
	args := make([]any, 0, 4)
	if keyword != "" {
		where += " AND (p.open_id LIKE ? OR p.nickname LIKE ?)"
		pattern := "%" + keyword + "%"
		args = append(args, pattern, pattern)
	}
	if status != "" {
		where += " AND p.status = ?"
		args = append(args, status)
	}
	var total int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM players p"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `SELECT p.id, p.open_id, p.nickname, p.avatar_url, p.status, p.last_login_at, p.created_at,
		COALESCE(pp.current_level, 1), COALESCE(pp.coins, 0), COALESCE(pp.stamina, 0),
		COALESCE(pp.max_stamina, 5), COALESCE(pp.hints, 0), COALESCE(JSON_LENGTH(pp.completed_levels), 0),
		COALESCE(stats.total_games, 0), COALESCE(stats.successful_games, 0)
		FROM players p LEFT JOIN player_progress pp ON pp.player_id = p.id
		LEFT JOIN (SELECT player_id, COUNT(*) total_games, SUM(success = 1) successful_games
			FROM level_results GROUP BY player_id) stats ON stats.player_id = p.id` + where +
		" ORDER BY COALESCE(p.last_login_at, p.created_at) DESC, p.id DESC LIMIT ? OFFSET ?"
	queryArgs := append(append([]any{}, args...), limit, offset)
	rows, err := s.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	players := make([]domain.AdminPlayerListItem, 0)
	for rows.Next() {
		var player domain.AdminPlayerListItem
		if err := rows.Scan(&player.ID, &player.OpenID, &player.Nickname, &player.AvatarURL, &player.Status,
			&player.LastLoginAt, &player.CreatedAt, &player.CurrentLevel, &player.Coins, &player.Stamina,
			&player.MaxStamina, &player.Hints, &player.CompletedCount, &player.TotalGames, &player.SuccessfulGames); err != nil {
			return nil, 0, err
		}
		players = append(players, player)
	}
	return players, total, rows.Err()
}

func (s *MySQLStore) GetAdminPlayer(ctx context.Context, id uint64) (domain.AdminPlayerDetail, error) {
	var detail domain.AdminPlayerDetail
	var starsJSON, completedJSON []byte
	err := s.db.QueryRowContext(ctx, `SELECT p.id, p.open_id, p.nickname, p.avatar_url, p.status, p.last_login_at, p.created_at,
		COALESCE(pp.current_level, 1), COALESCE(pp.coins, 0), COALESCE(pp.stamina, 0), COALESCE(pp.max_stamina, 5),
		COALESCE(pp.hints, 0), COALESCE(JSON_LENGTH(pp.completed_levels), 0),
		(SELECT COUNT(*) FROM level_results lr WHERE lr.player_id = p.id),
		(SELECT COUNT(*) FROM level_results lr WHERE lr.player_id = p.id AND lr.success = 1),
		COALESCE(pp.preview_again_count, 0), COALESCE(pp.remove_pair_count, 0),
		COALESCE(pp.level_stars, JSON_OBJECT()), COALESCE(pp.completed_levels, JSON_ARRAY()),
		pp.next_stamina_recover_at, pp.updated_at
		FROM players p LEFT JOIN player_progress pp ON pp.player_id = p.id WHERE p.id = ?`, id).Scan(
		&detail.Player.ID, &detail.Player.OpenID, &detail.Player.Nickname, &detail.Player.AvatarURL,
		&detail.Player.Status, &detail.Player.LastLoginAt, &detail.Player.CreatedAt,
		&detail.Player.CurrentLevel, &detail.Player.Coins, &detail.Player.Stamina, &detail.Player.MaxStamina,
		&detail.Player.Hints, &detail.Player.CompletedCount, &detail.Player.TotalGames, &detail.Player.SuccessfulGames,
		&detail.Progress.PreviewAgainCount, &detail.Progress.RemovePairCount, &starsJSON, &completedJSON,
		&detail.Progress.NextStaminaRecoverAt, &detail.Progress.UpdatedAt)
	if err != nil {
		return detail, err
	}
	detail.Progress.CurrentLevel = detail.Player.CurrentLevel
	detail.Progress.Coins = detail.Player.Coins
	detail.Progress.Stamina = detail.Player.Stamina
	detail.Progress.MaxStamina = detail.Player.MaxStamina
	detail.Progress.Hints = detail.Player.Hints
	if err := json.Unmarshal(starsJSON, &detail.Progress.LevelStars); err != nil {
		return detail, err
	}
	if err := json.Unmarshal(completedJSON, &detail.Progress.CompletedLevels); err != nil {
		return detail, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, level_id, success, reason, steps, mismatch_count,
		elapsed_ms, stars, coins_earned, completed_at FROM level_results
		WHERE player_id = ? ORDER BY completed_at DESC, id DESC LIMIT 20`, id)
	if err != nil {
		return detail, err
	}
	defer rows.Close()
	detail.RecentResults = make([]domain.AdminLevelResult, 0)
	for rows.Next() {
		var result domain.AdminLevelResult
		if err := rows.Scan(&result.ID, &result.LevelID, &result.Success, &result.Reason, &result.Steps,
			&result.MismatchCount, &result.ElapsedMs, &result.Stars, &result.CoinsEarned, &result.CompletedAt); err != nil {
			return detail, err
		}
		detail.RecentResults = append(detail.RecentResults, result)
	}
	return detail, rows.Err()
}

func (s *MySQLStore) ListAdmins(ctx context.Context) ([]domain.AdminUser, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, username, email, password_hash, display_name, role,
		permissions, status, failed_login_attempts, locked_until, password_changed_at, last_login_at,
		last_login_ip, created_by, created_at, updated_at FROM admin_users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	admins := make([]domain.AdminUser, 0)
	for rows.Next() {
		admin, err := scanAdmin(rows)
		if err != nil {
			return nil, err
		}
		admins = append(admins, admin)
	}
	return admins, rows.Err()
}

func (s *MySQLStore) GetAdmin(ctx context.Context, id uint64) (domain.AdminUser, error) {
	return scanAdmin(s.db.QueryRowContext(ctx, `SELECT id, username, email, password_hash, display_name, role,
		permissions, status, failed_login_attempts, locked_until, password_changed_at, last_login_at,
		last_login_ip, created_by, created_at, updated_at FROM admin_users WHERE id = ?`, id))
}

func (s *MySQLStore) CreateAdmin(ctx context.Context, admin domain.AdminUser) (uint64, error) {
	permissions, err := json.Marshal(admin.Permissions)
	if err != nil {
		return 0, err
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO admin_users
		(username, email, password_hash, display_name, role, permissions, status, created_by, password_changed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`, admin.Username, admin.Email, admin.PasswordHash,
		admin.DisplayName, admin.Role, nullableJSON(permissions, admin.Permissions != nil), admin.Status, admin.CreatedBy)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return uint64(id), err
}

func (s *MySQLStore) UpdateAdmin(ctx context.Context, admin domain.AdminUser, updatePassword bool) error {
	permissions, err := json.Marshal(admin.Permissions)
	if err != nil {
		return err
	}
	if updatePassword {
		_, err = s.db.ExecContext(ctx, `UPDATE admin_users SET username=?, email=?, password_hash=?, display_name=?,
			role=?, permissions=?, status=?, failed_login_attempts=0, locked_until=NULL,
			password_changed_at=CURRENT_TIMESTAMP WHERE id=?`, admin.Username, admin.Email, admin.PasswordHash,
			admin.DisplayName, admin.Role, nullableJSON(permissions, admin.Permissions != nil), admin.Status, admin.ID)
	} else {
		_, err = s.db.ExecContext(ctx, `UPDATE admin_users SET username=?, email=?, display_name=?, role=?,
			permissions=?, status=?, failed_login_attempts=IF(?='active',0,failed_login_attempts),
			locked_until=IF(?='active',NULL,locked_until) WHERE id=?`, admin.Username, admin.Email,
			admin.DisplayName, admin.Role, nullableJSON(permissions, admin.Permissions != nil), admin.Status,
			admin.Status, admin.Status, admin.ID)
	}
	return err
}

func (s *MySQLStore) DeleteAdmin(ctx context.Context, id uint64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `UPDATE admin_users SET created_by=NULL WHERE created_by=?`, id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM admin_users WHERE id=?`, id); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *MySQLStore) CountActiveOwners(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM admin_users WHERE role='owner' AND status='active'`).Scan(&count)
	return count, err
}

type adminScanner interface{ Scan(dest ...any) error }

func scanAdmin(scanner adminScanner) (domain.AdminUser, error) {
	var admin domain.AdminUser
	var permissionsJSON []byte
	err := scanner.Scan(&admin.ID, &admin.Username, &admin.Email, &admin.PasswordHash, &admin.DisplayName,
		&admin.Role, &permissionsJSON, &admin.Status, &admin.FailedLoginAttempts, &admin.LockedUntil,
		&admin.PasswordChangedAt, &admin.LastLoginAt, &admin.LastLoginIP, &admin.CreatedBy, &admin.CreatedAt, &admin.UpdatedAt)
	if err != nil {
		return admin, err
	}
	if len(permissionsJSON) > 0 {
		if err := json.Unmarshal(permissionsJSON, &admin.Permissions); err != nil {
			return admin, err
		}
	}
	return admin, nil
}

func nullableJSON(value []byte, valid bool) any {
	if !valid {
		return nil
	}
	return string(value)
}

func (s *MySQLStore) ListSystemControls(ctx context.Context) ([]domain.SystemControl, error) {
	rows, err := s.db.QueryContext(ctx, systemControlSelect+` ORDER BY control_group, sort_order, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	controls := make([]domain.SystemControl, 0)
	for rows.Next() {
		control, err := scanSystemControl(rows)
		if err != nil {
			return nil, err
		}
		controls = append(controls, control)
	}
	return controls, rows.Err()
}

func (s *MySQLStore) GetSystemControl(ctx context.Context, id uint64) (domain.SystemControl, error) {
	return scanSystemControl(s.db.QueryRowContext(ctx, systemControlSelect+` WHERE id=?`, id))
}

func (s *MySQLStore) CreateSystemControl(ctx context.Context, control domain.SystemControl) (uint64, error) {
	valueJSON, defaultJSON, err := marshalControlJSON(control)
	if err != nil {
		return 0, err
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO system_controls
		(control_key, control_group, name, description, value_type, value_text, value_json,
		default_value_text, default_value_json, enabled, is_public, sort_order, version,
		effective_from, effective_until, created_by, updated_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?, ?, ?)`, control.ControlKey,
		control.ControlGroup, control.Name, control.Description, control.ValueType, control.ValueText,
		valueJSON, control.DefaultValueText, defaultJSON, control.Enabled, control.IsPublic,
		control.SortOrder, control.EffectiveFrom, control.EffectiveUntil, control.CreatedBy, control.UpdatedBy)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return uint64(id), err
}

func (s *MySQLStore) UpdateSystemControl(ctx context.Context, control domain.SystemControl) error {
	valueJSON, defaultJSON, err := marshalControlJSON(control)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `UPDATE system_controls SET control_key=?, control_group=?, name=?,
		description=?, value_type=?, value_text=?, value_json=?, default_value_text=?, default_value_json=?,
		enabled=?, is_public=?, sort_order=?, version=version+1, effective_from=?, effective_until=?, updated_by=?
		WHERE id=?`, control.ControlKey, control.ControlGroup, control.Name, control.Description,
		control.ValueType, control.ValueText, valueJSON, control.DefaultValueText, defaultJSON,
		control.Enabled, control.IsPublic, control.SortOrder, control.EffectiveFrom, control.EffectiveUntil,
		control.UpdatedBy, control.ID)
	return err
}

func (s *MySQLStore) DeleteSystemControl(ctx context.Context, id uint64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM system_controls WHERE id=?`, id)
	return err
}

const systemControlSelect = `SELECT id, control_key, control_group, name, description, value_type,
	value_text, value_json, default_value_text, default_value_json, enabled, is_public, sort_order,
	version, effective_from, effective_until, created_by, updated_by, created_at, updated_at FROM system_controls`

type systemControlScanner interface{ Scan(dest ...any) error }

func scanSystemControl(scanner systemControlScanner) (domain.SystemControl, error) {
	var control domain.SystemControl
	var valueJSON, defaultJSON []byte
	err := scanner.Scan(&control.ID, &control.ControlKey, &control.ControlGroup, &control.Name,
		&control.Description, &control.ValueType, &control.ValueText, &valueJSON, &control.DefaultValueText,
		&defaultJSON, &control.Enabled, &control.IsPublic, &control.SortOrder, &control.Version,
		&control.EffectiveFrom, &control.EffectiveUntil, &control.CreatedBy, &control.UpdatedBy,
		&control.CreatedAt, &control.UpdatedAt)
	if err != nil {
		return control, err
	}
	if len(valueJSON) > 0 {
		if err := json.Unmarshal(valueJSON, &control.ValueJSON); err != nil {
			return control, err
		}
	}
	if len(defaultJSON) > 0 {
		if err := json.Unmarshal(defaultJSON, &control.DefaultValueJSON); err != nil {
			return control, err
		}
	}
	return control, nil
}

func marshalControlJSON(control domain.SystemControl) (any, any, error) {
	var current, fallback any
	if control.ValueJSON != nil {
		value, err := json.Marshal(control.ValueJSON)
		if err != nil {
			return nil, nil, err
		}
		current = string(value)
	}
	if control.DefaultValueJSON != nil {
		value, err := json.Marshal(control.DefaultValueJSON)
		if err != nil {
			return nil, nil, err
		}
		fallback = string(value)
	}
	return current, fallback, nil
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
		&excellentTime, &normalTime, &timeLimit, &stepLimit,
		&level.ShowSteps, &level.ShowTimer, &level.ShowMismatch, &level.HintHighlightMs, &level.CoinRewardBase, &level.StaminaCost,
		&enabled, &level.Version, &level.UpdatedAt,
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
	if !validLevelGrid(level.Rows, level.Cols, level.PairCount) {
		return fmt.Errorf("rows * cols must equal pairCount * 2, or equal pairCount * 2 + 1 for center empty slot")
	}
	return nil
}

func validLevelGrid(rows int, cols int, pairCount int) bool {
	totalSlots := rows * cols
	cardSlots := pairCount * 2
	return totalSlots == cardSlots || (totalSlots == cardSlots+1 && totalSlots%2 == 1)
}
