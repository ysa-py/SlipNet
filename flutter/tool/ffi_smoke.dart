// Host-side smoke test for the dart:ffi bridge. Requires the shim to be built
// and SLIPNET_BRIDGE_LIB to point at it (see tool/verify_ffi.sh).
//
//   dart run tool/ffi_smoke.dart
//
// Exits non-zero on any failed assertion.
import 'dart:io';

import 'package:slipnet_flutter/src/bridge/models.dart';
import 'package:slipnet_flutter/src/bridge/slipnet_bridge.dart';

void _check(bool ok, String what) {
  stdout.writeln('${ok ? 'PASS' : 'FAIL'}: $what');
  if (!ok) exitCode = 1;
}

void main() {
  final bridge = FfiSlipnetBridge.open();

  _check(bridge.abiVersion() == 1, 'abiVersion == 1');
  _check(bridge.version().contains('slipnet-bridge'), 'version string');
  _check(bridge.state() == TunnelState.disconnected, 'initial state disconnected');

  final connected = bridge.connect('{"tunnel":"slipstream"}');
  _check(connected == TunnelState.connected, 'connect -> connected');

  final stats = bridge.stats();
  _check(stats.protocol == 'slipstream', 'stats protocol == slipstream');
  _check(stats.rttMs == 42, 'stats rttMs == 42');

  final disconnected = bridge.disconnect();
  _check(disconnected == TunnelState.disconnected, 'disconnect -> disconnected');

  final bad = bridge.connect('');
  _check(bad == TunnelState.error, 'empty config -> error');
  _check(bridge.lastError().isNotEmpty, 'lastError populated on failure');

  stdout.writeln(exitCode == 0 ? 'ALL FFI CHECKS PASSED' : 'FFI CHECKS FAILED');
}
