---
name: testing-flutter-ffi-dashboard
description: Test the SlipNet Flutter One-Tap Connect dashboard and dart:ffi native bridge end-to-end on Linux desktop. Use when verifying flutter/ UI, bridge, or native shim changes.
---

# Testing the SlipNet Flutter app + dart:ffi bridge

## Environment setup
- Flutter SDK lives at `~/flutter` (Flutter 3.24.x / Dart 3.5.x). If missing, install it and add
  `~/flutter/bin` to PATH; the repo blueprint may already do this.
- Linux desktop toolchain needed: `clang cmake ninja-build pkg-config libgtk-3-dev`.
- No Android SDK/NDK is available in this env — test via the Linux desktop embedder instead. The
  Dart FFI code path is identical; only the platform embedder differs.

## Fast static checks (run first)
```bash
cd flutter && flutter analyze && flutter test
cd flutter && bash tool/verify_ffi.sh   # compiles the C shim on host + Dart FFI smoke test
```

## Runtime (GUI) test procedure
1. If `flutter/linux/` doesn't exist, generate the runner (project name matters — plain `flutter`
   is rejected):
   ```bash
   cd flutter && flutter create --project-name slipnet_flutter --platforms=linux .
   ```
   Do NOT commit the generated `linux/` runner or `.metadata` changes unless intentionally adding
   Linux support to the repo.
2. Build: `flutter build linux --debug`
3. Compile the shim for the host:
   ```bash
   cc -shared -fPIC -o /tmp/libslipnet_bridge.so flutter/native/slipnet_bridge/slipnet_bridge.c
   ```
4. Launch in a persistent tty shell (one-shot shells launching GUI apps can get killed):
   ```bash
   cd flutter/build/linux/x64/debug/bundle
   SLIPNET_BRIDGE_LIB=/path/to/libslipnet_bridge.so DISPLAY=:0 ./slipnet_flutter
   ```
   `SLIPNET_BRIDGE_LIB` overrides the library load path (see `slipnet_bindings.dart`).
5. Maximize the window: `DISPLAY=:0 wmctrl -r slipnet_flutter -b add,maximized_vert,maximized_horz`

## What proves FFI is live (not a Dart mock)
- Top-right version string `slipnet-bridge/0.1.0 (phase1)` comes from `slipnet_bridge_version()`
  in C; the app crashes at launch if the `.so` fails to load.
- Phase 1 shim emits deterministic values when connected: protocol `slipstream`, ping `42 ms`,
  bytes +1 KB up / +4 KB down per 1s poll. Assert bytes INCREASE across successive polls —
  a static value means the poll loop or FFI stats call is broken.
- Idle expects: `TAP TO CONNECT`, protocol `none`, `0 B`, `0 ms`, `0.0%`, `Idle`.
- Disconnect must reset everything back to idle values.

## Gotchas
- `pkill -f slipnet_flutter` in a one-shot shell may kill the shell itself; use a separate shell.
- Note that the Phase 1 shim only simulates connect/stats; real tunnel connectivity is not
  testable until the shim forwards to the Rust/Go cores (private submodules; may need
  `SUBMODULE_TOKEN`).
- Android APK and iOS builds are out of scope in this env (no SDK/NDK, no macOS).

## Devin Secrets Needed
- `SUBMODULE_TOKEN` (only if building the real Go/Rust cores; not needed for shim-based UI testing).
