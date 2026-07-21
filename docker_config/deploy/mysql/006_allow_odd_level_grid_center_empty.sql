USE fpxxl;

SET @drop_level_grid_check = (
  SELECT IF(
    EXISTS (
      SELECT 1 FROM information_schema.check_constraints
      WHERE constraint_schema = DATABASE()
        AND constraint_name = 'chk_level_grid'
    ),
    'ALTER TABLE level_configs DROP CHECK chk_level_grid',
    'SELECT 1'
  )
);
PREPARE stmt FROM @drop_level_grid_check;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @drop_level_configs_grid_check = (
  SELECT IF(
    EXISTS (
      SELECT 1 FROM information_schema.check_constraints
      WHERE constraint_schema = DATABASE()
        AND constraint_name = 'chk_level_configs_grid'
    ),
    'ALTER TABLE level_configs DROP CHECK chk_level_configs_grid',
    'SELECT 1'
  )
);
PREPARE stmt FROM @drop_level_configs_grid_check;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

ALTER TABLE level_configs
  ADD CONSTRAINT chk_level_grid_card_slots
  CHECK (
    rows_count * cols_count = pair_count * 2
    OR (
      rows_count * cols_count = pair_count * 2 + 1
      AND MOD(rows_count * cols_count, 2) = 1
    )
  );
