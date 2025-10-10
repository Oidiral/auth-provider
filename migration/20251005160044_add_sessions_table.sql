-- +goose Up
-- +goose StatementBegin
CREATE TABLE sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token VARCHAR(255),
    expires_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now()
);
CREATE INDEX ON sessions (user_id);
CREATE INDEX ON sessions (refresh_token);
CREATE INDEX ON sessions (expires_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS sessions;
-- +goose StatementEnd
