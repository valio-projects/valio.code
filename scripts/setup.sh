#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
if [ -e .env ]; then echo 'Existing .env preserved.'; exit 0; fi
umask 077
for name in SURREAL_ROOT_PASSWORD VALIO_API_DB_PASSWORD VALIO_WORKER_DB_PASSWORD VALIO_API_TOKEN; do
  printf '%s=%s\n' "$name" "$(openssl rand -hex 32)" >> .env
done
echo 'Generated local credentials in ignored .env.'
