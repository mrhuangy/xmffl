package repository

import (
	"context"
	cryptorand "crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"fpxxl/minigame-api/internal/domain"
)

const (
	initialPlayerCoins     = 100
	initialToolCharges     = 3
	staminaRecoverInterval = 2 * time.Minute
	staminaRecoverReason   = "auto_recover"
	staminaRecoverRefType  = "system"
	staminaRecoverRefID    = "natural_recovery"
)

var (
	ErrNotFound            = errors.New("not found")
	ErrInsufficientCoin    = errors.New("insufficient coins")
	ErrInsufficientTool    = errors.New("insufficient tool charges")
	ErrInsufficientStamina = errors.New("insufficient stamina")
	ErrProductUnavailable  = errors.New("product unavailable")
)

type Repository interface {
	UpsertPlayer(ctx context.Context, openID string, unionID *string, nickname string, avatarURL string) (domain.Player, domain.PlayerProgress, error)
	PlayerByOpenID(ctx context.Context, openID string) (domain.Player, error)
	Progress(ctx context.Context, playerID uint64) (domain.PlayerProgress, error)
	ChangeToolCount(ctx context.Context, playerID uint64, toolType string, delta int, source string) (domain.PlayerProgress, error)
	ListLevels(ctx context.Context) ([]domain.LevelConfig, error)
	AdConfig(ctx context.Context) (domain.AdFrequencyConfig, error)
	PublicSystemControls(ctx context.Context) (domain.PublicSystemControls, error)
	SystemControlBool(ctx context.Context, key string, fallback bool) (bool, error)
	StartLevel(ctx context.Context, playerID uint64, levelID int) (domain.PlayerProgress, error)
	SubmitLevelResult(ctx context.Context, player domain.Player, result domain.LevelResult) (domain.PlayerProgress, domain.LevelResultRewards, error)
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

func (r *MySQLRepository) UpsertPlayer(ctx context.Context, openID string, unionID *string, nickname string, avatarURL string) (domain.Player, domain.PlayerProgress, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Player{}, domain.PlayerProgress{}, err
	}
	defer rollback(tx)

	player, err := playerByOpenID(ctx, tx, openID)
	isNewPlayer := false
	if errors.Is(err, ErrNotFound) {
		isNewPlayer = true
		nickname, err = r.uniqueNickname(ctx, tx)
		if err != nil {
			return domain.Player{}, domain.PlayerProgress{}, err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO players (open_id, union_id, nickname, avatar_url, last_login_at)
			VALUES (?, ?, ?, ?, NOW())`, openID, unionID, nickname, avatarURL)
		if err != nil {
			return domain.Player{}, domain.PlayerProgress{}, err
		}
		player, err = playerByOpenID(ctx, tx, openID)
	}
	if err != nil {
		return domain.Player{}, domain.PlayerProgress{}, err
	}

	if !isNewPlayer {
		_, err = tx.ExecContext(ctx, `UPDATE players
			SET union_id = IF(? IS NULL OR ? = '', union_id, ?),
				avatar_url = IF(? = '', avatar_url, ?),
				last_login_at = NOW()
			WHERE id = ?`, unionID, unionID, unionID, avatarURL, avatarURL, player.ID)
		if err != nil {
			return domain.Player{}, domain.PlayerProgress{}, err
		}
		player, err = playerByOpenID(ctx, tx, openID)
		if err != nil {
			return domain.Player{}, domain.PlayerProgress{}, err
		}
	}

	result, err := tx.ExecContext(ctx, `INSERT IGNORE INTO player_progress (
		player_id, current_level, coins, stamina, max_stamina, hints, preview_again_count, remove_pair_count, level_stars, completed_levels
	) VALUES (?, 1, ?, 5, 5, ?, ?, ?, JSON_OBJECT(), JSON_ARRAY())`, player.ID, initialPlayerCoins, initialToolCharges, initialToolCharges, initialToolCharges)
	if err != nil {
		return domain.Player{}, domain.PlayerProgress{}, err
	}
	if affected, _ := result.RowsAffected(); affected > 0 {
		_, err = tx.ExecContext(ctx, `INSERT INTO coin_transactions (
			transaction_no, player_id, change_amount, balance_after, reason, ref_type, ref_id, note
		) VALUES (?, ?, ?, ?, 'activity_reward', 'register', ?, '新玩家注册初始金币')`,
			makeID("coin"), player.ID, initialPlayerCoins, initialPlayerCoins, player.OpenID,
		)
		if err != nil {
			return domain.Player{}, domain.PlayerProgress{}, err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO reward_grants (
			reward_id, player_id, source, source_ref, reward_type, amount
		) VALUES (?, ?, 'activity', ?, 'coins', ?)`,
			makeID("reward"), player.ID, "register", initialPlayerCoins,
		)
		if err != nil {
			return domain.Player{}, domain.PlayerProgress{}, err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO tool_transactions (
			transaction_no, player_id, tool_type, change_amount, balance_after, source, ref_type, ref_id, note
		) VALUES
			(?, ?, 'hint', ?, ?, 'register', 'player', ?, 'initial tool charges'),
			(?, ?, 'preview_again', ?, ?, 'register', 'player', ?, 'initial tool charges'),
			(?, ?, 'remove_pair', ?, ?, 'register', 'player', ?, 'initial tool charges')`,
			makeID("tool"), player.ID, initialToolCharges, initialToolCharges, player.OpenID,
			makeID("tool"), player.ID, initialToolCharges, initialToolCharges, player.OpenID,
			makeID("tool"), player.ID, initialToolCharges, initialToolCharges, player.OpenID,
		)
		if err != nil {
			return domain.Player{}, domain.PlayerProgress{}, err
		}
	}

	progress, err := settleStaminaRecovery(ctx, tx, player.ID)
	if err != nil {
		return domain.Player{}, domain.PlayerProgress{}, err
	}

	return player, progress, tx.Commit()
}

func (r *MySQLRepository) uniqueNickname(ctx context.Context, q queryer) (string, error) {
	for i := 0; i < 80; i++ {
		nickname, err := randomChineseNickname()
		if err != nil {
			return "", err
		}
		exists, err := nicknameExists(ctx, q, nickname)
		if err != nil {
			return "", err
		}
		if !exists {
			return nickname, nil
		}
	}
	return fmt.Sprintf("玩家%d", time.Now().UnixNano()%100000000), nil
}

func (r *MySQLRepository) PlayerByOpenID(ctx context.Context, openID string) (domain.Player, error) {
	return playerByOpenID(ctx, r.db, openID)
}

func (r *MySQLRepository) Progress(ctx context.Context, playerID uint64) (domain.PlayerProgress, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.PlayerProgress{}, err
	}
	defer rollback(tx)

	progress, err := settleStaminaRecovery(ctx, tx, playerID)
	if err != nil {
		return domain.PlayerProgress{}, err
	}
	return progress, tx.Commit()
}

func (r *MySQLRepository) ChangeToolCount(ctx context.Context, playerID uint64, toolType string, delta int, source string) (domain.PlayerProgress, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.PlayerProgress{}, err
	}
	defer rollback(tx)

	progress, err := progressForUpdate(ctx, tx, playerID)
	if err != nil {
		return domain.PlayerProgress{}, err
	}

	balanceAfter := 0
	var toolColumn string
	switch toolType {
	case "hint":
		balanceAfter = progress.Hints + delta
		toolColumn = "hints"
	case "preview_again":
		balanceAfter = progress.PreviewAgainCount + delta
		toolColumn = "preview_again_count"
	case "remove_pair":
		balanceAfter = progress.RemovePairCount + delta
		toolColumn = "remove_pair_count"
	default:
		return domain.PlayerProgress{}, ErrNotFound
	}
	if balanceAfter < 0 {
		return domain.PlayerProgress{}, ErrInsufficientTool
	}

	switch toolType {
	case "hint":
		progress.Hints = balanceAfter
	case "preview_again":
		progress.PreviewAgainCount = balanceAfter
	case "remove_pair":
		progress.RemovePairCount = balanceAfter
	}

	note := "tool change"
	_, err = tx.ExecContext(ctx, `INSERT INTO tool_transactions (
		transaction_no, player_id, tool_type, change_amount, balance_after, source, ref_type, ref_id, note
	) VALUES (?, ?, ?, ?, ?, ?, 'system', '', ?)`,
		makeID("tool"), playerID, toolType, delta, balanceAfter, source, note,
	)
	if err != nil {
		return domain.PlayerProgress{}, err
	}

	_, err = tx.ExecContext(ctx, `UPDATE player_progress SET `+toolColumn+` = ?, updated_at = NOW() WHERE player_id = ?`, balanceAfter, playerID)
	if err != nil {
		return domain.PlayerProgress{}, err
	}

	progress, err = progressByPlayerID(ctx, tx, playerID)
	if err != nil {
		return domain.PlayerProgress{}, err
	}
	return progress, tx.Commit()
}

func (r *MySQLRepository) ListLevels(ctx context.Context) ([]domain.LevelConfig, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT level_id, rows_count, cols_count, pair_count, mode, theme_id,
		initial_preview_ms, flip_back_delay_ms, level_time_limit_seconds, max_mismatch_count,
		show_steps, show_timer, show_mismatch, hint_highlight_ms, coin_reward_base, coin_reward_star1, coin_reward_star2, coin_reward_star3, stamina_cost,
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
			&level.ShowSteps, &level.ShowTimer, &level.ShowMismatch, &level.HintHighlightMs, &level.CoinRewardBase, &level.CoinRewardStar1, &level.CoinRewardStar2, &level.CoinRewardStar3, &level.StaminaCost,
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

func (r *MySQLRepository) PublicSystemControls(ctx context.Context) (domain.PublicSystemControls, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT control_key, value_type, value_text, value_json
		FROM system_controls
		WHERE enabled = 1
			AND is_public = 1
			AND (effective_from IS NULL OR effective_from <= NOW())
			AND (effective_until IS NULL OR effective_until >= NOW())
		ORDER BY sort_order, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	controls := domain.PublicSystemControls{}
	for rows.Next() {
		var key, valueType string
		var valueText, valueJSON sql.NullString
		if err := rows.Scan(&key, &valueType, &valueText, &valueJSON); err != nil {
			return nil, err
		}
		value, err := systemControlValue(valueType, valueText, valueJSON)
		if err != nil {
			return nil, err
		}
		controls[key] = value
	}
	return controls, rows.Err()
}

func (r *MySQLRepository) SystemControlBool(ctx context.Context, key string, fallback bool) (bool, error) {
	var valueType string
	var valueText, valueJSON sql.NullString
	err := r.db.QueryRowContext(ctx, `SELECT value_type, value_text, value_json
		FROM system_controls
		WHERE control_key = ?
			AND enabled = 1
			AND (effective_from IS NULL OR effective_from <= NOW())
			AND (effective_until IS NULL OR effective_until >= NOW())`, key).Scan(&valueType, &valueText, &valueJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return fallback, nil
	}
	if err != nil {
		return fallback, err
	}

	value, err := systemControlValue(valueType, valueText, valueJSON)
	if err != nil {
		return fallback, err
	}
	boolValue, ok := value.(bool)
	if !ok {
		return fallback, nil
	}
	return boolValue, nil
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
	progress, err = applyStaminaRecovery(ctx, tx, progress)
	if err != nil {
		return domain.PlayerProgress{}, err
	}
	unlimitedStamina, err := systemControlBoolTx(ctx, tx, "game.unlimited_stamina", false)
	if err != nil {
		return domain.PlayerProgress{}, err
	}
	if unlimitedStamina {
		return progress, tx.Commit()
	}
	if progress.Stamina <= 0 {
		return domain.PlayerProgress{}, ErrInsufficientStamina
	}
	progress.Stamina--
	if progress.Stamina < progress.MaxStamina && progress.NextStaminaRecoverAt == nil {
		next := time.Now().Add(staminaRecoverInterval)
		progress.NextStaminaRecoverAt = &next
	}

	_, err = tx.ExecContext(ctx, `UPDATE player_progress
		SET stamina = ?, next_stamina_recover_at = ?, updated_at = NOW()
		WHERE player_id = ?`, progress.Stamina, nullableTime(progress.NextStaminaRecoverAt), playerID)
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

func (r *MySQLRepository) SubmitLevelResult(ctx context.Context, player domain.Player, result domain.LevelResult) (domain.PlayerProgress, domain.LevelResultRewards, error) {
	var rewards domain.LevelResultRewards
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.PlayerProgress{}, rewards, err
	}
	defer rollback(tx)

	progress, err := progressForUpdate(ctx, tx, player.ID)
	if err != nil {
		return domain.PlayerProgress{}, rewards, err
	}
	level, err := levelByID(ctx, tx, result.LevelID)
	if err != nil {
		return domain.PlayerProgress{}, rewards, err
	}
	previousBestStars := progress.LevelStars[strconv.Itoa(result.LevelID)]
	result.Success = result.Success && normalizeReason(result.Reason) == "completed"
	if result.Success {
		result.Stars = calculateResultStars(result, level)
		result.CoinsEarned = coinRewardForStars(level, result.Stars)
	} else {
		result.Stars = 0
		result.CoinsEarned = 0
	}

	res, err := tx.ExecContext(ctx, `INSERT INTO level_results (
		player_id, level_id, success, reason, steps, mismatch_count, elapsed_ms, stars, coins_earned, used_hints
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		player.ID, result.LevelID, result.Success, normalizeReason(result.Reason), result.Steps, result.MismatchCount,
		result.ElapsedMs, result.Stars, result.CoinsEarned, result.UsedHints,
	)
	if err != nil {
		return domain.PlayerProgress{}, rewards, err
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
				return domain.PlayerProgress{}, rewards, err
			}
			_, err = tx.ExecContext(ctx, `INSERT INTO reward_grants (
				reward_id, player_id, source, source_ref, reward_type, amount, level_id
			) VALUES (?, ?, 'level_complete', ?, 'coins', ?, ?)`,
				makeID("reward"), player.ID, strconv.FormatInt(levelResultID, 10), result.CoinsEarned, result.LevelID,
			)
			if err != nil {
				return domain.PlayerProgress{}, rewards, err
			}
		}

		levelStarsJSON, err := json.Marshal(progress.LevelStars)
		if err != nil {
			return domain.PlayerProgress{}, rewards, err
		}
		completedJSON, err := json.Marshal(progress.CompletedLevels)
		if err != nil {
			return domain.PlayerProgress{}, rewards, err
		}
		if result.Stars == 3 && previousBestStars < 3 {
			rewards.Stamina = 1
			progress.Stamina += rewards.Stamina
			if progress.Stamina >= progress.MaxStamina {
				progress.NextStaminaRecoverAt = nil
			}
			_, err = tx.ExecContext(ctx, `INSERT INTO stamina_transactions (
				transaction_no, player_id, change_amount, balance_after, reason, ref_type, ref_id, note
			) VALUES (?, ?, ?, ?, 'activity_reward', 'level_result', ?, 'first_3_star')`,
				makeID("stamina"), player.ID, rewards.Stamina, progress.Stamina, strconv.FormatInt(levelResultID, 10),
			)
			if err != nil {
				return domain.PlayerProgress{}, rewards, err
			}
			_, err = tx.ExecContext(ctx, `INSERT INTO reward_grants (
				reward_id, player_id, source, source_ref, reward_type, amount, level_id
			) VALUES (?, ?, 'activity', ?, 'stamina', ?, ?)`,
				makeID("reward"), player.ID, strconv.FormatInt(levelResultID, 10), rewards.Stamina, result.LevelID,
			)
			if err != nil {
				return domain.PlayerProgress{}, rewards, err
			}
		}
		_, err = tx.ExecContext(ctx, `UPDATE player_progress
			SET current_level = ?, coins = ?, stamina = ?, next_stamina_recover_at = ?, level_stars = ?, completed_levels = ?, updated_at = NOW()
			WHERE player_id = ?`, progress.CurrentLevel, progress.Coins, progress.Stamina, nullableTime(progress.NextStaminaRecoverAt), string(levelStarsJSON), string(completedJSON), player.ID)
		if err != nil {
			return domain.PlayerProgress{}, rewards, err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO leaderboard_entries (
			player_id, level_id, nickname, stars, steps, elapsed_ms
		) VALUES (?, ?, ?, ?, ?, ?)`, player.ID, result.LevelID, player.Nickname, result.Stars, result.Steps, result.ElapsedMs)
		if err != nil {
			return domain.PlayerProgress{}, rewards, err
		}
	}

	progress, err = progressByPlayerID(ctx, tx, player.ID)
	if err != nil {
		return domain.PlayerProgress{}, rewards, err
	}
	return progress, rewards, tx.Commit()
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
	progress, err = applyStaminaRecovery(ctx, tx, progress)
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
		if progress.Stamina >= progress.MaxStamina {
			progress.NextStaminaRecoverAt = nil
		} else if progress.NextStaminaRecoverAt == nil {
			next := time.Now().Add(staminaRecoverInterval)
			progress.NextStaminaRecoverAt = &next
		}
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
		SET coins = ?, stamina = ?, hints = ?, next_stamina_recover_at = ?, updated_at = NOW()
		WHERE player_id = ?`, progress.Coins, progress.Stamina, progress.Hints, nullableTime(progress.NextStaminaRecoverAt), playerID)
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

func nicknameExists(ctx context.Context, q queryer, nickname string) (bool, error) {
	var count int
	err := q.QueryRowContext(ctx, `SELECT COUNT(*) FROM players WHERE nickname = ?`, nickname).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
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
		preview_again_count, remove_pair_count, level_stars, completed_levels, next_stamina_recover_at, updated_at
		FROM player_progress WHERE player_id = ?`, playerID).Scan(
		&progress.PlayerID, &progress.CurrentLevel, &progress.Coins, &progress.Stamina, &progress.MaxStamina,
		&progress.Hints, &progress.PreviewAgainCount, &progress.RemovePairCount, &levelStarsJSON, &completedLevelsJSON, &nextRecover, &progress.UpdatedAt,
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
		preview_again_count, remove_pair_count, level_stars, completed_levels, next_stamina_recover_at, updated_at
		FROM player_progress WHERE player_id = ? FOR UPDATE`, playerID).Scan(
		&progress.PlayerID, &progress.CurrentLevel, &progress.Coins, &progress.Stamina, &progress.MaxStamina,
		&progress.Hints, &progress.PreviewAgainCount, &progress.RemovePairCount, &levelStarsJSON, &completedLevelsJSON, &nextRecover, &progress.UpdatedAt,
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

func settleStaminaRecovery(ctx context.Context, tx *sql.Tx, playerID uint64) (domain.PlayerProgress, error) {
	progress, err := progressForUpdate(ctx, tx, playerID)
	if err != nil {
		return domain.PlayerProgress{}, err
	}
	return applyStaminaRecovery(ctx, tx, progress)
}

func applyStaminaRecovery(ctx context.Context, tx *sql.Tx, progress domain.PlayerProgress) (domain.PlayerProgress, error) {
	now := time.Now()
	if progress.Stamina >= progress.MaxStamina {
		if progress.NextStaminaRecoverAt != nil {
			_, err := tx.ExecContext(ctx, `UPDATE player_progress
				SET next_stamina_recover_at = NULL, updated_at = NOW()
				WHERE player_id = ?`, progress.PlayerID)
			if err != nil {
				return domain.PlayerProgress{}, err
			}
			progress.NextStaminaRecoverAt = nil
			progress.UpdatedAt = now
		}
		return progress, nil
	}

	if progress.NextStaminaRecoverAt == nil {
		next := now.Add(staminaRecoverInterval)
		_, err := tx.ExecContext(ctx, `UPDATE player_progress
			SET next_stamina_recover_at = ?, updated_at = NOW()
			WHERE player_id = ?`, next, progress.PlayerID)
		if err != nil {
			return domain.PlayerProgress{}, err
		}
		progress.NextStaminaRecoverAt = &next
		progress.UpdatedAt = now
		return progress, nil
	}

	next := *progress.NextStaminaRecoverAt
	if now.Before(next) {
		return progress, nil
	}

	recoverCount := 1 + int(now.Sub(next)/staminaRecoverInterval)
	needCount := progress.MaxStamina - progress.Stamina
	if recoverCount > needCount {
		recoverCount = needCount
	}
	progress.Stamina += recoverCount
	if progress.Stamina >= progress.MaxStamina {
		progress.NextStaminaRecoverAt = nil
	} else {
		next = next.Add(time.Duration(recoverCount) * staminaRecoverInterval)
		progress.NextStaminaRecoverAt = &next
	}

	_, err := tx.ExecContext(ctx, `UPDATE player_progress
		SET stamina = ?, next_stamina_recover_at = ?, updated_at = NOW()
		WHERE player_id = ?`, progress.Stamina, nullableTime(progress.NextStaminaRecoverAt), progress.PlayerID)
	if err != nil {
		return domain.PlayerProgress{}, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO stamina_transactions (
		transaction_no, player_id, change_amount, balance_after, reason, ref_type, ref_id, note
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		makeID("stamina"), progress.PlayerID, recoverCount, progress.Stamina,
		staminaRecoverReason, staminaRecoverRefType, staminaRecoverRefID, "natural stamina recovery",
	)
	if err != nil {
		return domain.PlayerProgress{}, err
	}
	progress.UpdatedAt = now
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

func levelByID(ctx context.Context, q queryer, levelID int) (domain.LevelConfig, error) {
	var level domain.LevelConfig
	var excellentTime, normalTime, timeLimit, stepLimit sql.NullInt64
	err := q.QueryRowContext(ctx, `SELECT level_id, rows_count, cols_count, pair_count, mode, theme_id,
		initial_preview_ms, flip_back_delay_ms, level_time_limit_seconds, max_mismatch_count,
		show_steps, show_timer, show_mismatch, hint_highlight_ms, coin_reward_base, coin_reward_star1, coin_reward_star2, coin_reward_star3, stamina_cost,
		excellent_step_threshold, normal_step_threshold, excellent_time_threshold, normal_time_threshold,
		time_limit_seconds, step_limit, version, updated_at
		FROM level_configs WHERE level_id = ? AND enabled = 1`, levelID).Scan(
		&level.LevelID, &level.Rows, &level.Cols, &level.PairCount, &level.Mode, &level.ThemeID,
		&level.InitialPreviewMs, &level.FlipBackDelayMs, &level.LevelTimeLimitSeconds, &level.MaxMismatchCount,
		&level.ShowSteps, &level.ShowTimer, &level.ShowMismatch, &level.HintHighlightMs, &level.CoinRewardBase, &level.CoinRewardStar1, &level.CoinRewardStar2, &level.CoinRewardStar3, &level.StaminaCost,
		&level.ExcellentStepThreshold, &level.NormalStepThreshold, &excellentTime, &normalTime,
		&timeLimit, &stepLimit, &level.Version, &level.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return level, ErrNotFound
	}
	if err != nil {
		return level, err
	}
	level.ExcellentTimeThreshold = nullableInt(excellentTime)
	level.NormalTimeThreshold = nullableInt(normalTime)
	level.TimeLimitSeconds = nullableInt(timeLimit)
	level.StepLimit = nullableInt(stepLimit)
	return level, nil
}

func systemControlBoolTx(ctx context.Context, tx *sql.Tx, key string, fallback bool) (bool, error) {
	var valueType string
	var valueText, valueJSON sql.NullString
	err := tx.QueryRowContext(ctx, `SELECT value_type, value_text, value_json
		FROM system_controls
		WHERE control_key = ?
			AND enabled = 1
			AND (effective_from IS NULL OR effective_from <= NOW())
			AND (effective_until IS NULL OR effective_until >= NOW())`, key).Scan(&valueType, &valueText, &valueJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return fallback, nil
	}
	if err != nil {
		return fallback, err
	}

	value, err := systemControlValue(valueType, valueText, valueJSON)
	if err != nil {
		return fallback, err
	}
	boolValue, ok := value.(bool)
	if !ok {
		return fallback, nil
	}
	return boolValue, nil
}

func systemControlValue(valueType string, valueText sql.NullString, valueJSON sql.NullString) (any, error) {
	switch valueType {
	case "bool":
		text := strings.ToLower(strings.TrimSpace(valueText.String))
		return text == "1" || text == "true" || text == "yes" || text == "on", nil
	case "int":
		if !valueText.Valid || strings.TrimSpace(valueText.String) == "" {
			return 0, nil
		}
		return strconv.Atoi(strings.TrimSpace(valueText.String))
	case "decimal":
		if !valueText.Valid || strings.TrimSpace(valueText.String) == "" {
			return float64(0), nil
		}
		return strconv.ParseFloat(strings.TrimSpace(valueText.String), 64)
	case "json":
		if !valueJSON.Valid || strings.TrimSpace(valueJSON.String) == "" {
			return map[string]any{}, nil
		}
		var value any
		if err := json.Unmarshal([]byte(valueJSON.String), &value); err != nil {
			return nil, err
		}
		return value, nil
	default:
		if !valueText.Valid {
			return "", nil
		}
		return valueText.String, nil
	}
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

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return *value
}

func normalizeReason(reason string) string {
	switch reason {
	case "completed", "time_out", "mismatch_limit", "quit":
		return reason
	default:
		return "unknown"
	}
}

func calculateResultStars(result domain.LevelResult, level domain.LevelConfig) int {
	elapsedSeconds := (result.ElapsedMs + 999) / 1000
	if result.Steps <= level.ExcellentStepThreshold && withinThreshold(elapsedSeconds, level.ExcellentTimeThreshold) {
		return 3
	}
	if result.Steps <= level.NormalStepThreshold && withinThreshold(elapsedSeconds, level.NormalTimeThreshold) {
		return 2
	}
	return 1
}

func coinRewardForStars(level domain.LevelConfig, stars int) int {
	switch stars {
	case 1:
		if level.CoinRewardStar1 > 0 {
			return level.CoinRewardStar1
		}
	case 2:
		if level.CoinRewardStar2 > 0 {
			return level.CoinRewardStar2
		}
	case 3:
		if level.CoinRewardStar3 > 0 {
			return level.CoinRewardStar3
		}
	}

	if stars <= 0 {
		return 0
	}
	if level.CoinRewardBase > 0 {
		return stars * level.CoinRewardBase
	}
	return stars * 10
}

func withinThreshold(value int, threshold *int) bool {
	return threshold == nil || value <= *threshold
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

func randomChineseNickname() (string, error) {
	surnames := []string{
		"赵", "钱", "孙", "李", "周", "吴", "郑", "王", "冯", "陈",
		"褚", "卫", "蒋", "沈", "韩", "杨", "朱", "秦", "尤", "许",
		"何", "吕", "施", "张", "孔", "曹", "严", "华", "金", "魏",
		"陶", "姜", "戚", "谢", "邹", "喻", "柏", "水", "窦", "章",
		"云", "苏", "潘", "葛", "奚", "范", "彭", "郎", "鲁", "韦",
	}
	givenChars := []string{
		"安", "柏", "辰", "澄", "川", "岚", "宁", "清", "然", "若",
		"书", "思", "望", "微", "溪", "晓", "言", "一", "予", "知",
		"舟", "景", "星", "月", "云", "禾", "秋", "南", "北", "青",
		"夏", "冬", "明", "远", "初", "白", "锦", "乐", "可", "宜",
	}

	surname, err := randomFrom(surnames)
	if err != nil {
		return "", err
	}
	first, err := randomFrom(givenChars)
	if err != nil {
		return "", err
	}
	second, err := randomFrom(givenChars)
	if err != nil {
		return "", err
	}
	return surname + first + second, nil
}

func randomFrom(values []string) (string, error) {
	if len(values) == 0 {
		return "", errors.New("empty random values")
	}
	index, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(len(values))))
	if err != nil {
		return "", err
	}
	return values[index.Int64()], nil
}
