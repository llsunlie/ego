CREATE TABLE chat_sessions (
  id                 UUID PRIMARY KEY,
  user_id            UUID NOT NULL,
  star_id            UUID NOT NULL,
  created_at         TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_chat_sessions_user ON chat_sessions(user_id);

CREATE TABLE chat_messages (
  id                 UUID PRIMARY KEY,
  user_id            UUID NOT NULL,
  session_id         UUID NOT NULL,
  role               VARCHAR(10) NOT NULL,
  content            TEXT NOT NULL,
  referenced_moments JSONB,
  created_at         TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_chat_messages_session ON chat_messages(session_id, created_at);
