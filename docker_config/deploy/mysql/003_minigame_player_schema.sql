USE fpxxl;

SET @rename_player_progress = (
  SELECT IF(
    EXISTS (
      SELECT 1 FROM information_schema.tables
      WHERE table_schema = DATABASE()
        AND table_name = 'player_progress'
    )
    AND NOT EXISTS (
      SELECT 1 FROM information_schema.columns
      WHERE table_schema = DATABASE()
        AND table_name = 'player_progress'
        AND column_name = 'player_id'
    )
    AND NOT EXISTS (
      SELECT 1 FROM information_schema.tables
      WHERE table_schema = DATABASE()
        AND table_name = 'player_progress_legacy'
    ),
    'RENAME TABLE player_progress TO player_progress_legacy',
    'SELECT 1'
  )
);
PREPARE stmt FROM @rename_player_progress;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @rename_level_results = (
  SELECT IF(
    EXISTS (
      SELECT 1 FROM information_schema.tables
      WHERE table_schema = DATABASE()
        AND table_name = 'level_results'
    )
    AND NOT EXISTS (
      SELECT 1 FROM information_schema.columns
      WHERE table_schema = DATABASE()
        AND table_name = 'level_results'
        AND column_name = 'player_id'
    )
    AND NOT EXISTS (
      SELECT 1 FROM information_schema.tables
      WHERE table_schema = DATABASE()
        AND table_name = 'level_results_legacy'
    ),
    'RENAME TABLE level_results TO level_results_legacy',
    'SELECT 1'
  )
);
PREPARE stmt FROM @rename_level_results;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @rename_leaderboard_entries = (
  SELECT IF(
    EXISTS (
      SELECT 1 FROM information_schema.tables
      WHERE table_schema = DATABASE()
        AND table_name = 'leaderboard_entries'
    )
    AND NOT EXISTS (
      SELECT 1 FROM information_schema.columns
      WHERE table_schema = DATABASE()
        AND table_name = 'leaderboard_entries'
        AND column_name = 'player_id'
    )
    AND NOT EXISTS (
      SELECT 1 FROM information_schema.tables
      WHERE table_schema = DATABASE()
        AND table_name = 'leaderboard_entries_legacy'
    ),
    'RENAME TABLE leaderboard_entries TO leaderboard_entries_legacy',
    'SELECT 1'
  )
);
PREPARE stmt FROM @rename_leaderboard_entries;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @rename_game_events = (
  SELECT IF(
    EXISTS (
      SELECT 1 FROM information_schema.tables
      WHERE table_schema = DATABASE()
        AND table_name = 'game_events'
    )
    AND NOT EXISTS (
      SELECT 1 FROM information_schema.columns
      WHERE table_schema = DATABASE()
        AND table_name = 'game_events'
        AND column_name = 'player_id'
    )
    AND NOT EXISTS (
      SELECT 1 FROM information_schema.tables
      WHERE table_schema = DATABASE()
        AND table_name = 'game_events_legacy'
    ),
    'RENAME TABLE game_events TO game_events_legacy',
    'SELECT 1'
  )
);
PREPARE stmt FROM @rename_game_events;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @add_level_version = (
  SELECT IF(
    NOT EXISTS (
      SELECT 1 FROM information_schema.columns
      WHERE table_schema = DATABASE()
        AND table_name = 'level_configs'
        AND column_name = 'version'
    ),
    'ALTER TABLE level_configs ADD COLUMN version INT UNSIGNED NOT NULL DEFAULT 1 AFTER enabled',
    'SELECT 1'
  )
);
PREPARE stmt FROM @add_level_version;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @add_ad_version = (
  SELECT IF(
    NOT EXISTS (
      SELECT 1 FROM information_schema.columns
      WHERE table_schema = DATABASE()
        AND table_name = 'ad_frequency_configs'
        AND column_name = 'version'
    ),
    'ALTER TABLE ad_frequency_configs ADD COLUMN version INT UNSIGNED NOT NULL DEFAULT 1 AFTER banner_enabled_scenes',
    'SELECT 1'
  )
);
PREPARE stmt FROM @add_ad_version;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

