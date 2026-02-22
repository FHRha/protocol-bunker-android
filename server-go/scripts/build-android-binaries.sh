#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SERVER_DIR="$ROOT_DIR/server-go"
ASSETS_DIR="$ROOT_DIR/android-app/app/src/main/assets/server-binaries"
ANDROID_APP_DIR="$ROOT_DIR/android-app"

resolve_sdk_root() {
  if [[ -n "${ANDROID_SDK_ROOT:-}" ]]; then
    printf '%s' "$ANDROID_SDK_ROOT"
    return
  fi
  if [[ -n "${ANDROID_HOME:-}" ]]; then
    printf '%s' "$ANDROID_HOME"
    return
  fi

  local local_props="$ANDROID_APP_DIR/local.properties"
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

  if [[ -d "$ANDROID_APP_DIR/.android-sdk" ]]; then
    printf '%s' "$ANDROID_APP_DIR/.android-sdk"
    return
  fi

  echo "Android SDK path not found. Configure ANDROID_SDK_ROOT/ANDROID_HOME or android-app/local.properties." >&2
  exit 1
}

resolve_llvm_bin() {
  local sdk_root="$1"
  local ndk_root="$sdk_root/ndk"
  if [[ ! -d "$ndk_root" ]]; then
    echo "Android NDK not found in $ndk_root. Install package 'ndk;27.2.12479018' via sdkmanager." >&2
    exit 1
  fi

  local ndk_version_dir
  ndk_version_dir="$(ls -1 "$ndk_root" | sort -r | head -n1)"
  if [[ -z "$ndk_version_dir" ]]; then
    echo "No NDK versions found in $ndk_root." >&2
    exit 1
  fi

  local host_tag=""
  case "$(uname -s)" in
    Linux*) host_tag="linux-x86_64" ;;
    Darwin*)
      if [[ "$(uname -m)" == "arm64" ]]; then
        host_tag="darwin-arm64"
      else
        host_tag="darwin-x86_64"
      fi
      ;;
    *)
      echo "Unsupported host OS for NDK toolchain: $(uname -s)" >&2
      exit 1
      ;;
  esac

  local llvm_bin="$ndk_root/$ndk_version_dir/toolchains/llvm/prebuilt/$host_tag/bin"
  if [[ ! -d "$llvm_bin" ]]; then
    echo "NDK LLVM toolchain not found: $llvm_bin" >&2
    exit 1
  fi
  printf '%s' "$llvm_bin"
}

build_one() {
  local goarch="$1"
  local abi="$2"
  local clang_target="$3"
  local llvm_bin="$4"
  local goarm="${5:-}"
  local out="$ASSETS_DIR/$abi/server-go"

  mkdir -p "$(dirname "$out")"
  echo "Building $abi..."

  local cc="$llvm_bin/$clang_target"
  if [[ ! -x "$cc" ]]; then
    echo "Clang target wrapper not found for $abi: $cc" >&2
    exit 1
  fi

  if [[ -n "$goarm" ]]; then
    (
      cd "$SERVER_DIR"
      GOOS=android GOARCH="$goarch" GOARM="$goarm" CGO_ENABLED=1 CC="$cc" go build -trimpath -ldflags="-s -w" -o "$out" .
    )
  else
    (
      cd "$SERVER_DIR"
      GOOS=android GOARCH="$goarch" CGO_ENABLED=1 CC="$cc" go build -trimpath -ldflags="-s -w" -o "$out" .
    )
  fi
  chmod +x "$out"
}

SDK_ROOT="$(resolve_sdk_root)"
LLVM_BIN="$(resolve_llvm_bin "$SDK_ROOT")"

build_one "arm64" "arm64-v8a" "aarch64-linux-android26-clang" "$LLVM_BIN"
build_one "arm" "armeabi-v7a" "armv7a-linux-androideabi26-clang" "$LLVM_BIN" "7"
build_one "amd64" "x86_64" "x86_64-linux-android26-clang" "$LLVM_BIN"
build_one "386" "x86" "i686-linux-android26-clang" "$LLVM_BIN"

echo "Done. Binaries copied to $ASSETS_DIR"
