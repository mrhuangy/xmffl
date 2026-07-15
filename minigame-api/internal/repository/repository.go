package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"fpxxl/minigame-api/internal/domain"
)

var (
	ErrNotFound            = errors.New("not found")
	ErrInsufficientCoin    = errors.New("insufficient coins")
	ErrInsufficientStamina = errors.New("insufficient stamina")
	ErrProductUnavailable  = errors.New("product unavailable")
)

type Repository interface {
	UpsertPlayer(ctx context.Context, openID string, nickname string, avatarURL string) (domain.Player, domain.PlayerProgress, error)
	PlayerByOpenID(ctx context.Context, openID string) (domain.Player, error)
	Progress(ctx context.Context, playerID uint64) (domain.PlayerProgress, error)
	ListLevels(ctx context.Context) ([]domain.LevelConfig, error)
	AdConfig(ctx context.Context) (domain.AdFrequencyConfig, error)
	StartLevel(ctx context.Context, playerID uint64, levelID int) (domain.PlayerProgress, error)
	SubmitLevelResult(ctx context.Context, player domain.Player, result domain.LevelResult) (domain.PlayerProgress, error)
	ListShopProducts(ctx context.Context) ([]domain.ShopProduct, error)
	PurchaseProduct(ctx context.Context, playerID uint64, productKey string) (domain.PlayerProgress, string, error)
	ListLeaderboard(ctx context.Context, levelID int, limit int) ([]domain.LeaderboardEntry, error)
	SaveEvents(ctx context.Context, playerID uint64, events []domain.GameEvent) error
}

type MySQLRepository struct {
	db *sql.DB
}

func NewMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

func (r *MySQLRepository) UpsertPlayer(ctx context.Context, openID string, nickname string, avatarURL string) (domain.Player, domain.PlayerProgress, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Player{}, domain.PlayerProgress{}, err
	}
	defer rollback(tx)

	_, err = tx.ExecContext(ctx, `INSERT INTO players (open_id, nickname, avatar_url, last_login_at)
		VALUES (?, ?, ?, NOW())
		ON DUPLICATE KEY UPDATE
			nickname = IF(VALUES(nickname) = '', nickname, VALUES(nickname)),
			avatar_url = IF(VALUES(avatar_url) = '', avatar_url, VALUES(avatar_url)),
			last_login_at = NOW()`, openID, nickname, avatarURL)
	if err != nil {
		return domain.Player{}, domain.PlayerProgress{}, err
	}

	player, err := playerByOpenID(ctx, tx, openID)
	if err != nil {
		return domain.Player{}, domain.PlayerProgress{}, err
	}

	_, err = tx.ExecContext(ctx, `INSERT IGNORE INTO player_progress (
		player_id, current_level, coins, stamina, max_stamina, hints, level_stars, completed_levels
	) VALUES (?, 1, 0, 5, 5, 0, JSON_OBJECT(), JSON_ARRAY())`, player.ID)
	if err != nil {
		return domain.Player{}, domain.PlayerProgress{}, err
	}

	progress, err := progressByPlayerID(ctx, tx, player.ID)
	if err != nil {
		return domain.Player{}, domain.PlayerProgress{}, err
	}

	return player, progress, tx.Commit()
}

func (r *MySQLRepository) PlayerByOpenID(ctx context.Context, openID string) (domain.Player, error) {
	return playerByOpenID(ctx, r.db, openID)
}

func (r *MySQLRepository) Progress(ctx context.Context, playerID uint64) (domain.PlayerProgress, error) {
	return progressByPlayerID(ctx, r.db, playerID)
}

