CREATE DATABASE IF NOT EXISTS fpxxl CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE fpxxl;

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

CREATE TABLE IF NOT EXISTS level_configs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  level_id INT UNSIGNED NOT NULL,
  rows_count INT UNSIGNED NOT NULL,
  cols_count INT UNSIGNED NOT NULL,
  pair_count INT UNSIGNED NOT NULL,
  mode ENUM('normal', 'time_limit', 'step_limit') NOT NULL DEFAULT 'normal',
  theme_id VARCHAR(64) NOT NULL DEFAULT 'animal',
  initial_preview_ms INT UNSIGNED NOT NULL DEFAULT 2000,
  flip_back_delay_ms INT UNSIGNED NOT NULL DEFAULT 700,
  level_time_limit_seconds INT UNSIGNED NOT NULL DEFAULT 120,
  max_mismatch_count INT UNSIGNED NOT NULL DEFAULT 12,
  show_steps TINYINT(1) NOT NULL DEFAULT 1,
  show_timer TINYINT(1) NOT NULL DEFAULT 1,
  show_mismatch TINYINT(1) NOT NULL DEFAULT 1,
  hint_highlight_ms INT UNSIGNED NOT NULL DEFAULT 1300,
  coin_reward_base INT UNSIGNED NOT NULL DEFAULT 10,
  coin_reward_star1 INT UNSIGNED NOT NULL DEFAULT 10,
  coin_reward_star2 INT UNSIGNED NOT NULL DEFAULT 20,
  coin_reward_star3 INT UNSIGNED NOT NULL DEFAULT 30,
  stamina_cost INT UNSIGNED NOT NULL DEFAULT 1,
  excellent_step_threshold INT UNSIGNED NOT NULL,
  normal_step_threshold INT UNSIGNED NOT NULL,
  excellent_time_threshold INT UNSIGNED NULL,
  normal_time_threshold INT UNSIGNED NULL,
  time_limit_seconds INT UNSIGNED NULL,
  step_limit INT UNSIGNED NULL,
  enabled TINYINT(1) NOT NULL DEFAULT 1,
  version INT UNSIGNED NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_level_configs_level_id (level_id),
  KEY idx_level_configs_enabled (enabled, level_id),
  KEY idx_level_configs_theme_id (theme_id),
  CONSTRAINT chk_level_configs_grid CHECK (
    rows_count * cols_count = pair_count * 2
    OR (
      rows_count * cols_count = pair_count * 2 + 1
      AND MOD(rows_count * cols_count, 2) = 1
    )
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS ad_frequency_configs (
  id TINYINT UNSIGNED NOT NULL PRIMARY KEY,
  no_interstitial_before_level INT UNSIGNED NOT NULL DEFAULT 4,
  interstitial_every_levels INT UNSIGNED NOT NULL DEFAULT 4,
  max_interstitial_per_day INT UNSIGNED NOT NULL DEFAULT 10,
  max_revive_per_level INT UNSIGNED NOT NULL DEFAULT 1,
  banner_enabled_scenes JSON NOT NULL,
  version INT UNSIGNED NOT NULL DEFAULT 1,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS system_controls (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  control_key VARCHAR(128) NOT NULL,
  control_group VARCHAR(64) NOT NULL DEFAULT 'general',
  name VARCHAR(128) NOT NULL DEFAULT '',
  description VARCHAR(512) NOT NULL DEFAULT '',
  value_type ENUM('bool', 'int', 'decimal', 'string', 'json') NOT NULL DEFAULT 'string',
  value_text TEXT NULL,
  value_json JSON NULL,
  default_value_text TEXT NULL,
  default_value_json JSON NULL,
  enabled TINYINT(1) NOT NULL DEFAULT 1,
  is_public TINYINT(1) NOT NULL DEFAULT 0,
  sort_order INT NOT NULL DEFAULT 0,
  version INT UNSIGNED NOT NULL DEFAULT 1,
  effective_from DATETIME NULL,
  effective_until DATETIME NULL,
  created_by BIGINT UNSIGNED NULL,
  updated_by BIGINT UNSIGNED NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_system_controls_key (control_key),
  KEY idx_system_controls_group_enabled (control_group, enabled, sort_order),
  KEY idx_system_controls_public_enabled (is_public, enabled, sort_order),
  KEY idx_system_controls_effective (effective_from, effective_until),
  KEY idx_system_controls_updated_at (updated_at)
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
  reason ENUM(
    'level_complete',
    'ad_reward',
    'daily_reward',
    'activity_reward',
    'shop_purchase',
    'refund',
    'admin_adjust'
  ) NOT NULL,
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
  reason ENUM(
    'level_start',
    'coin_exchange',
    'auto_recover',
    'ad_reward',
    'activity_reward',
    'refund',
    'admin_adjust'
  ) NOT NULL,
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
  source ENUM('register', 'use', 'ad_reward', 'admin_adjust') NOT NULL,
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

CREATE TABLE IF NOT EXISTS admin_users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(64) NOT NULL,
  email VARCHAR(255) NULL,
  password_hash VARCHAR(255) NOT NULL,
  display_name VARCHAR(64) NOT NULL DEFAULT '',
  role ENUM('owner', 'operator', 'viewer') NOT NULL DEFAULT 'operator',
  permissions JSON NULL,
  status ENUM('active', 'disabled', 'locked') NOT NULL DEFAULT 'active',
  failed_login_attempts INT UNSIGNED NOT NULL DEFAULT 0,
  locked_until DATETIME NULL,
  password_changed_at DATETIME NULL,
  last_login_at DATETIME NULL,
  last_login_ip VARCHAR(45) NOT NULL DEFAULT '',
  created_by BIGINT UNSIGNED NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_admin_users_username (username),
  UNIQUE KEY uk_admin_users_email (email),
  KEY idx_admin_users_status_role (status, role),
  KEY idx_admin_users_locked_until (locked_until),
  CONSTRAINT fk_admin_users_created_by FOREIGN KEY (created_by) REFERENCES admin_users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO theme_configs (theme_id, name, asset_base_path, icon_keys)
VALUES (
  'animal',
  '动物主题',
  'assets/animals',
  JSON_ARRAY('panda', 'tiger', 'rabbit', 'bear', 'fox', 'cat', 'dog', 'penguin', 'koala', 'cow', 'chicken', 'dinosaur', 'elefants', 'frog', 'girafa', 'hedgehog', 'monkey', 'owl', 'seal', 'sheep')
)
ON DUPLICATE KEY UPDATE theme_id = theme_id;

INSERT INTO ad_frequency_configs (
  id,
  no_interstitial_before_level,
  interstitial_every_levels,
  max_interstitial_per_day,
  max_revive_per_level,
  banner_enabled_scenes
) VALUES (1, 4, 4, 10, 1, JSON_ARRAY('home', 'result'))
ON DUPLICATE KEY UPDATE id = id;

INSERT INTO system_controls (
  control_key,
  control_group,
  name,
  description,
  value_type,
  value_text,
  value_json,
  default_value_text,
  default_value_json,
  enabled,
  is_public,
  sort_order
) VALUES
  ('system.maintenance_mode', 'system', '维护模式', '开启后可用于后台或接口统一拦截非管理流量', 'bool', 'false', NULL, 'false', NULL, 1, 1, 10),
  ('client.min_version', 'client', '最低客户端版本', '低于该版本时可提示升级', 'string', '1.0.0', NULL, '1.0.0', NULL, 1, 1, 20),
  ('client.force_update', 'client', '强制更新', '开启后客户端可根据版本策略强制升级', 'bool', 'false', NULL, 'false', NULL, 1, 1, 30),
  ('game.unlimited_stamina', 'game', '无限体力', '开启后隐藏客户端体力栏，并且开始关卡不消耗体力', 'bool', 'false', NULL, 'false', NULL, 1, 1, 35),
  ('game.feature_flags', 'game', '玩法功能开关', '全局玩法功能开关集合', 'json', NULL, JSON_OBJECT('timeMode', false, 'leaderboard', true, 'shop', true), NULL, JSON_OBJECT('timeMode', false, 'leaderboard', true, 'shop', true), 1, 1, 40),
  ('notice.home_banner', 'notice', '首页公告', '小游戏首页公告或运营提示', 'json', NULL, JSON_OBJECT('enabled', false, 'title', '', 'content', ''), NULL, JSON_OBJECT('enabled', false, 'title', '', 'content', ''), 1, 1, 50)
ON DUPLICATE KEY UPDATE control_key = control_key;

INSERT INTO ad_placements (placement_key, ad_type, scene, reward_type, reward_amount)
VALUES
  ('hint', 'rewarded_video', 'game', 'hints', 1),
  ('double_reward', 'rewarded_video', 'result', 'coins', 0),
  ('extra_coins', 'rewarded_video', 'result', 'coins', 20),
  ('revive', 'rewarded_video', 'failed', 'revive', 1),
  ('between_levels', 'interstitial', 'result', '', 0),
  ('home_banner', 'banner', 'home', '', 0)
ON DUPLICATE KEY UPDATE placement_key = placement_key;

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
ON DUPLICATE KEY UPDATE product_key = product_key;
