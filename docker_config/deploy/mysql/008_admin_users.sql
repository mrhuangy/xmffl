USE fpxxl;
SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS admin_users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '管理员自增 ID' PRIMARY KEY,
  username VARCHAR(64) NOT NULL COMMENT '登录用户名',
  email VARCHAR(255) NULL COMMENT '邮箱地址',
  password_hash VARCHAR(255) NOT NULL COMMENT '密码哈希',
  display_name VARCHAR(64) NOT NULL DEFAULT '' COMMENT '显示名称',
  role ENUM('owner', 'operator', 'viewer') NOT NULL DEFAULT 'operator' COMMENT '角色：owner 所有者，operator 运营，viewer 只读',
  permissions JSON NULL COMMENT '细粒度权限 JSON',
  status ENUM('active', 'disabled', 'locked') NOT NULL DEFAULT 'active' COMMENT '账号状态',
  failed_login_attempts INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '连续登录失败次数',
  locked_until DATETIME NULL COMMENT '锁定截止时间',
  password_changed_at DATETIME NULL COMMENT '密码最后修改时间',
  last_login_at DATETIME NULL COMMENT '最后登录时间',
  last_login_ip VARCHAR(45) NOT NULL DEFAULT '' COMMENT '最后登录 IP',
  created_by BIGINT UNSIGNED NULL COMMENT '创建人管理员 ID',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  UNIQUE KEY uk_admin_users_username (username),
  UNIQUE KEY uk_admin_users_email (email),
  KEY idx_admin_users_status_role (status, role),
  KEY idx_admin_users_locked_until (locked_until),
  CONSTRAINT fk_admin_users_created_by FOREIGN KEY (created_by) REFERENCES admin_users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='后台管理员账号表';
