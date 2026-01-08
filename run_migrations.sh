#!/bin/bash
# Manual migration runner for Railway PostgreSQL
# Usage: ./run_migrations.sh <DATABASE_URL>

if [ -z "$1" ]; then
    echo "Usage: ./run_migrations.sh <DATABASE_URL>"
    echo "Example: ./run_migrations.sh postgresql://user:pass@host:port/dbname"
    exit 1
fi

DATABASE_URL="$1"

echo "Running migrations against: $DATABASE_URL"
echo ""

# Run first migration
echo "=== Running 001_initial.up.sql ==="
psql "$DATABASE_URL" -f migrations/postgres/001_initial.up.sql
if [ $? -ne 0 ]; then
    echo "ERROR: Migration 001_initial failed"
    exit 1
fi

echo ""
echo "=== Running 002_tables.up.sql ==="
psql "$DATABASE_URL" -f migrations/postgres/002_tables.up.sql
if [ $? -ne 0 ]; then
    echo "ERROR: Migration 002_tables failed"
    exit 1
fi

echo ""
echo "✅ All migrations completed successfully!"
echo ""
echo "Verifying tables..."
psql "$DATABASE_URL" -c "\dt"

