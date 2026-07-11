# SlipNet Flutter Re-architecture — Architecture & Phased Plan

Status: **Phase 1 (scaffold + FFI bridge) in progress.** This document is the living
architecture reference for migrating the SlipNet client UI from Kotlin/Jetpack Compose
to a cross-platform Flutter application, while keeping the existing Go/Rust/C network
cores untouched.

---

## 1. Goals & non-goals

**Goals**
- Replace the Android UI (Kotlin/Compose, `app/src/main/java/app/slipnet/presentation/**`)
  with a single Flutter codebase targeting Android and iOS.
- Bind the Flutter UI to the existing native network cores via `dart:ffi` through a thin,
  stable C-ABI bridge shim — **without modifying or re-implementing the cores**.
- Provide a "One-Tap Connect" dashboard with real-time analytics (ping, packet loss,
  active protocol, node-switching status).

**Non-goals (explicitly out of scope / preserved as-is)**
- Rewriting or translating the Rust (`slipstream-rust`), Go (`noizdns`, `vaydns-mobile`,
  `dnstt`), or C (`hev-socks5-tunnel`) cores. These stay byte-for-byte.
- Removing the Android-native VPN plumbing that legally must live in Android
  (`VpnService`, TUN fd ownership, socket protection). Flutter cannot own a `VpnService`;
  see §4.

---

## 2. Current architecture (as found)

```
Kotlin/Compose UI  (presentation/**, MVVM + Hilt)
        │
Kotlin domain/data (usecases, repositories, Room, DataStore)
        │
Kotlin "middleware" (service/**, tunnel/**)  ← VpnService, SOCKS bridges, orchestration
        │  JNI (System.loadLibrary)
        ▼
Native cores:
  • libslipstream.so     — Rust QUIC/DNS core (JNI: nativeStartTunnel, nativeGetTrafficStats, …)
  • golibs*.aar          — Go (gomobile bind: noizdns/mobile, vaydns-mobile, snowflake)
  • libhev-socks5-tunnel — C tun2socks (JNI: nativeStart/nativeStop)
  • libslipnet-sockops   — C socket options helper
```

Key native entry points discovered (JNI, in `data/native/NativeBridge.kt`,
`tunnel/HevSocks5Tunnel.kt`, `tunnel/SlipstreamBridge.kt`, `tunnel/NativeSocket.kt`):
- `nativeStartTunnel(domain, resolverHosts[], resolverPorts[], …, tunFd, dnsServer) : Int`
- `nativeStopTunnel()`, `nativeIsConnected() : Boolean`, `nativeGetTrafficStats() : long[]`,
  `nativeGetVersion() : String`
- Callbacks JNI→Kotlin: `updateStats(...)`, `updateState(int)`, `reportError(code, msg)`

> The core submodules are **private and not checked out** in this environment
> (`git submodule status` shows them unfetched; cloning returns HTTP 403 without
> `SUBMODULE_TOKEN`). Phase 1 therefore ships a bridge shim with a stable C ABI that
> compiles and runs today, with core-forwarding stubbed behind clearly-marked TODOs.

---

## 3. Target architecture

```
Flutter UI (lib/features/**)  ← Glassmorphism, One-Tap Connect, live charts
        │  Riverpod state
Flutter services (lib/src/bridge/**)  ← ConnectionController, StatsStream
        │  dart:ffi
        ▼
libslipnet_bridge.so / .a  ── thin C ABI shim (this repo, versioned)
        │  (C / cgo / extern "C")
        ▼
Existing native cores (UNTOUCHED): libslipstream, golibs, hev-socks5-tunnel
```

The bridge shim (`flutter/native/slipnet_bridge/`) is the **only** new native code. It
exposes a minimal, versioned C ABI (`slipnet_bridge.h`) that the Dart FFI layer binds to.
Internally it will (in later phases) call the existing cores' exported symbols; today it
returns real values from the shim itself so the whole pipeline is verifiable end-to-end.

### 3.1 FFI contract (Phase 1)

| C symbol | Signature | Purpose |
|---|---|---|
| `slipnet_bridge_abi_version` | `int32_t(void)` | ABI compatibility guard |
| `slipnet_bridge_version` | `const char*(void)` | human-readable build/version string |
| `slipnet_bridge_connect` | `int32_t(const char* config_json)` | begin connect; returns status code |
| `slipnet_bridge_disconnect` | `int32_t(void)` | tear down |
| `slipnet_bridge_state` | `int32_t(void)` | 0=disconnected 1=connecting 2=connected 3=error |
| `slipnet_bridge_stats_json` | `const char*(void)` | JSON: bytesUp/Down, rttMs, packetLoss, protocol |
| `slipnet_bridge_last_error` | `const char*(void)` | last error message |

Strings returned are owned by the shim (thread-local static buffers); Dart copies them and
must not free them. This keeps the FFI boundary allocation-safe and simple.

---

## 4. Android VPN reality (important constraint)

Flutter/Dart cannot host an Android `VpnService` or own the TUN file descriptor — that is a
platform service that must be a Kotlin/Java `Service` in the Android host. The migration
therefore keeps a **thin** Android platform layer:

- A minimal `VpnService` (reused/trimmed from the current one) obtains the TUN fd and passes
  it down to the cores via the bridge, and handles `protect(fd)`.
- Everything above the fd — UI, profile management, orchestration decisions, analytics — moves
  to Flutter/Dart.

This is standard for Flutter VPN apps and does not violate "don't touch the cores": the cores
are unchanged; only the Kotlin UI/orchestration is replaced.

---

## 5. Phased delivery plan

- **Phase 1 (this PR): scaffold + working FFI bridge.**
  Flutter app skeleton (Android+iOS), the `slipnet_bridge` C-ABI shim wired into the Android
  Gradle/CMake build, Dart FFI bindings, a One-Tap Connect dashboard driven by the bridge
  (version + simulated state/stats), unit + widget tests, and host-side FFI verification.
- **Phase 2:** real core forwarding for the Rust `slipstream` path (map `slipnet_bridge_connect`
  → the core's exported start/stop), plus the thin Kotlin `VpnService`/TUN handshake and
  method-channel-free fd hand-off. Requires `SUBMODULE_TOKEN` to fetch cores.
- **Phase 3:** multi-core orchestration & fallback (DNSTT / VayDNS / NoizDNS / VLESS / SSH),
  latency-based switching, node tagging ("VIP") for load balancing.
- **Phase 4:** SQLite persistence (`./db/custom.db`, relative path), profile import/export,
  Developer Mode logs, real-time charts wired to `stats_json`.
- **Phase 5:** iOS packaging (**requires macOS + Xcode**: build cores as `.framework`/`.a`,
  NetworkExtension `PacketTunnelProvider`). Cannot be produced on this Linux environment.

### External dependencies requiring decisions (flagged, not yet added)
- "GFW-knocker" Psiphon fork — needs an exact repo URL/commit before integration.
- Xray / sing-box / AmneziaVPN fallback engines — need to confirm licensing and whether they
  ship as Go c-shared libs consumable through the same bridge.

---

## 6. Repository layout (new)

```
flutter/
  lib/
    main.dart
    src/
      bridge/            # dart:ffi bindings + high-level ConnectionController
      features/
        dashboard/       # One-Tap Connect screen + widgets
      theme/             # glassmorphism theme
  native/
    slipnet_bridge/      # C ABI shim (CMake), the only new native code
  test/                  # unit + widget tests
  tool/                  # host-side FFI verification script
docs/flutter-rearchitecture.md   # this file
```
