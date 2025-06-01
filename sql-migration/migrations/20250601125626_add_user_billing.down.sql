-- 20250601125626_add_user_billing.down.sql
DROP INDEX IF EXISTS idx_user_billing_is_default;
DROP INDEX IF EXISTS idx_user_billing_user_id;
DROP TABLE IF EXISTS user_billing;