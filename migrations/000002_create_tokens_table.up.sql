CREATE TABLE IF NOT EXISTS tokens (
  hash bytea PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users ON DELETE CASCADE,
  expiry timestamp with time zone NOT NULL,
  scope text NOT NULL
);
