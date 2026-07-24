USE fpxxl;

SET @add_show_steps = (
  SELECT IF(
    NOT EXISTS (
      SELECT 1 FROM information_schema.columns
      WHERE table_schema = DATABASE()
        AND table_name = 'level_configs'
        AND column_name = 'show_steps'
    ),
    'ALTER TABLE level_configs ADD COLUMN show_steps TINYINT(1) NOT NULL DEFAULT 1 AFTER max_mismatch_count',
    'SELECT 1'
  )
);
PREPARE stmt FROM @add_show_steps;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @add_show_timer = (
  SELECT IF(
    NOT EXISTS (
      SELECT 1 FROM information_schema.columns
      WHERE table_schema = DATABASE()
        AND table_name = 'level_configs'
        AND column_name = 'show_timer'
    ),
    'ALTER TABLE level_configs ADD COLUMN show_timer TINYINT(1) NOT NULL DEFAULT 1 AFTER show_steps',
    'SELECT 1'
  )
);
PREPARE stmt FROM @add_show_timer;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @add_show_mismatch = (
  SELECT IF(
    NOT EXISTS (
      SELECT 1 FROM information_schema.columns
      WHERE table_schema = DATABASE()
        AND table_name = 'level_configs'
        AND column_name = 'show_mismatch'
    ),
    'ALTER TABLE level_configs ADD COLUMN show_mismatch TINYINT(1) NOT NULL DEFAULT 1 AFTER show_timer',
    'SELECT 1'
  )
);
PREPARE stmt FROM @add_show_mismatch;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @add_hint_highlight_ms = (
  SELECT IF(
    NOT EXISTS (
      SELECT 1 FROM information_schema.columns
      WHERE table_schema = DATABASE()
        AND table_name = 'level_configs'
        AND column_name = 'hint_highlight_ms'
    ),
    'ALTER TABLE level_configs ADD COLUMN hint_highlight_ms INT NOT NULL DEFAULT 1300 AFTER show_mismatch',
    'SELECT 1'
  )
);
PREPARE stmt FROM @add_hint_highlight_ms;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @add_coin_reward_base = (
  SELECT IF(
    NOT EXISTS (
      SELECT 1 FROM information_schema.columns
      WHERE table_schema = DATABASE()
        AND table_name = 'level_configs'
        AND column_name = 'coin_reward_base'
    ),
    'ALTER TABLE level_configs ADD COLUMN coin_reward_base INT NOT NULL DEFAULT 10 AFTER hint_highlight_ms',
    'SELECT 1'
  )
);
PREPARE stmt FROM @add_coin_reward_base;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @add_stamina_cost = (
  SELECT IF(
    NOT EXISTS (
      SELECT 1 FROM information_schema.columns
      WHERE table_schema = DATABASE()
        AND table_name = 'level_configs'
        AND column_name = 'stamina_cost'
    ),
    'ALTER TABLE level_configs ADD COLUMN stamina_cost INT NOT NULL DEFAULT 1 AFTER coin_reward_base',
    'SELECT 1'
  )
);
PREPARE stmt FROM @add_stamina_cost;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @add_coin_reward_star1 = (
  SELECT IF(
    NOT EXISTS (
      SELECT 1 FROM information_schema.columns
      WHERE table_schema = DATABASE()
        AND table_name = 'level_configs'
        AND column_name = 'coin_reward_star1'
    ),
    'ALTER TABLE level_configs ADD COLUMN coin_reward_star1 INT UNSIGNED NOT NULL DEFAULT 10 AFTER coin_reward_base',
    'SELECT 1'
  )
);
PREPARE stmt FROM @add_coin_reward_star1;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @add_coin_reward_star2 = (
  SELECT IF(
    NOT EXISTS (
      SELECT 1 FROM information_schema.columns
      WHERE table_schema = DATABASE()
        AND table_name = 'level_configs'
        AND column_name = 'coin_reward_star2'
    ),
    'ALTER TABLE level_configs ADD COLUMN coin_reward_star2 INT UNSIGNED NOT NULL DEFAULT 20 AFTER coin_reward_star1',
    'SELECT 1'
  )
);
PREPARE stmt FROM @add_coin_reward_star2;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @add_coin_reward_star3 = (
  SELECT IF(
    NOT EXISTS (
      SELECT 1 FROM information_schema.columns
      WHERE table_schema = DATABASE()
        AND table_name = 'level_configs'
        AND column_name = 'coin_reward_star3'
    ),
    'ALTER TABLE level_configs ADD COLUMN coin_reward_star3 INT UNSIGNED NOT NULL DEFAULT 30 AFTER coin_reward_star2',
    'SELECT 1'
  )
);
PREPARE stmt FROM @add_coin_reward_star3;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

