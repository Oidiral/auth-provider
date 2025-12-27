-- +goose Up
-- +goose StatementBegin

-- Функция для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Триггер для таблицы users (UPDATE)
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Триггер для таблицы roles (UPDATE)
CREATE TRIGGER update_roles_updated_at
    BEFORE UPDATE ON roles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Функция для мягкого удаления (soft delete)
CREATE OR REPLACE FUNCTION soft_delete_user()
RETURNS TRIGGER AS $$
BEGIN
    -- Вместо физического удаления, устанавливаем deleted_at
    IF OLD.deleted_at IS NULL THEN
        UPDATE users
        SET deleted_at = now()
        WHERE id = OLD.id;
        RETURN NULL; -- Отменяем физическое удаление
    ELSE
        -- Если уже помечено как удаленное, разрешаем физическое удаление
        RETURN OLD;
    END IF;
END;
$$ language 'plpgsql';

-- Триггер для мягкого удаления пользователей
CREATE TRIGGER soft_delete_users
    BEFORE DELETE ON users
    FOR EACH ROW
    EXECUTE FUNCTION soft_delete_user();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Удаление триггеров
DROP TRIGGER IF EXISTS soft_delete_users ON users;
DROP TRIGGER IF EXISTS update_roles_updated_at ON roles;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Удаление функций
DROP FUNCTION IF EXISTS soft_delete_user();
DROP FUNCTION IF EXISTS update_updated_at_column();

-- +goose StatementEnd