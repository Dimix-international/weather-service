-- +goose Up
-- +goose StatementBegin
CREATE TABLE reading (
    id SERIAL,
    name VARCHAR(255) NOT NULL,
    temperature DOUBLE PRECISION NOT NULL, 
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS reading;
-- +goose StatementEnd
