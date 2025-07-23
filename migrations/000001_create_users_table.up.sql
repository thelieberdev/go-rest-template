CREATE EXTENSION citext;

CREATE TABLE IF NOT EXISTS users (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  email citext UNIQUE NOT NULL,
  first_name citext NOT NULL,
  last_name citext NOT NULL,
  password_hash bytea NOT NULL,
  created_at timestamp with time zone NOT NULL DEFAULT NOW(),
  last_updated timestamp with time zone NOT NULL DEFAULT NOW(),
  activated boolean NOT NULL
);
