-- +goose Up
-- +goose StatementBegin
CREATE TABLE messages
(
    id   SERIAL PRIMARY KEY,
    nickname TEXT NOT NULL,
    text TEXT NOT NULL
);
-- +goose StatementEnd
