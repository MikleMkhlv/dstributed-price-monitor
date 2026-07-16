-- migrations/000001_create_table_data.up.sql
CREATE TABLE IF NOT EXISTS data (
    id SERIAL PRIMARY KEY,
    object_id INTEGER NOT NULL UNIQUE,
    object_data TEXT NOT NULL
);
