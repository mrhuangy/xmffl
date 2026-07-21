CREATE DATABASE IF NOT EXISTS fpxxl CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE fpxxl;

CREATE TABLE IF NOT EXISTS level_configs (
  level_id INT NOT NULL PRIMARY KEY,
  rows_count INT NOT NULL,
  cols_count INT NOT NULL,
  pair_count INT NOT NULL,
  mode ENUM('normal', 'time_limit', 'step_limit') NOT NULL DEFAULT 'normal',
  theme_id VARCHAR(64) NOT NULL DEFAULT 'animal',
  initial_preview_ms INT NOT NULL DEFAULT 2000,
  flip_back_delay_ms INT NOT NULL DEFAULT 700,
  level_time_limit_seconds INT NOT NULL DEFAULT 120,
  max_mismatch_count INT NOT NULL DEFAULT 12,
  show_steps TINYINT(1) NOT NULL DEFAULT 1,
  show_timer TINYINT(1) NOT NULL DEFAULT 1,
  show_mismatch TINYINT(1) NOT NULL DEFAULT 1,
  hint_highlight_ms INT NOT NULL DEFAULT 1300,
  coin_reward_base INT NOT NULL DEFAULT 10,
  stamina_cost INT NOT NULL DEFAULT 1,
  excellent_step_threshold INT NOT NULL,
  normal_step_threshold INT NOT NULL,
  excellent_time_threshold INT NULL,
  normal_time_threshold INT NULL,
  time_limit_seconds INT NULL,
  step_limit INT NULL,
  enabled TINYINT(1) NOT NULL DEFAULT 1,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  CONSTRAINT chk_level_grid CHECK (
    rows_count * cols_count = pair_count * 2
    OR (
      rows_count * cols_count = pair_count * 2 + 1
      AND MOD(rows_count * cols_count, 2) = 1
    )
  )
);

CREATE TABLE IF NOT EXISTS ad_frequency_configs (
  id TINYINT NOT NULL PRIMARY KEY,
  no_interstitial_before_level INT NOT NULL DEFAULT 4,
  interstitial_every_levels INT NOT NULL DEFAULT 4,
  max_interstitial_per_day INT NOT NULL DEFAULT 10,
  max_revive_per_level INT NOT NULL DEFAULT 1,
  banner_enabled_scenes JSON NOT NULL,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS player_progress (
  open_id VARCHAR(128) NOT NULL PRIMARY KEY,
  current_level INT NOT NULL DEFAULT 1,
  coins INT NOT NULL DEFAULT 0,
  hints INT NOT NULL DEFAULT 0,
  level_stars JSON NOT NULL,
  completed_levels JSON NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS level_results (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  open_id VARCHAR(128) NOT NULL,
  level_id INT NOT NULL,
  success TINYINT(1) NOT NULL,
  reason VARCHAR(32) NOT NULL,
  steps INT NOT NULL,
  mismatch_count INT NOT NULL,
  elapsed_ms INT NOT NULL,
  stars INT NOT NULL,
  coins_earned INT NOT NULL DEFAULT 0,
  used_hints INT NOT NULL DEFAULT 0,
  completed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_level_results_player (open_id, level_id),
  INDEX idx_level_results_level (level_id, stars, steps, elapsed_ms)
);

CREATE TABLE IF NOT EXISTS leaderboard_entries (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  open_id VARCHAR(128) NOT NULL,
  nickname VARCHAR(64) NOT NULL DEFAULT '',
  level_id INT NOT NULL,
  stars INT NOT NULL,
  steps INT NOT NULL,
  elapsed_ms INT NOT NULL,
  submitted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_leaderboard_rank (level_id, stars, steps, elapsed_ms, submitted_at)
);

CREATE TABLE IF NOT EXISTS game_events (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  event_id VARCHAR(128) NOT NULL,
  open_id VARCHAR(128) NOT NULL DEFAULT '',
  event_type VARCHAR(64) NOT NULL,
  payload JSON NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_game_events_event_id (event_id),
  INDEX idx_game_events_type_time (event_type, created_at),
  INDEX idx_game_events_player_time (open_id, created_at)
);
