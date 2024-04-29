#!/bin/sh
set -e

cd "$(dirname "$0")"

date="$(date -u "+%Y-%m-%dT%H-%M-%SZ")"
export DATE="$date"

git_hash="$(git rev-parse HEAD)"
export GIT_HASH="$git_hash"

docker compose build --quiet

echo "ghcr.io/glizzus/trf/ministry:build-${date}"
