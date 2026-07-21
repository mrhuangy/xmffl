USE fpxxl;

SET @add_preview_again_count = (
  SELECT IF(
    NOT EXISTS (
      SELECT 1 FROM information_schema.columns
      WHERE table_schema = DATABASE()
        AND table_name = 'player_progress'
        AND column_name = 'preview_again_count'
    ),
    'ALTER TABLE player_progress ADD COLUMN preview_again_count INT NOT NULL DEFAULT 3 AFTER hints',
    'SELECT 1'
  )
);
PREPARE stmt FROM @add_preview_again_count;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @add_remove_pair_count = (
  SELECT IF(
    NOT EXISTS (
      SELECT 1 FROM information_schema.columns
      WHERE table_schema = DATABASE()
        AND table_name = 'player_progress'
        AND column_name = 'remove_pair_count'
    ),
    'ALTER TABLE player_progress ADD COLUMN remove_pair_count INT NOT NULL DEFAULT 3 AFTER preview_again_count',
    'SELECT 1'
  )
);
PREPARE stmt FROM @add_remove_pair_count;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

ALTER TABLE player_progress
  MODIFY COLUMN hints INT NOT NULL DEFAULT 3;

UPDATE player_progress
SET
  hints = IF(hints <= 0, 3, hints),
  preview_again_count = IF(preview_again_count <= 0, 3, preview_again_count),
  remove_pair_count = IF(remove_pair_count <= 0, 3, remove_pair_count);

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

INSERT IGNORE INTO tool_transactions (
  transaction_no, player_id, tool_type, change_amount, balance_after, source, ref_type, ref_id, note
)
SELECT CONCAT('tool_init_', player_id, '_hint'), player_id, 'hint', 3, hints, 'register', 'player', CAST(player_id AS CHAR), 'backfill initial tool charges'
FROM player_progress;

INSERT IGNORE INTO tool_transactions (
  transaction_no, player_id, tool_type, change_amount, balance_after, source, ref_type, ref_id, note
)
SELECT CONCAT('tool_init_', player_id, '_preview_again'), player_id, 'preview_again', 3, preview_again_count, 'register', 'player', CAST(player_id AS CHAR), 'backfill initial tool charges'
FROM player_progress;

INSERT IGNORE INTO tool_transactions (
  transaction_no, player_id, tool_type, change_amount, balance_after, source, ref_type, ref_id, note
)
SELECT CONCAT('tool_init_', player_id, '_remove_pair'), player_id, 'remove_pair', 3, remove_pair_count, 'register', 'player', CAST(player_id AS CHAR), 'backfill initial tool charges'
FROM player_progress;
