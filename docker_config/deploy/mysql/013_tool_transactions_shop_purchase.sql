ALTER TABLE tool_transactions
  MODIFY COLUMN source ENUM('register', 'use', 'ad_reward', 'shop_purchase', 'admin_adjust') NOT NULL COMMENT '来源：注册、使用、广告奖励、金币购买、后台调整';
