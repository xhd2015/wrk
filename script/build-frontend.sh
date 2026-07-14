#!/usr/bin/env bash
# Build wrk-react and copy dist into wrkcli/web/dist for //go:embed.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
WRK_REACT="$ROOT/wrk-react"
DST="$ROOT/wrkcli/web/dist"

if ! command -v bun >/dev/null 2>&1; then
  echo "bun is not installed; see https://bun.sh/docs/installation" >&2
  exit 1
fi

if [[ ! -d "$WRK_REACT/node_modules" ]]; then
  (cd "$WRK_REACT" && bun install)
fi

(cd "$WRK_REACT" && bun run build)

rm -rf "$DST"
mkdir -p "$DST"
cp -R "$WRK_REACT/dist/." "$DST/"
echo "built frontend → $DST"