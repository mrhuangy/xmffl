USE fpxxl;

INSERT INTO ad_frequency_configs (
  id, no_interstitial_before_level, interstitial_every_levels, max_interstitial_per_day,
  max_revive_per_level, banner_enabled_scenes
) VALUES (1, 4, 4, 10, 1, JSON_ARRAY('home', 'result'))
ON DUPLICATE KEY UPDATE id = id;

INSERT INTO level_configs (
  level_id, rows_count, cols_count, pair_count, mode, theme_id, initial_preview_ms,
  flip_back_delay_ms, level_time_limit_seconds, max_mismatch_count,
  excellent_step_threshold, normal_step_threshold, excellent_time_threshold, normal_time_threshold,
  enabled
) VALUES
(1, 2, 2, 2, 'normal', 'animal', 2000, 700, 45, 6, 3, 5, 20, 35, 1),
(2, 2, 2, 2, 'normal', 'animal', 2000, 700, 45, 6, 3, 5, 20, 35, 1),
(3, 3, 4, 6, 'normal', 'animal', 2000, 700, 90, 10, 9, 14, 45, 75, 1),
(4, 3, 4, 6, 'normal', 'animal', 2000, 700, 90, 10, 9, 14, 45, 75, 1),
(5, 3, 4, 6, 'normal', 'animal', 2000, 700, 90, 10, 9, 14, 45, 75, 1),
(6, 4, 4, 8, 'normal', 'animal', 2000, 700, 120, 12, 12, 18, 70, 105, 1),
(7, 4, 4, 8, 'normal', 'animal', 2000, 700, 120, 12, 12, 18, 70, 105, 1),
(8, 4, 4, 8, 'normal', 'animal', 2000, 700, 120, 12, 12, 18, 70, 105, 1),
(9, 4, 4, 8, 'normal', 'animal', 2000, 700, 120, 12, 12, 18, 70, 105, 1),
(10, 4, 4, 8, 'normal', 'animal', 2000, 700, 120, 12, 12, 18, 70, 105, 1),
(11, 4, 4, 8, 'normal', 'animal', 2000, 700, 120, 12, 12, 18, 70, 105, 1),
(12, 4, 4, 8, 'normal', 'animal', 2000, 700, 120, 12, 12, 18, 70, 105, 1),
(13, 4, 4, 8, 'normal', 'animal', 2000, 700, 120, 12, 12, 18, 70, 105, 1),
(14, 4, 4, 8, 'normal', 'animal', 2000, 700, 120, 12, 12, 18, 70, 105, 1),
(15, 4, 4, 8, 'normal', 'animal', 2000, 700, 120, 12, 12, 18, 70, 105, 1),
(16, 4, 4, 8, 'normal', 'animal', 2000, 700, 120, 12, 12, 18, 70, 105, 1),
(17, 4, 4, 8, 'normal', 'animal', 2000, 700, 120, 12, 12, 18, 70, 105, 1),
(18, 4, 4, 8, 'normal', 'animal', 2000, 700, 120, 12, 12, 18, 70, 105, 1),
(19, 4, 4, 8, 'normal', 'animal', 2000, 700, 120, 12, 12, 18, 70, 105, 1),
(20, 4, 4, 8, 'normal', 'animal', 2000, 700, 120, 12, 12, 18, 70, 105, 1)
ON DUPLICATE KEY UPDATE level_id = level_id;
