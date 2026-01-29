-- +goose Up
-- +goose StatementBegin

CREATE TABLE users (
                       id UUID PRIMARY KEY,
                       email TEXT NOT NULL UNIQUE,
                       password_hash TEXT NOT NULL,
                       role TEXT NOT NULL,
                       created_at TIMESTAMP NOT NULL
);

CREATE TABLE pvzs (
                      id UUID PRIMARY KEY,
                      city TEXT NOT NULL,
                      created_at TIMESTAMP NOT NULL
);

CREATE TABLE receptions (
                            id UUID PRIMARY KEY,
                            pvz_id UUID REFERENCES pvzs(id),
                            started_at TIMESTAMP NOT NULL,
                            status TEXT NOT NULL
);

CREATE TABLE products (
                          id UUID PRIMARY KEY,
                          reception_id UUID REFERENCES receptions(id),
                          added_at TIMESTAMP NOT NULL,
                          type TEXT NOT NULL
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS receptions;
DROP TABLE IF EXISTS pvzs;
DROP TABLE IF EXISTS users;

-- +goose StatementEnd
