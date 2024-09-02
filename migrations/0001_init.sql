-- +goose Up
-- +goose StatementBegin
CREATE TABLE messages
(
    id   SERIAL PRIMARY KEY,
    text TEXT NOT NULL
);
-- +goose StatementEnd
