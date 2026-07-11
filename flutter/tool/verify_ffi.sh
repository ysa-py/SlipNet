#!/usr/bin/env bash
# Builds the slipnet_bridge shim for the host and runs the dart:ffi smoke test.
# This verifies the full Dart <-> native FFI pipeline without an Android device.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FLUTTER_DIR="$(dirname "$SCRIPT_DIR")"
NATIVE_DIR="$FLUTTER_DIR/native/slipnet_bridge"
BUILD_DIR="$(mktemp -d)"

case "$(uname -s)" in
  Darwin) LIB_NAME="libslipnet_bridge.dylib" ;;
  MINGW*|MSYS*|CYGWIN*) LIB_NAME="slipnet_bridge.dll" ;;
  *) LIB_NAME="libslipnet_bridge.so" ;;
esac

echo "==> Building shim ($LIB_NAME)"
cc -shared -fPIC -O2 \
  -I "$NATIVE_DIR" \
  "$NATIVE_DIR/slipnet_bridge.c" \
  -o "$BUILD_DIR/$LIB_NAME"

echo "==> Running dart:ffi smoke test"
cd "$FLUTTER_DIR"
SLIPNET_BRIDGE_LIB="$BUILD_DIR/$LIB_NAME" dart run tool/ffi_smoke.dart
