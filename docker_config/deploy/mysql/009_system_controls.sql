USE fpxxl;
SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS system_controls (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '系统控制项自增 ID' PRIMARY KEY,
  control_key VARCHAR(128) NOT NULL COMMENT '系统控制项唯一键，例如 system.maintenance_mode',
  control_group VARCHAR(64) NOT NULL DEFAULT 'general' COMMENT '控制项分组，例如 system、client、game、notice',
  name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '控制项显示名称',
  description VARCHAR(512) NOT NULL DEFAULT '' COMMENT '控制项说明',
  value_type ENUM('bool', 'int', 'decimal', 'string', 'json') NOT NULL DEFAULT 'string' COMMENT '配置值类型',
  value_text TEXT NULL COMMENT '非 JSON 类型当前值，统一用文本保存',
  value_json JSON NULL COMMENT 'JSON 类型当前值',
  default_value_text TEXT NULL COMMENT '非 JSON 类型默认值',
  default_value_json JSON NULL COMMENT 'JSON 类型默认值',
  enabled TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用',
  is_public TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否允许下发给客户端',
  sort_order INT NOT NULL DEFAULT 0 COMMENT '排序值',
  version INT UNSIGNED NOT NULL DEFAULT 1 COMMENT '版本号，配置变更时递增',
  effective_from DATETIME NULL COMMENT '生效开始时间，为空表示立即生效',
  effective_until DATETIME NULL COMMENT '生效结束时间，为空表示长期生效',
  created_by BIGINT UNSIGNED NULL COMMENT '创建人管理员 ID',
  updated_by BIGINT UNSIGNED NULL COMMENT '最后更新人管理员 ID',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  UNIQUE KEY uk_system_controls_key (control_key),
  KEY idx_system_controls_group_enabled (control_group, enabled, sort_order),
  KEY idx_system_controls_public_enabled (is_public, enabled, sort_order),
  KEY idx_system_controls_effective (effective_from, effective_until),
  KEY idx_system_controls_updated_at (updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统全局控制配置表';

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
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  description = VALUES(description),
  value_type = VALUES(value_type),
  default_value_text = VALUES(default_value_text),
  default_value_json = VALUES(default_value_json),
  enabled = VALUES(enabled),
  is_public = VALUES(is_public),
  sort_order = VALUES(sort_order);
