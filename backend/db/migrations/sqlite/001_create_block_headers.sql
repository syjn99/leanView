-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS block_headers (
    slot INTEGER NOT NULL,
    proposer_index INTEGER NOT NULL,
    parent_root BLOB NOT NULL,
    state_root BLOB NOT NULL,
    body_root BLOB NOT NULL,
    CONSTRAINT block_headers_pkey PRIMARY KEY (slot)
);

CREATE INDEX IF NOT EXISTS block_headers_proposer_idx 
    ON block_headers (proposer_index ASC);

CREATE INDEX IF NOT EXISTS block_headers_parent_root_idx 
    ON block_headers (parent_root);

CREATE INDEX IF NOT EXISTS block_headers_state_root_idx 
    ON block_headers (state_root);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS block_headers;
-- +goose StatementEnd