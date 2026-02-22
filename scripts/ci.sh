#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APP_ID="${PB_APP_ID:-com.protocolbunker.host}"
RUN_DEVICE_TESTS="${RUN_DEVICE_TESTS:-0}"

resolve_sdk_root() {
  if [[ -n "${ANDROID_SDK_ROOT:-}" ]]; then
    printf '%s' "$ANDROID_SDK_ROOT"
    return
  fi
  if [[ -n "${ANDROID_HOME:-}" ]]; then
    printf '%s' "$ANDROID_HOME"
    return
  fi

  local local_props="$ROOT_DIR/android-app/local.properties"
  if [[ -f "$local_props" ]]; then
    local raw
    raw="$(grep -E '^sdk\.dir=' "$local_props" | head -n1 | cut -d'=' -f2- || true)"
    if [[ -n "$raw" ]]; then
      raw="${raw//\\:/:}"
      raw="${raw//\\\\/\/}"
      printf '%s' "$raw"
      return
    fi
  fi

  if [[ -d "$ROOT_DIR/android-app/.android-sdk" ]]; then
    printf '%s' "$ROOT_DIR/android-app/.android-sdk"
    return
  fi
}

resolve_aapt2_override_arg() {
  local sdk_root="$1"
  [[ -z "$sdk_root" ]] && return
  local candidate=""
  for candidate in \
    "$sdk_root/build-tools/35.0.0/aapt2" \
    "$sdk_root/build-tools/34.0.0/aapt2" \
    "$sdk_root/build-tools/35.0.0/aapt2.exe" \
    "$sdk_root/build-tools/34.0.0/aapt2.exe"; do
    if [[ -f "$candidate" ]]; then
      printf '%s' "-Pandroid.aapt2FromMavenOverride=$candidate"
      return
    fi
  done
}

aapt2_arg="$(resolve_aapt2_override_arg "$(resolve_sdk_root)")"

echo "==> go test"
pushd "$ROOT_DIR/server-go" >/dev/null
go test ./...
./scripts/build-android-binaries.sh
popd >/dev/null

echo "==> android assembleDebug/assembleRelease"
pushd "$ROOT_DIR/android-app" >/dev/null
if [[ -n "${aapt2_arg:-}" ]]; then
  echo "Using local aapt2 override: $aapt2_arg"
  ./gradlew assembleDebug assembleRelease "$aapt2_arg"
else
  ./gradlew assembleDebug assembleRelease
fi

if [[ "$RUN_DEVICE_TESTS" == "1" ]]; then
  echo "==> connectedDebugAndroidTest"
  if [[ -n "${aapt2_arg:-}" ]]; then
    ./gradlew connectedDebugAndroidTest "$aapt2_arg"
  else
    ./gradlew connectedDebugAndroidTest
  fi
  echo "==> smoke-e2e"
  ./scripts/smoke-e2e.sh "$APP_ID"
else
  echo "==> device tests skipped (set RUN_DEVICE_TESTS=1 to enable)"
fi

popd >/dev/null

echo "CI pipeline commands finished"
