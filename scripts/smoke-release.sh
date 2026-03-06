#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

echo "[smoke] Validate encoding"
pwsh -File ./scripts/check-encoding.ps1

echo "[smoke] Install JS deps (shared/client/scenarios)"
npm --prefix shared ci --no-audit --no-fund
npm --prefix client ci --no-audit --no-fund
npm --prefix scenarios install --no-audit --no-fund --package-lock=false

echo "[smoke] Verify generated special-effect contract"
node ./scripts/generate-special-effect-contract.mjs --check

echo "[smoke] Go tests"
(
  cd server-go
  go test ./...
)

echo "[smoke] Shared build + tests"
npm --prefix shared run build
npm --prefix shared run test

echo "[smoke] Scenarios tests + build"
npm --prefix scenarios run test
npm --prefix scenarios run build

echo "[smoke] Client typecheck + build"
npm --prefix client run typecheck
npm --prefix client run build

echo "[smoke] OK"
