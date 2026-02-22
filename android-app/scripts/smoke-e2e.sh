#!/usr/bin/env bash

set -euo pipefail

PACKAGE="${1:-com.protocolbunker.host}"

resolve_sdk_root() {
  if [[ -n "${ANDROID_SDK_ROOT:-}" ]]; then
    printf '%s' "$ANDROID_SDK_ROOT"
    return
  fi
  if [[ -n "${ANDROID_HOME:-}" ]]; then
    printf '%s' "$ANDROID_HOME"
    return
  fi

  local root_dir
  root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
  local local_props="$root_dir/local.properties"
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
  if [[ -d "$root_dir/.android-sdk" ]]; then
    printf '%s' "$root_dir/.android-sdk"
    return
  fi
}

ADB_BIN="${ADB_BIN:-adb}"
if ! command -v "$ADB_BIN" >/dev/null 2>&1; then
  sdk_root="$(resolve_sdk_root || true)"
  if [[ -n "${sdk_root:-}" && -x "$sdk_root/platform-tools/adb" ]]; then
    ADB_BIN="$sdk_root/platform-tools/adb"
  elif [[ -n "${sdk_root:-}" && -x "$sdk_root/platform-tools/adb.exe" ]]; then
    ADB_BIN="$sdk_root/platform-tools/adb.exe"
  else
    echo "adb not found. Install platform-tools and configure ANDROID_SDK_ROOT/ANDROID_HOME (or local.properties)." >&2
    exit 1
  fi
fi

if [[ "$("$ADB_BIN" devices | grep -cE $'\tdevice$')" -eq 0 ]]; then
  echo "No connected Android device/emulator found (adb devices)." >&2
  exit 1
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GRADLEW="$ROOT_DIR/gradlew"
if [[ ! -x "$GRADLEW" ]]; then
  GRADLEW="$ROOT_DIR/gradlew.bat"
fi
if [[ ! -f "$GRADLEW" ]]; then
  echo "Gradle wrapper not found in $ROOT_DIR" >&2
  exit 1
fi

echo "1) Uninstalling existing app to avoid signature mismatch..."
"$ADB_BIN" uninstall "$PACKAGE" >/dev/null 2>&1 || true

run_test() {
  local test_class="$1"
  echo "2) Running $test_class ..."
  "$GRADLEW" :app:connectedDebugAndroidTest "-Pandroid.testInstrumentationRunnerArguments.class=$test_class"
}

run_test "com.protocolbunker.host.ServerHostInstrumentedTest#startStopService_updatesRuntimeState"
run_test "com.protocolbunker.host.ServerHostInstrumentedTest#serverRemainsActiveAfterAppBackgrounded"
run_test "com.protocolbunker.host.ServerHostInstrumentedTest#healthEndpointRespondsAfterStart"

echo "Smoke e2e passed"
