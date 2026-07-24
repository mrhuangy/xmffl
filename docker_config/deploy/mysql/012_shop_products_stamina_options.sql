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
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  product_type = VALUES(product_type),
  currency_type = VALUES(currency_type),
  currency_amount = VALUES(currency_amount),
  grant_type = VALUES(grant_type),
  grant_amount = VALUES(grant_amount),
  daily_buy_limit = VALUES(daily_buy_limit),
  enabled = VALUES(enabled),
  sort_order = VALUES(sort_order);
