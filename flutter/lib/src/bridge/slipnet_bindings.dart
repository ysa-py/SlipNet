import 'dart:ffi';
import 'dart:io';

import 'package:ffi/ffi.dart';

/// Low-level `dart:ffi` bindings to `libslipnet_bridge` (the C-ABI shim in
/// `native/slipnet_bridge/`). This is the single place that touches raw FFI;
/// higher layers use [SlipnetBridge].
typedef _AbiVersionC = Int32 Function();
typedef _AbiVersionDart = int Function();

typedef _VersionC = Pointer<Utf8> Function();
typedef _VersionDart = Pointer<Utf8> Function();

typedef _ConnectC = Int32 Function(Pointer<Utf8>);
typedef _ConnectDart = int Function(Pointer<Utf8>);

typedef _VoidReturnsInt32C = Int32 Function();
typedef _VoidReturnsInt32Dart = int Function();

typedef _StatsC = Pointer<Utf8> Function();
typedef _StatsDart = Pointer<Utf8> Function();

/// Thrown when the native bridge library cannot be located/loaded.
class BridgeLoadException implements Exception {
  BridgeLoadException(this.message);
  final String message;
  @override
  String toString() => 'BridgeLoadException: $message';
}

/// Resolves and opens the platform-specific shared library.
DynamicLibrary openBridgeLibrary() {
  // Allow tests / desktop to point at a locally-built shim.
  final override = Platform.environment['SLIPNET_BRIDGE_LIB'];
  if (override != null && override.isNotEmpty) {
    return DynamicLibrary.open(override);
  }
  if (Platform.isAndroid) {
    return DynamicLibrary.open('libslipnet_bridge.so');
  }
  if (Platform.isIOS || Platform.isMacOS) {
    // Statically linked into the runner on Apple platforms.
    return DynamicLibrary.process();
  }
  if (Platform.isLinux) {
    return DynamicLibrary.open('libslipnet_bridge.so');
  }
  if (Platform.isWindows) {
    return DynamicLibrary.open('slipnet_bridge.dll');
  }
  throw BridgeLoadException('unsupported platform: ${Platform.operatingSystem}');
}

/// Typed function pointers into the bridge shim.
class SlipnetBindings {
  SlipnetBindings(DynamicLibrary lib)
      : _abiVersion =
            lib.lookupFunction<_AbiVersionC, _AbiVersionDart>('slipnet_bridge_abi_version'),
        _version =
            lib.lookupFunction<_VersionC, _VersionDart>('slipnet_bridge_version'),
        _connect =
            lib.lookupFunction<_ConnectC, _ConnectDart>('slipnet_bridge_connect'),
        _disconnect = lib.lookupFunction<_VoidReturnsInt32C, _VoidReturnsInt32Dart>(
            'slipnet_bridge_disconnect'),
        _state = lib.lookupFunction<_VoidReturnsInt32C, _VoidReturnsInt32Dart>(
            'slipnet_bridge_state'),
        _statsJson = lib.lookupFunction<_StatsC, _StatsDart>('slipnet_bridge_stats_json'),
        _lastError = lib.lookupFunction<_StatsC, _StatsDart>('slipnet_bridge_last_error');

  factory SlipnetBindings.open() => SlipnetBindings(openBridgeLibrary());

  final _AbiVersionDart _abiVersion;
  final _VersionDart _version;
  final _ConnectDart _connect;
  final _VoidReturnsInt32Dart _disconnect;
  final _VoidReturnsInt32Dart _state;
  final _StatsDart _statsJson;
  final _StatsDart _lastError;

  int abiVersion() => _abiVersion();

  String version() => _version().toDartString();

  int connect(String configJson) {
    final ptr = configJson.toNativeUtf8();
    try {
      return _connect(ptr);
    } finally {
      malloc.free(ptr);
    }
  }

  int disconnect() => _disconnect();

  int state() => _state();

  String statsJson() => _statsJson().toDartString();

  String lastError() => _lastError().toDartString();
}
