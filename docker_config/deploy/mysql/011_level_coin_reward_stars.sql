USE fpxxl;
SET NAMES utf8mb4;

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

ALTER TABLE level_configs
  MODIFY COLUMN coin_reward_star1 INT UNSIGNED NOT NULL DEFAULT 10 COMMENT '1 星通关金币奖励',
  MODIFY COLUMN coin_reward_star2 INT UNSIGNED NOT NULL DEFAULT 20 COMMENT '2 星通关金币奖励',
  MODIFY COLUMN coin_reward_star3 INT UNSIGNED NOT NULL DEFAULT 30 COMMENT '3 星通关金币奖励';