CREATE TABLE IF NOT EXISTS players (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  open_id VARCHAR(128) NOT NULL,
  union_id VARCHAR(128) NULL,
  nickname VARCHAR(64) NOT NULL DEFAULT '',
  avatar_url VARCHAR(512) NOT NULL DEFAULT '',
  status ENUM('active', 'blocked', 'deleted') NOT NULL DEFAULT 'active',
  last_login_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_players_open_id (open_id),
  KEY idx_players_union_id (union_id),
  KEY idx_players_last_login_at (last_login_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS theme_configs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  theme_id VARCHAR(64) NOT NULL,
  name VARCHAR(64) NOT NULL,
  asset_base_path VARCHAR(255) NOT NULL DEFAULT '',
  icon_keys JSON NOT NULL,
  enabled TINYINT(1) NOT NULL DEFAULT 1,
  version INT UNSIGNED NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_theme_configs_theme_id (theme_id),
  KEY idx_theme_configs_enabled (enabled, updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS ad_placements (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  placement_key VARCHAR(64) NOT NULL,
  ad_type ENUM('rewarded_video', 'interstitial', 'banner') NOT NULL,
  scene VARCHAR(64) NOT NULL DEFAULT '',
  reward_type VARCHAR(64) NOT NULL DEFAULT '',
  reward_amount INT NOT NULL DEFAULT 0,
  enabled TINYINT(1) NOT NULL DEFAULT 1,
  extra_config JSON NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_ad_placements_key (placement_key),
  KEY idx_ad_placements_enabled_type (enabled, ad_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS player_progress (
  player_id BIGINT UNSIGNED NOT NULL PRIMARY KEY,
  current_level INT UNSIGNED NOT NULL DEFAULT 1,
  coins INT NOT NULL DEFAULT 0,
  stamina INT NOT NULL DEFAULT 5,
  max_stamina INT UNSIGNED NOT NULL DEFAULT 5,
  hints INT NOT NULL DEFAULT 3,
  preview_again_count INT NOT NULL DEFAULT 3,
  remove_pair_count INT NOT NULL DEFAULT 3,
  level_stars JSON NOT NULL,
  completed_levels JSON NOT NULL,
  next_stamina_recover_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_player_progress_next_recover (next_stamina_recover_at),
  CONSTRAINT fk_player_progress_player FOREIGN KEY (player_id) REFERENCES players(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS level_results (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  player_id BIGINT UNSIGNED NOT NULL,
  level_id INT UNSIGNED NOT NULL,
  success TINYINT(1) NOT NULL,
  reason ENUM('completed', 'time_out', 'mismatch_limit', 'quit', 'unknown') NOT NULL DEFAULT 'unknown',
  steps INT UNSIGNED NOT NULL DEFAULT 0,
  mismatch_count INT UNSIGNED NOT NULL DEFAULT 0,
  elapsed_ms INT UNSIGNED NOT NULL DEFAULT 0,
  stars TINYINT UNSIGNED NOT NULL DEFAULT 0,
  coins_earned INT NOT NULL DEFAULT 0,
  used_hints INT UNSIGNED NOT NULL DEFAULT 0,
  completed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_level_results_player FOREIGN KEY (player_id) REFERENCES players(id),
  KEY idx_level_results_player_level (player_id, level_id, completed_at),
  KEY idx_level_results_level_rank (level_id, stars, steps, elapsed_ms),
  KEY idx_level_results_completed_at (completed_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS ad_events (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  event_id VARCHAR(128) NOT NULL,
  player_id BIGINT UNSIGNED NULL,
  ad_type ENUM('rewarded_video', 'interstitial', 'banner') NOT NULL,
  placement_key VARCHAR(64) NOT NULL,
  status ENUM('requested', 'shown', 'completed', 'closed', 'failed', 'reward_granted') NOT NULL,
  level_id INT UNSIGNED NULL,
  reward_id VARCHAR(128) NULL,
  error_message VARCHAR(255) NOT NULL DEFAULT '',
  payload JSON NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_ad_events_event_id (event_id),
  KEY idx_ad_events_player_time (player_id, created_at),
  KEY idx_ad_events_placement_status (placement_key, status, created_at),
  KEY idx_ad_events_reward_id (reward_id),
  CONSTRAINT fk_ad_events_player FOREIGN KEY (player_id) REFERENCES players(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS reward_grants (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  reward_id VARCHAR(128) NOT NULL,
  player_id BIGINT UNSIGNED NOT NULL,
  source ENUM('level_complete', 'ad', 'daily', 'activity', 'shop', 'admin') NOT NULL,
  source_ref VARCHAR(128) NOT NULL DEFAULT '',
  reward_type ENUM('coins', 'stamina', 'hints', 'revive', 'time', 'steps') NOT NULL,
  amount INT NOT NULL,
  level_id INT UNSIGNED NULL,
  granted_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_reward_grants_reward_id (reward_id),
  KEY idx_reward_grants_player_time (player_id, granted_at),
  KEY idx_reward_grants_source (source, source_ref),
  CONSTRAINT fk_reward_grants_player FOREIGN KEY (player_id) REFERENCES players(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS coin_transactions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  transaction_no VARCHAR(128) NOT NULL,
  player_id BIGINT UNSIGNED NOT NULL,
  change_amount INT NOT NULL,
  balance_after INT NOT NULL,
  reason ENUM('level_complete', 'ad_reward', 'daily_reward', 'activity_reward', 'shop_purchase', 'refund', 'admin_adjust') NOT NULL,
  ref_type VARCHAR(64) NOT NULL DEFAULT '',
  ref_id VARCHAR(128) NOT NULL DEFAULT '',
  note VARCHAR(255) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_coin_transactions_no (transaction_no),
  KEY idx_coin_transactions_player_time (player_id, created_at),
  KEY idx_coin_transactions_reason_time (reason, created_at),
  KEY idx_coin_transactions_ref (ref_type, ref_id),
  CONSTRAINT fk_coin_transactions_player FOREIGN KEY (player_id) REFERENCES players(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS stamina_transactions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  transaction_no VARCHAR(128) NOT NULL,
  player_id BIGINT UNSIGNED NOT NULL,
  change_amount INT NOT NULL,
  balance_after INT NOT NULL,
  reason ENUM('level_start', 'coin_exchange', 'auto_recover', 'ad_reward', 'activity_reward', 'refund', 'admin_adjust') NOT NULL,
  ref_type VARCHAR(64) NOT NULL DEFAULT '',
  ref_id VARCHAR(128) NOT NULL DEFAULT '',
  note VARCHAR(255) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_stamina_transactions_no (transaction_no),
  KEY idx_stamina_transactions_player_time (player_id, created_at),
  KEY idx_stamina_transactions_reason_time (reason, created_at),
  KEY idx_stamina_transactions_ref (ref_type, ref_id),
  CONSTRAINT fk_stamina_transactions_player FOREIGN KEY (player_id) REFERENCES players(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS tool_transactions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  transaction_no VARCHAR(128) NOT NULL,
  player_id BIGINT UNSIGNED NOT NULL,
  tool_type ENUM('hint', 'preview_again', 'remove_pair') NOT NULL,
  change_amount INT NOT NULL,
  balance_after INT NOT NULL,
  source ENUM('register', 'use', 'ad_reward', 'shop_purchase', 'admin_adjust') NOT NULL,
  ref_type VARCHAR(64) NOT NULL DEFAULT '',
  ref_id VARCHAR(128) NOT NULL DEFAULT '',
  note VARCHAR(255) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_tool_transactions_no (transaction_no),
  KEY idx_tool_transactions_player_time (player_id, created_at),
  KEY idx_tool_transactions_tool_source (tool_type, source, created_at),
  CONSTRAINT fk_tool_transactions_player FOREIGN KEY (player_id) REFERENCES players(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS shop_products (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  product_key VARCHAR(64) NOT NULL,
  name VARCHAR(64) NOT NULL,
  product_type ENUM('stamina', 'item', 'bundle') NOT NULL,
  currency_type ENUM('coins', 'money', 'ad') NOT NULL DEFAULT 'coins',
  currency_amount INT UNSIGNED NOT NULL DEFAULT 0,
  grant_type ENUM('stamina', 'hints', 'coins', 'bundle') NOT NULL,
  grant_amount INT UNSIGNED NOT NULL DEFAULT 0,
  daily_buy_limit INT UNSIGNED NULL,
  enabled TINYINT(1) NOT NULL DEFAULT 1,
  sort_order INT NOT NULL DEFAULT 0,
  extra_config JSON NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_shop_products_key (product_key),
  KEY idx_shop_products_enabled_sort (enabled, sort_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS purchase_orders (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  order_no VARCHAR(128) NOT NULL,
  player_id BIGINT UNSIGNED NOT NULL,
  product_id BIGINT UNSIGNED NOT NULL,
  product_key VARCHAR(64) NOT NULL,
  product_name VARCHAR(64) NOT NULL,
  currency_type ENUM('coins', 'money', 'ad') NOT NULL,
  currency_amount INT UNSIGNED NOT NULL,
  grant_type ENUM('stamina', 'hints', 'coins', 'bundle') NOT NULL,
  grant_amount INT UNSIGNED NOT NULL,
  status ENUM('created', 'paid', 'fulfilled', 'cancelled', 'failed', 'refunded') NOT NULL DEFAULT 'created',
  paid_at DATETIME NULL,
  fulfilled_at DATETIME NULL,
  failed_reason VARCHAR(255) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_purchase_orders_no (order_no),
  KEY idx_purchase_orders_player_time (player_id, created_at),
  KEY idx_purchase_orders_product_time (product_key, created_at),
  KEY idx_purchase_orders_status_time (status, created_at),
  CONSTRAINT fk_purchase_orders_player FOREIGN KEY (player_id) REFERENCES players(id),
  CONSTRAINT fk_purchase_orders_product FOREIGN KEY (product_id) REFERENCES shop_products(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS leaderboard_entries (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  player_id BIGINT UNSIGNED NOT NULL,
  level_id INT UNSIGNED NOT NULL,
  nickname VARCHAR(64) NOT NULL DEFAULT '',
  stars TINYINT UNSIGNED NOT NULL DEFAULT 0,
  steps INT UNSIGNED NOT NULL DEFAULT 0,
  elapsed_ms INT UNSIGNED NOT NULL DEFAULT 0,
  submitted_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_leaderboard_entries_player FOREIGN KEY (player_id) REFERENCES players(id),
  KEY idx_leaderboard_entries_rank (level_id, stars, steps, elapsed_ms, submitted_at),
  KEY idx_leaderboard_entries_player (player_id, level_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS game_events (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  event_id VARCHAR(128) NOT NULL,
  player_id BIGINT UNSIGNED NULL,
  event_type VARCHAR(64) NOT NULL,
  level_id INT UNSIGNED NULL,
  payload JSON NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_game_events_event_id (event_id),
  KEY idx_game_events_player_time (player_id, created_at),
  KEY idx_game_events_type_time (event_type, created_at),
  KEY idx_game_events_level_time (level_id, created_at),
  CONSTRAINT fk_game_events_player FOREIGN KEY (player_id) REFERENCES players(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO theme_configs (theme_id, name, asset_base_path, icon_keys)
VALUES (
  'animal',
  '动物主题',
  'assets/animals',
  JSON_ARRAY('panda', 'tiger', 'rabbit', 'bear', 'fox', 'cat', 'dog', 'penguin', 'koala', 'cow', 'chicken', 'dinosaur', 'elefants', 'frog', 'girafa', 'hedgehog', 'monkey', 'owl', 'seal', 'sheep')
)
ON DUPLICATE KEY UPDATE theme_id = theme_id;

INSERT INTO ad_placements (placement_key, ad_type, scene, reward_type, reward_amount)
VALUES
  ('hint', 'rewarded_video', 'game', 'hints', 1),
  ('double_reward', 'rewarded_video', 'result', 'coins', 0),
  ('extra_coins', 'rewarded_video', 'result', 'coins', 20),
  ('revive', 'rewarded_video', 'failed', 'revive', 1),
  ('between_levels', 'interstitial', 'result', '', 0),
  ('home_banner', 'banner', 'home', '', 0)
ON DUPLICATE KEY UPDATE placement_key = placement_key;

/*
INSERT INTO shop_products (
  product_key,
  name,
  product_type,
  currency_type,
  currency_amount,
  grant_type,
  grant_amount,
  daily_buy_limit,
  enabled,
  sort_order
) VALUES (
  'stamina_5_by_coins',
  '金币兑换 5 点体力',
  'stamina',
  'coins',
  100,
  'stamina',
  5,
  10,
  1,
  10
)
ON DUPLICATE KEY UPDATE product_key = product_key;
*/

INSERT INTO shop_products (
  product_key,
  name,
  product_type,
  currency_type,
  currency_amount,
  grant_type,
  grant_amount,
  daily_buy_limit,
  enabled,
  sort_order
) VALUES
  ('stamina_1_by_coins', '金币兑换 1 点体力', 'stamina', 'coins', 99, 'stamina', 1, 10, 1, 10),
  ('stamina_3_by_coins', '金币兑换 3 点体力', 'stamina', 'coins', 266, 'stamina', 3, 10, 1, 20),
  ('stamina_5_by_coins', '金币兑换 5 点体力', 'stamina', 'coins', 388, 'stamina', 5, 10, 1, 30)
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  currency_amount = VALUES(currency_amount),
  grant_amount = VALUES(grant_amount),
  daily_buy_limit = VALUES(daily_buy_limit),
  enabled = VALUES(enabled),
  sort_order = VALUES(sort_order);
