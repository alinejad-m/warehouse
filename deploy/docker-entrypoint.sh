#!/bin/sh
set -e
mkdir -p /data/repo /data/downloads 2>/dev/null || true
exec "$@"
