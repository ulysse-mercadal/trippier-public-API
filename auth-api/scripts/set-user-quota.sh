#!/usr/bin/env bash
# Set a user's token quota by email (or user id) on the deployed auth-api.
# Reads INTERNAL_SECRET from .env in the current directory, signs the request,
# and posts to the in-network admin endpoint via a throwaway curl container.
#
# Usage (run from ~/trippier on the VPS):
#   ./auth-api/scripts/set-user-quota.sh <email> <tokens_limit> [reset_interval_secs]
#
# Examples:
#   ./set-user-quota.sh alice@example.com 100000
#   ./set-user-quota.sh alice@example.com 100000 2592000

set -euo pipefail

if [[ $# -lt 2 ]]; then
    echo "usage: $0 <email> <tokens_limit> [reset_interval_secs]" >&2
    exit 1
fi

EMAIL="$1"
LIMIT="$2"
INTERVAL="${3:-0}"   # 0 = keep existing interval server-side

if [[ ! -f .env ]]; then
    echo "error: .env not found in $(pwd)" >&2
    exit 1
fi

# shellcheck disable=SC1091
INTERNAL_SECRET=$(grep -E '^INTERNAL_SECRET=' .env | head -n1 | cut -d= -f2-)
if [[ -z "${INTERNAL_SECRET:-}" ]]; then
    echo "error: INTERNAL_SECRET missing from .env" >&2
    exit 1
fi

NETWORK=$(sudo docker network ls --format '{{.Name}}' | grep -E '_default$' | grep -i trippier | head -n1)
if [[ -z "$NETWORK" ]]; then
    echo "error: could not find trippier docker network" >&2
    exit 1
fi

TS=$(date +%s)
SIG=$(printf '%s' "$TS" | openssl dgst -sha256 -hmac "$INTERNAL_SECRET" -binary | xxd -p -c 256)

BODY=$(printf '{"email":"%s","tokens_limit":%s,"tokens_reset_interval_secs":%s}' "$EMAIL" "$LIMIT" "$INTERVAL")

sudo docker run --rm --network "$NETWORK" curlimages/curl:latest \
    -fsS -X POST http://auth-api:8081/internal/admin/user-quota \
    -H "X-Internal-Auth: ${TS}.${SIG}" \
    -H "Content-Type: application/json" \
    -d "$BODY"
echo
