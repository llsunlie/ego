ALTER TABLE users RENAME COLUMN account TO phone;
ALTER INDEX idx_users_account RENAME TO idx_users_phone;
