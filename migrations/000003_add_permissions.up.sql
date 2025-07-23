CREATE TABLE IF NOT EXISTS permissions (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    code text NOT NULL
);

CREATE TABLE IF NOT EXISTS users_permissions (
    user_id uuid NOT NULL REFERENCES users ON DELETE CASCADE,
    permission_id uuid NOT NULL REFERENCES permissions ON DELETE CASCADE,
    PRIMARY KEY (user_id, permission_id)
);

INSERT INTO permissions (code)
VALUES ('admin')
