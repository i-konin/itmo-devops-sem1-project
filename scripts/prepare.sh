#!/bin/bash

go get github.com/lib/pq

host="${POSTGRES_HOST}"
port="${POSTGRES_PORT}"
user="${POSTGRES_USER}"
password="${POSTGRES_PASSWORD}"
dbname="${POSTGRES_DB}"

export PGPASSWORD="$password"

psql -h "$host" -p "$port" -U "$user" -d "$dbname" -c "
CREATE TABLE IF NOT EXISTS prices (
    id SERIAL PRIMARY KEY,
    create_date DATE,
    name TEXT,
    category TEXT,
    price NUMERIC
);"

unset PGPASSWORD