INSERT INTO level_configs (
  level_id, rows_count, cols_count, pair_count, mode, theme_id,
  initial_preview_ms, flip_back_delay_ms, level_time_limit_seconds, max_mismatch_count,
  show_steps, show_timer, show_mismatch, hint_highlight_ms, coin_reward_base,
  coin_reward_star1, coin_reward_star2, coin_reward_star3, stamina_cost,
  excellent_step_threshold, normal_step_threshold, excellent_time_threshold, normal_time_threshold,
  time_limit_seconds, step_limit, enabled, version
)
WITH RECURSIVE level_numbers AS (
  SELECT 1 AS level_id
  UNION ALL
  SELECT level_id + 1 FROM level_numbers WHERE level_id < 100
),
level_defaults AS (
  SELECT
    level_id,
    CASE
      WHEN level_id = 1 THEN 2
      WHEN level_id <= 5 THEN 3
      ELSE 4
    END AS rows_count,
    CASE
      WHEN level_id = 1 THEN 2
      WHEN level_id <= 5 THEN 4
      ELSE 4
    END AS cols_count,
    CASE
      WHEN level_id = 1 THEN 2
      WHEN level_id <= 5 THEN 6
      ELSE 8
    END AS pair_count,
    'normal' AS mode,
    'animal' AS theme_id,
    CASE
      WHEN level_id <= 10 THEN 2500
      WHEN level_id <= 40 THEN 2200
      ELSE 2000
    END AS initial_preview_ms,
    CASE
      WHEN level_id <= 30 THEN 800
      WHEN level_id <= 70 THEN 700
      ELSE 600
    END AS flip_back_delay_ms,
    CASE
      WHEN level_id = 1 THEN 45
      WHEN level_id <= 5 THEN 90
      WHEN level_id <= 20 THEN 120
      WHEN level_id <= 50 THEN 110
      WHEN level_id <= 80 THEN 100
      ELSE 90
    END AS level_time_limit_seconds,
    CASE
      WHEN level_id = 1 THEN 6
      WHEN level_id <= 5 THEN 10
      WHEN level_id <= 30 THEN 12
      WHEN level_id <= 60 THEN 11
      WHEN level_id <= 85 THEN 10
      ELSE 9
    END AS max_mismatch_count,
    1 AS show_steps,
    1 AS show_timer,
    1 AS show_mismatch,
    1300 AS hint_highlight_ms,
    10 AS coin_reward_base,
    10 AS coin_reward_star1,
    20 AS coin_reward_star2,
    30 AS coin_reward_star3,
    1 AS stamina_cost,
    CASE
      WHEN level_id = 1 THEN 3
      WHEN level_id <= 5 THEN 9
      WHEN level_id <= 30 THEN 12
      WHEN level_id <= 70 THEN 11
      ELSE 10
    END AS excellent_step_threshold,
    CASE
      WHEN level_id = 1 THEN 5
      WHEN level_id <= 5 THEN 14
      WHEN level_id <= 30 THEN 18
      WHEN level_id <= 70 THEN 17
      ELSE 16
    END AS normal_step_threshold,
    CASE
      WHEN level_id = 1 THEN 20
      WHEN level_id <= 5 THEN 45
      WHEN level_id <= 30 THEN 70
      WHEN level_id <= 70 THEN 65
      ELSE 60
    END AS excellent_time_threshold,
    CASE
      WHEN level_id = 1 THEN 35
      WHEN level_id <= 5 THEN 75
      WHEN level_id <= 30 THEN 105
      WHEN level_id <= 70 THEN 98
      ELSE 90
    END AS normal_time_threshold,
    NULL AS time_limit_seconds,
    NULL AS step_limit,
    1 AS enabled,
    1 AS version
  FROM level_numbers
)
SELECT
  level_id, rows_count, cols_count, pair_count, mode, theme_id,
  initial_preview_ms, flip_back_delay_ms, level_time_limit_seconds, max_mismatch_count,
  show_steps, show_timer, show_mismatch, hint_highlight_ms, coin_reward_base,
  coin_reward_star1, coin_reward_star2, coin_reward_star3, stamina_cost,
  excellent_step_threshold, normal_step_threshold, excellent_time_threshold, normal_time_threshold,
  time_limit_seconds, step_limit, enabled, version
FROM level_defaults
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
  show_steps = VALUES(show_steps),
  show_timer = VALUES(show_timer),
  show_mismatch = VALUES(show_mismatch),
  hint_highlight_ms = VALUES(hint_highlight_ms),
  coin_reward_base = VALUES(coin_reward_base),
  coin_reward_star1 = VALUES(coin_reward_star1),
  coin_reward_star2 = VALUES(coin_reward_star2),
  coin_reward_star3 = VALUES(coin_reward_star3),
  stamina_cost = VALUES(stamina_cost),
  excellent_step_threshold = VALUES(excellent_step_threshold),
  normal_step_threshold = VALUES(normal_step_threshold),
  excellent_time_threshold = VALUES(excellent_time_threshold),
  normal_time_threshold = VALUES(normal_time_threshold),
  time_limit_seconds = VALUES(time_limit_seconds),
  step_limit = VALUES(step_limit),
  enabled = VALUES(enabled),
  version = VALUES(version);

ALTER TABLE level_configs
  MODIFY COLUMN show_steps TINYINT(1) NOT NULL DEFAULT 1 COMMENT 'Show step counter in level',
  MODIFY COLUMN show_timer TINYINT(1) NOT NULL DEFAULT 1 COMMENT 'Show timer in level',
  MODIFY COLUMN show_mismatch TINYINT(1) NOT NULL DEFAULT 1 COMMENT 'Show mismatch counter in level',
  MODIFY COLUMN hint_highlight_ms INT NOT NULL DEFAULT 1300 COMMENT 'Hint highlight duration in milliseconds',
  MODIFY COLUMN coin_reward_base INT NOT NULL DEFAULT 10 COMMENT 'Base coin reward per star',
  MODIFY COLUMN coin_reward_star1 INT UNSIGNED NOT NULL DEFAULT 10 COMMENT '1 星通关金币奖励',
  MODIFY COLUMN coin_reward_star2 INT UNSIGNED NOT NULL DEFAULT 20 COMMENT '2 星通关金币奖励',
  MODIFY COLUMN coin_reward_star3 INT UNSIGNED NOT NULL DEFAULT 30 COMMENT '3 星通关金币奖励',
  MODIFY COLUMN stamina_cost INT NOT NULL DEFAULT 1 COMMENT 'Stamina cost to start level';
