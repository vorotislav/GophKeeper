CREATE TABLE passwords(
    id BIGINT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY, -- уникальный идентификатор записи
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE, -- идентификатор пользователя
    title TEXT NOT NULL, -- название записи
    login TEXT NOT NULL, -- логин
    password TEXT NOT NULL, -- пароль
    site TEXT, -- сайт
    note TEXT, -- заметка к записи
    expiration_at TIMESTAMP WITH TIME ZONE, -- дата истечения срока действия пароля
    created_at TIMESTAMP WITH TIME ZONE, -- отметка времени создания записи
    updated_at TIMESTAMP WITH TIME ZONE -- отметка времени обновления записи
);