#!/usr/bin/env bash

set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
cd "$script_dir"

new_uuid() {
    if command -v uuidgen >/dev/null 2>&1; then
        uuidgen | tr '[:upper:]' '[:lower:]'
        return
    fi

    if [[ -r /proc/sys/kernel/random/uuid ]]; then
        tr -d '\n' < /proc/sys/kernel/random/uuid
        printf '\n'
        return
    fi

    printf 'uuidgen is required\n' >&2
    exit 1
}

start_db="${START_DB:-}"
if [[ -z "${DATABASE_URL:-}" ]]; then
    export DATABASE_URL="postgres://postgres@localhost:15433/products?sslmode=disable"
    start_db="${start_db:-1}"
else
    export DATABASE_URL
    start_db="${start_db:-0}"
fi

product_id="${PRODUCT_ID:-$(new_uuid)}"
user_id="${USER_ID:-$(new_uuid)}"
product_name="${PRODUCT_NAME:-Coffee}"
product_new_name="${PRODUCT_NEW_NAME:-Tea}"

if [[ "$start_db" == "1" ]]; then
    if ! command -v docker >/dev/null 2>&1; then
        printf 'docker is required when START_DB=1\n' >&2
        exit 1
    fi

    docker compose up -d db

    for _ in {1..30}; do
        if docker compose exec -T db pg_isready -U postgres -d products >/dev/null 2>&1; then
            break
        fi

        sleep 1
    done

    if ! docker compose exec -T db pg_isready -U postgres -d products >/dev/null 2>&1; then
        printf 'database is not ready\n' >&2
        exit 1
    fi
fi

printf 'DATABASE_URL=%s\n' "$DATABASE_URL"
printf 'PRODUCT_ID=%s\n' "$product_id"
printf 'USER_ID=%s\n' "$user_id"
printf 'PRODUCT_NAME=%s\n' "$product_name"
printf 'PRODUCT_NEW_NAME=%s\n' "$product_new_name"

for variant_dir in app/*/; do
    variant="$(basename "$variant_dir")"
    variant="${variant//-/_}"

    export APP_VARIANT="$variant"

    variant_product_id="${PRODUCT_ID:-$(new_uuid)}"
    variant_user_id="${USER_ID:-$(new_uuid)}"

    printf '\n========================================\n'
    printf 'Variant: %s\n' "$APP_VARIANT"
    printf '========================================\n'

    printf '\n1. History before product creation\n'
    go run ./cmd product history --id "$variant_product_id"

    printf '\n2. Create product\n'
    go run ./cmd product create --id "$variant_product_id" --user-id "$variant_user_id" --name "$product_name"

    printf '\n3. Get created product\n'
    go run ./cmd product get --id "$variant_product_id"

    printf '\n4. History after product creation\n'
    go run ./cmd product history --id "$variant_product_id"

    printf '\n5. Update product name\n'
    go run ./cmd product update --id "$variant_product_id" --user-id "$variant_user_id" --name "$product_new_name"

    printf '\n6. Get updated product\n'
    go run ./cmd product get --id "$variant_product_id"

    printf '\n7. History after product update\n'
    go run ./cmd product history --id "$variant_product_id"

    printf '\n'
done

printf 'All product lifecycles completed\n'
