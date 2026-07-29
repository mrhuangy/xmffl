-- ============================================================
-- MySQL 初始化脚本 — 在容器首次启动时自动执行
-- ============================================================

-- 确保使用正确的数据库
USE `dietitian`;

-- 设置时区
SET time_zone = '+08:00';

-- ── 营养师表 ──────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS `nutritionists` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name`       VARCHAR(50)     NOT NULL COMMENT '姓名',
  `avatar`     VARCHAR(255)             COMMENT '头像URL',
  `title`      VARCHAR(100)             COMMENT '职称',
  `intro`      TEXT                     COMMENT '简介',
  `account`    VARCHAR(50)     NOT NULL UNIQUE COMMENT '登录账号',
  `password`   VARCHAR(255)    NOT NULL COMMENT '密码(bcrypt)',
  `status`     TINYINT         NOT NULL DEFAULT 1 COMMENT '1正常 0禁用',
  `created_at` TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='营养师表';

-- ── 用户表 ────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS `users` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `openid`            VARCHAR(64)     NOT NULL UNIQUE COMMENT '微信openid',
  `nickname`          VARCHAR(50)              COMMENT '微信昵称',
  `avatar`            VARCHAR(255)             COMMENT '头像URL',
  `gender`            TINYINT         NOT NULL DEFAULT 0 COMMENT '0未知 1男 2女',
  `age`               TINYINT UNSIGNED         COMMENT '年龄',
  `height`            DECIMAL(5,1)             COMMENT '身高(cm)',
  `weight_init`       DECIMAL(5,2)             COMMENT '初始体重(kg)',
  `weight_goal`       DECIMAL(5,2)             COMMENT '目标体重(kg)',
  `health_goal`       TINYINT                  COMMENT '1减脂 2增肌 3维持 4改善亚健康',
  `diet_pref`         JSON                     COMMENT '饮食偏好',
  `medical_history`   JSON                     COMMENT '既往病史',
  `exercise_freq`     TINYINT                  COMMENT '运动频率(次/周)',
  `nutritionist_id`   BIGINT UNSIGNED          COMMENT '绑定营养师ID',
  `status`            TINYINT         NOT NULL DEFAULT 1 COMMENT '1跟进中 2已完成 0暂停',
  `created_at`        TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`        TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_nutritionist` (`nutritionist_id`),
  CONSTRAINT `fk_user_nutritionist` FOREIGN KEY (`nutritionist_id`) REFERENCES `nutritionists` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- ── 饮食方案表 ────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS `diet_plans` (
  `id`               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`          BIGINT UNSIGNED NOT NULL,
  `nutritionist_id`  BIGINT UNSIGNED NOT NULL,
  `title`            VARCHAR(100)    NOT NULL COMMENT '方案标题',
  `calorie_target`   INT UNSIGNED             COMMENT '每日目标热量(kcal)',
  `protein_ratio`    DECIMAL(4,1)             COMMENT '蛋白质比例(%)',
  `fat_ratio`        DECIMAL(4,1)             COMMENT '脂肪比例(%)',
  `carb_ratio`       DECIMAL(4,1)             COMMENT '碳水比例(%)',
  `content`          LONGTEXT                 COMMENT '方案详细内容',
  `forbidden_foods`  JSON                     COMMENT '禁忌食物',
  `version`          INT UNSIGNED    NOT NULL DEFAULT 1 COMMENT '版本号',
  `status`           TINYINT         NOT NULL DEFAULT 0 COMMENT '0草稿 1已发布',
  `published_at`     TIMESTAMP                COMMENT '发布时间',
  `created_at`       TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`       TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_nutritionist_id` (`nutritionist_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='饮食方案表';

-- ── 打卡记录表 ────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS `checkins` (
  `id`           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`      BIGINT UNSIGNED NOT NULL,
  `checkin_date` DATE            NOT NULL COMMENT '打卡日期',
  `meal_type`    TINYINT         NOT NULL COMMENT '1早餐 2午餐 3晚餐 4加餐',
  `description`  TEXT                     COMMENT '饮食描述',
  `images`       JSON                     COMMENT '图片URL列表',
  `portion_size` TINYINT                  COMMENT '1少 2中 3多',
  `meal_time`    TIME                     COMMENT '用餐时间',
  `created_at`   TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`   TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_date` (`user_id`, `checkin_date`),
  UNIQUE KEY `uq_user_date_meal` (`user_id`, `checkin_date`, `meal_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='每日打卡记录表';

-- ── 营养师点评表 ──────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS `reviews` (
  `id`               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`          BIGINT UNSIGNED NOT NULL,
  `nutritionist_id`  BIGINT UNSIGNED NOT NULL,
  `checkin_date`     DATE            NOT NULL COMMENT '点评针对的日期',
  `content`          TEXT            NOT NULL COMMENT '点评内容',
  `rating`           TINYINT                  COMMENT '评分 1-5',
  `is_read`          TINYINT         NOT NULL DEFAULT 0 COMMENT '0未读 1已读',
  `reviewed_at`      TIMESTAMP                COMMENT '点评时间',
  `created_at`       TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`       TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_date` (`user_id`, `checkin_date`),
  UNIQUE KEY `uq_review_user_date` (`user_id`, `checkin_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='营养师点评表';

-- ── 体重记录表 ────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS `weight_records` (
  `id`           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`      BIGINT UNSIGNED NOT NULL,
  `weight`       DECIMAL(5,2)    NOT NULL COMMENT '体重(kg)',
  `bmi`          DECIMAL(4,1)             COMMENT 'BMI值(后端计算写入)',
  `record_date`  DATE            NOT NULL COMMENT '记录日期',
  `note`         VARCHAR(255)             COMMENT '备注',
  `created_at`   TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`   TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_date` (`user_id`, `record_date`),
  UNIQUE KEY `uq_user_record_date` (`user_id`, `record_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='体重记录表';

-- ── 初始营养师账号（密码: Admin@123，生产务必修改）────────────
INSERT IGNORE INTO `nutritionists` (`name`, `account`, `password`, `title`, `status`)
VALUES ('系统管理员', 'admin', '$2y$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', '首席营养师', 1);
-- 注: 上方 bcrypt hash 对应明文 password，正式上线前请用 artisan tinker 重置
