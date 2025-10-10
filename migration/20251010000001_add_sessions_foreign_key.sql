-- +goose Up
-- +goose StatementBegin
-- Сначала удаляем "висячие" сессии (если есть)
DELETE FROM sessions WHERE user_id NOT IN (SELECT id FROM users);

-- Добавляем внешний ключ с каскадным удалением
ALTER TABLE sessions
ADD CONSTRAINT fk_sessions_user_id
FOREIGN KEY (user_id)
REFERENCES users(id)
ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE sessions DROP CONSTRAINT IF EXISTS fk_sessions_user_id;
-- +goose StatementEnd