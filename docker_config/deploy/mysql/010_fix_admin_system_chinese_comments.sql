USE fpxxl;
SET NAMES utf8mb4;

ALTER TABLE admin_users COMMENT='后台管理员账号表';
ALTER TABLE admin_users
  MODIFY COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '管理员自增 ID',
  MODIFY COLUMN username VARCHAR(64) NOT NULL COMMENT '登录用户名',
  MODIFY COLUMN email VARCHAR(255) NULL COMMENT '邮箱地址',
  MODIFY COLUMN password_hash VARCHAR(255) NOT NULL COMMENT '密码哈希',
  MODIFY COLUMN display_name VARCHAR(64) NOT NULL DEFAULT '' COMMENT '显示名称',
  MODIFY COLUMN role ENUM('owner','operator','viewer') NOT NULL DEFAULT 'operator' COMMENT '角色：owner 所有者，operator 运营，viewer 只读',
  MODIFY COLUMN permissions JSON NULL COMMENT '细粒度权限 JSON',
  MODIFY COLUMN status ENUM('active','disabled','locked') NOT NULL DEFAULT 'active' COMMENT '账号状态',
  MODIFY COLUMN failed_login_attempts INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '连续登录失败次数',
  MODIFY COLUMN locked_until DATETIME NULL COMMENT '锁定截止时间',
  MODIFY COLUMN password_changed_at DATETIME NULL COMMENT '密码最后修改时间',
  MODIFY COLUMN last_login_at DATETIME NULL COMMENT '最后登录时间',
  MODIFY COLUMN last_login_ip VARCHAR(45) NOT NULL DEFAULT '' COMMENT '最后登录 IP',
  MODIFY COLUMN created_by BIGINT UNSIGNED NULL COMMENT '创建人管理员 ID',
  MODIFY COLUMN created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  MODIFY COLUMN updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间';

ALTER TABLE system_controls COMMENT='系统全局控制配置表';
ALTER TABLE system_controls
  MODIFY COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '系统控制项自增 ID',
  MODIFY COLUMN control_key VARCHAR(128) NOT NULL COMMENT '系统控制项唯一键，例如 system.maintenance_mode',
  MODIFY COLUMN control_group VARCHAR(64) NOT NULL DEFAULT 'general' COMMENT '控制项分组，例如 system、client、game、notice',
  MODIFY COLUMN name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '控制项显示名称',
  MODIFY COLUMN description VARCHAR(512) NOT NULL DEFAULT '' COMMENT '控制项说明',
  MODIFY COLUMN value_type ENUM('bool','int','decimal','string','json') NOT NULL DEFAULT 'string' COMMENT '配置值类型',
  MODIFY COLUMN value_text TEXT NULL COMMENT '非 JSON 类型当前值，统一用文本保存',
  MODIFY COLUMN value_json JSON NULL COMMENT 'JSON 类型当前值',
  MODIFY COLUMN default_value_text TEXT NULL COMMENT '非 JSON 类型默认值',
  MODIFY COLUMN default_value_json JSON NULL COMMENT 'JSON 类型默认值',
  MODIFY COLUMN enabled TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用',
  MODIFY COLUMN is_public TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否允许下发给客户端',
  MODIFY COLUMN sort_order INT NOT NULL DEFAULT 0 COMMENT '排序值',
  MODIFY COLUMN version INT UNSIGNED NOT NULL DEFAULT 1 COMMENT '版本号，配置变更时递增',
  MODIFY COLUMN effective_from DATETIME NULL COMMENT '生效开始时间，为空表示立即生效',
  MODIFY COLUMN effective_until DATETIME NULL COMMENT '生效结束时间，为空表示长期生效',
  MODIFY COLUMN created_by BIGINT UNSIGNED NULL COMMENT '创建人管理员 ID',
  MODIFY COLUMN updated_by BIGINT UNSIGNED NULL COMMENT '最后更新人管理员 ID',
  MODIFY COLUMN created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  MODIFY COLUMN updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间';

UPDATE system_controls SET
  name = '维护模式',
  description = '开启后可用于后台或接口统一拦截非管理流量'
WHERE control_key = 'system.maintenance_mode';

UPDATE system_controls SET
  name = '最低客户端版本',
  description = '低于该版本时可提示升级'
WHERE control_key = 'client.min_version';

UPDATE system_controls SET
  name = '强制更新',
  description = '开启后客户端可根据版本策略强制升级'
WHERE control_key = 'client.force_update';

UPDATE system_controls SET
  name = '玩法功能开关',
  description = '全局玩法功能开关集合'
WHERE control_key = 'game.feature_flags';

UPDATE system_controls SET
  name = '首页公告',
  description = '小游戏首页公告或运营提示'
WHERE control_key = 'notice.home_banner';
