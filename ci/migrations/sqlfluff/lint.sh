#!/bin/sh
set -e

# Change to the directory of this script so that relative paths resolve correctly
cd "$(dirname "$0")"

docker compose up --no-color --quiet-pull
