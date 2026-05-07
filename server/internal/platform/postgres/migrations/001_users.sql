CREATE TABLE users (
  id            UUID PRIMARY KEY,
  account       VARCHAR(100) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX idx_users_account ON users(account);