func (r *MySQLRepository) ListLevels(ctx context.Context) ([]domain.LevelConfig, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT level_id, rows_count, cols_count, pair_count, mode, theme_id,
		initial_preview_ms, flip_back_delay_ms, level_time_limit_seconds, max_mismatch_count,
		excellent_step_threshold, normal_step_threshold, excellent_time_threshold, normal_time_threshold,
		time_limit_seconds, step_limit, version, updated_at
		FROM level_configs WHERE enabled = 1 ORDER BY level_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	levels := make([]domain.LevelConfig, 0)
	for rows.Next() {
		var level domain.LevelConfig
		var excellentTime, normalTime, timeLimit, stepLimit sql.NullInt64
		err := rows.Scan(
			&level.LevelID, &level.Rows, &level.Cols, &level.PairCount, &level.Mode, &level.ThemeID,
			&level.InitialPreviewMs, &level.FlipBackDelayMs, &level.LevelTimeLimitSeconds, &level.MaxMismatchCount,
			&level.ExcellentStepThreshold, &level.NormalStepThreshold, &excellentTime, &normalTime,
			&timeLimit, &stepLimit, &level.Version, &level.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		level.ExcellentTimeThreshold = nullableInt(excellentTime)
		level.NormalTimeThreshold = nullableInt(normalTime)
		level.TimeLimitSeconds = nullableInt(timeLimit)
		level.StepLimit = nullableInt(stepLimit)
		levels = append(levels, level)
	}
	return levels, rows.Err()
}

func (r *MySQLRepository) AdConfig(ctx context.Context) (domain.AdFrequencyConfig, error) {
	var cfg domain.AdFrequencyConfig
	var scenesJSON string
	var updatedAt time.Time
	err := r.db.QueryRowContext(ctx, `SELECT no_interstitial_before_level, interstitial_every_levels,
		max_interstitial_per_day, max_revive_per_level, banner_enabled_scenes, version, updated_at
		FROM ad_frequency_configs WHERE id = 1`).Scan(
		&cfg.NoInterstitialBeforeLevel, &cfg.InterstitialEveryLevels, &cfg.MaxInterstitialPerDay,
		&cfg.MaxRevivePerLevel, &scenesJSON, &cfg.Version, &updatedAt,
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

func (r *MySQLRepository) StartLevel(ctx context.Context, playerID uint64, levelID int) (domain.PlayerProgress, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.PlayerProgress{}, err
	}
	defer rollback(tx)

	progress, err := progressForUpdate(ctx, tx, playerID)
	if err != nil {
		return domain.PlayerProgress{}, err
	}
	if progress.Stamina <= 0 {
		return domain.PlayerProgress{}, ErrInsufficientStamina
	}
	progress.Stamina--

	_, err = tx.ExecContext(ctx, `UPDATE player_progress SET stamina = ?, updated_at = NOW() WHERE player_id = ?`, progress.Stamina, playerID)
	if err != nil {
		return domain.PlayerProgress{}, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO stamina_transactions (
		transaction_no, player_id, change_amount, balance_after, reason, ref_type, ref_id
	) VALUES (?, ?, -1, ?, 'level_start', 'level', ?)`, makeID("stamina"), playerID, progress.Stamina, strconv.Itoa(levelID))
	if err != nil {
		return domain.PlayerProgress{}, err
	}

	progress, err = progressByPlayerID(ctx, tx, playerID)
	if err != nil {
		return domain.PlayerProgress{}, err
	}
	return progress, tx.Commit()
}

func (r *MySQLRepository) SubmitLevelResult(ctx context.Context, player domain.Player, result domain.LevelResult) (domain.PlayerProgress, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.PlayerProgress{}, err
	}
	defer rollback(tx)

	progress, err := progressForUpdate(ctx, tx, player.ID)
	if err != nil {
		return domain.PlayerProgress{}, err
	}

	res, err := tx.ExecContext(ctx, `INSERT INTO level_results (
		player_id, level_id, success, reason, steps, mismatch_count, elapsed_ms, stars, coins_earned, used_hints
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		player.ID, result.LevelID, result.Success, normalizeReason(result.Reason), result.Steps, result.MismatchCount,
		result.ElapsedMs, result.Stars, result.CoinsEarned, result.UsedHints,
	)
	if err != nil {
		return domain.PlayerProgress{}, err
	}
	levelResultID, _ := res.LastInsertId()

	if result.Success {
		progress.CurrentLevel = max(progress.CurrentLevel, result.LevelID+1)
		progress.LevelStars[strconv.Itoa(result.LevelID)] = max(progress.LevelStars[strconv.Itoa(result.LevelID)], result.Stars)
		progress.CompletedLevels = appendUnique(progress.CompletedLevels, result.LevelID)
		if result.CoinsEarned > 0 {
			progress.Coins += result.CoinsEarned
			_, err = tx.ExecContext(ctx, `INSERT INTO coin_transactions (
				transaction_no, player_id, change_amount, balance_after, reason, ref_type, ref_id
			) VALUES (?, ?, ?, ?, 'level_complete', 'level_result', ?)`,
				makeID("coin"), player.ID, result.CoinsEarned, progress.Coins, strconv.FormatInt(levelResultID, 10),
			)
			if err != nil {
				return domain.PlayerProgress{}, err
			}
			_, err = tx.ExecContext(ctx, `INSERT INTO reward_grants (
				reward_id, player_id, source, source_ref, reward_type, amount, level_id
			) VALUES (?, ?, 'level_complete', ?, 'coins', ?, ?)`,
				makeID("reward"), player.ID, strconv.FormatInt(levelResultID, 10), result.CoinsEarned, result.LevelID,
			)
			if err != nil {
				return domain.PlayerProgress{}, err
			}
		}

		levelStarsJSON, err := json.Marshal(progress.LevelStars)
		if err != nil {
			return domain.PlayerProgress{}, err
		}
		completedJSON, err := json.Marshal(progress.CompletedLevels)
		if err != nil {
			return domain.PlayerProgress{}, err
		}
		_, err = tx.ExecContext(ctx, `UPDATE player_progress
			SET current_level = ?, coins = ?, level_stars = ?, completed_levels = ?, updated_at = NOW()
			WHERE player_id = ?`, progress.CurrentLevel, progress.Coins, string(levelStarsJSON), string(completedJSON), player.ID)
		if err != nil {
			return domain.PlayerProgress{}, err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO leaderboard_entries (
			player_id, level_id, nickname, stars, steps, elapsed_ms
		) VALUES (?, ?, ?, ?, ?, ?)`, player.ID, result.LevelID, player.Nickname, result.Stars, result.Steps, result.ElapsedMs)
		if err != nil {
			return domain.PlayerProgress{}, err
		}
	}

	progress, err = progressByPlayerID(ctx, tx, player.ID)
	if err != nil {
		return domain.PlayerProgress{}, err
	}
	return progress, tx.Commit()
}

func (r *MySQLRepository) ListShopProducts(ctx context.Context) ([]domain.ShopProduct, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, product_key, name, product_type, currency_type,
		currency_amount, grant_type, grant_amount, daily_buy_limit, sort_order, updated_at
		FROM shop_products WHERE enabled = 1 ORDER BY sort_order, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]domain.ShopProduct, 0)
	for rows.Next() {
		var product domain.ShopProduct
		var dailyLimit sql.NullInt64
		if err := rows.Scan(&product.ID, &product.ProductKey, &product.Name, &product.ProductType, &product.CurrencyType,
			&product.CurrencyAmount, &product.GrantType, &product.GrantAmount, &dailyLimit, &product.SortOrder, &product.UpdatedAt); err != nil {
			return nil, err
		}
		product.DailyBuyLimit = nullableInt(dailyLimit)
		products = append(products, product)
	}
	return products, rows.Err()
}

func (r *MySQLRepository) PurchaseProduct(ctx context.Context, playerID uint64, productKey string) (domain.PlayerProgress, string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.PlayerProgress{}, "", err
	}
	defer rollback(tx)

	product, err := productByKeyForUpdate(ctx, tx, productKey)
	if err != nil {
		return domain.PlayerProgress{}, "", err
	}
	if product.CurrencyType != "coins" {
		return domain.PlayerProgress{}, "", ErrProductUnavailable
	}

	progress, err := progressForUpdate(ctx, tx, playerID)
	if err != nil {
		return domain.PlayerProgress{}, "", err
	}
	if progress.Coins < product.CurrencyAmount {
		return domain.PlayerProgress{}, "", ErrInsufficientCoin
	}

	orderNo := makeID("order")
	_, err = tx.ExecContext(ctx, `INSERT INTO purchase_orders (
		order_no, player_id, product_id, product_key, product_name, currency_type, currency_amount,
		grant_type, grant_amount, status, paid_at, fulfilled_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'fulfilled', NOW(), NOW())`,
		orderNo, playerID, product.ID, product.ProductKey, product.Name, product.CurrencyType, product.CurrencyAmount,
		product.GrantType, product.GrantAmount,
	)
	if err != nil {
		return domain.PlayerProgress{}, "", err
	}

	progress.Coins -= product.CurrencyAmount
	_, err = tx.ExecContext(ctx, `INSERT INTO coin_transactions (
		transaction_no, player_id, change_amount, balance_after, reason, ref_type, ref_id
	) VALUES (?, ?, ?, ?, 'shop_purchase', 'purchase_order', ?)`,
		makeID("coin"), playerID, -product.CurrencyAmount, progress.Coins, orderNo,
	)
	if err != nil {
		return domain.PlayerProgress{}, "", err
	}

	switch product.GrantType {
	case "stamina":
		progress.Stamina += product.GrantAmount
		_, err = tx.ExecContext(ctx, `INSERT INTO stamina_transactions (
			transaction_no, player_id, change_amount, balance_after, reason, ref_type, ref_id
		) VALUES (?, ?, ?, ?, 'coin_exchange', 'purchase_order', ?)`,
			makeID("stamina"), playerID, product.GrantAmount, progress.Stamina, orderNo,
		)
	case "hints":
		progress.Hints += product.GrantAmount
	case "coins":
		progress.Coins += product.GrantAmount
	default:
		return domain.PlayerProgress{}, "", ErrProductUnavailable
	}
	if err != nil {
		return domain.PlayerProgress{}, "", err
	}

	_, err = tx.ExecContext(ctx, `UPDATE player_progress
		SET coins = ?, stamina = ?, hints = ?, updated_at = NOW()
		WHERE player_id = ?`, progress.Coins, progress.Stamina, progress.Hints, playerID)
	if err != nil {
		return domain.PlayerProgress{}, "", err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO reward_grants (
		reward_id, player_id, source, source_ref, reward_type, amount
	) VALUES (?, ?, 'shop', ?, ?, ?)`,
		makeID("reward"), playerID, orderNo, product.GrantType, product.GrantAmount,
	)
	if err != nil {
		return domain.PlayerProgress{}, "", err
	}

	progress, err = progressByPlayerID(ctx, tx, playerID)
	if err != nil {
		return domain.PlayerProgress{}, "", err
	}
	return progress, orderNo, tx.Commit()
}

func (r *MySQLRepository) ListLeaderboard(ctx context.Context, levelID int, limit int) ([]domain.LeaderboardEntry, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT p.open_id, le.nickname, le.level_id, le.stars, le.steps, le.elapsed_ms, le.submitted_at
		FROM leaderboard_entries le
		JOIN players p ON p.id = le.player_id
		WHERE (? = 0 OR le.level_id = ?)
		ORDER BY le.stars DESC, le.steps ASC, le.elapsed_ms ASC, le.submitted_at ASC
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

func (r *MySQLRepository) SaveEvents(ctx context.Context, playerID uint64, events []domain.GameEvent) error {
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
		args = append(args, event.EventID, playerID, event.EventType, event.LevelID, string(payload))
	}
	_, err := r.db.ExecContext(ctx, "INSERT INTO game_events (event_id, player_id, event_type, level_id, payload) VALUES "+strings.Join(parts, ","), args...)
	return err
}

type queryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func playerByOpenID(ctx context.Context, q queryer, openID string) (domain.Player, error) {
	var player domain.Player
	var unionID sql.NullString
	var lastLoginAt sql.NullTime
	err := q.QueryRowContext(ctx, `SELECT id, open_id, union_id, nickname, avatar_url, status, last_login_at, created_at, updated_at
		FROM players WHERE open_id = ? AND status <> 'deleted'`, openID).Scan(
		&player.ID, &player.OpenID, &unionID, &player.Nickname, &player.AvatarURL, &player.Status,
		&lastLoginAt, &player.CreatedAt, &player.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return player, ErrNotFound
	}
	if err != nil {
		return player, err
	}
	if unionID.Valid {
		player.UnionID = &unionID.String
	}
	if lastLoginAt.Valid {
		player.LastLoginAt = &lastLoginAt.Time
	}
	return player, nil
}

func progressByPlayerID(ctx context.Context, q queryer, playerID uint64) (domain.PlayerProgress, error) {
	var progress domain.PlayerProgress
	var levelStarsJSON, completedLevelsJSON string
	var nextRecover sql.NullTime
	err := q.QueryRowContext(ctx, `SELECT player_id, current_level, coins, stamina, max_stamina, hints,
		level_stars, completed_levels, next_stamina_recover_at, updated_at
		FROM player_progress WHERE player_id = ?`, playerID).Scan(
		&progress.PlayerID, &progress.CurrentLevel, &progress.Coins, &progress.Stamina, &progress.MaxStamina,
		&progress.Hints, &levelStarsJSON, &completedLevelsJSON, &nextRecover, &progress.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return progress, ErrNotFound
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
	if nextRecover.Valid {
		progress.NextStaminaRecoverAt = &nextRecover.Time
	}
	return progress, nil
}

func progressForUpdate(ctx context.Context, tx *sql.Tx, playerID uint64) (domain.PlayerProgress, error) {
	var progress domain.PlayerProgress
	var levelStarsJSON, completedLevelsJSON string
	var nextRecover sql.NullTime
	err := tx.QueryRowContext(ctx, `SELECT player_id, current_level, coins, stamina, max_stamina, hints,
		level_stars, completed_levels, next_stamina_recover_at, updated_at
		FROM player_progress WHERE player_id = ? FOR UPDATE`, playerID).Scan(
		&progress.PlayerID, &progress.CurrentLevel, &progress.Coins, &progress.Stamina, &progress.MaxStamina,
		&progress.Hints, &levelStarsJSON, &completedLevelsJSON, &nextRecover, &progress.UpdatedAt,
	)
	if err != nil {
		return progress, err
	}
	if err := json.Unmarshal([]byte(levelStarsJSON), &progress.LevelStars); err != nil {
		return progress, err
	}
	if err := json.Unmarshal([]byte(completedLevelsJSON), &progress.CompletedLevels); err != nil {
		return progress, err
	}
	if nextRecover.Valid {
		progress.NextStaminaRecoverAt = &nextRecover.Time
	}
	return progress, nil
}

func productByKeyForUpdate(ctx context.Context, tx *sql.Tx, productKey string) (domain.ShopProduct, error) {
	var product domain.ShopProduct
	var dailyLimit sql.NullInt64
	err := tx.QueryRowContext(ctx, `SELECT id, product_key, name, product_type, currency_type,
		currency_amount, grant_type, grant_amount, daily_buy_limit, sort_order, updated_at
		FROM shop_products WHERE product_key = ? AND enabled = 1 FOR UPDATE`, productKey).Scan(
		&product.ID, &product.ProductKey, &product.Name, &product.ProductType, &product.CurrencyType,
		&product.CurrencyAmount, &product.GrantType, &product.GrantAmount, &dailyLimit, &product.SortOrder, &product.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return product, ErrProductUnavailable
	}
	if err != nil {
		return product, err
	}
	product.DailyBuyLimit = nullableInt(dailyLimit)
	return product, nil
}

func rollback(tx *sql.Tx) {
	_ = tx.Rollback()
}

func nullableInt(value sql.NullInt64) *int {
	if !value.Valid {
		return nil
	}
	v := int(value.Int64)
	return &v
}

func normalizeReason(reason string) string {
	switch reason {
	case "completed", "time_out", "mismatch_limit", "quit":
		return reason
	default:
		return "unknown"
	}
}

func appendUnique(values []int, value int) []int {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func makeID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}
