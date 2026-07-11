import 'models.dart';
import 'slipnet_bindings.dart';

/// High-level, testable wrapper over the native FFI bindings. UI/state code
/// depends on this abstraction rather than on `dart:ffi` directly, which also
/// lets tests substitute a fake implementation.
abstract class SlipnetBridge {
  int abiVersion();
  String version();
  TunnelState connect(String configJson);
  TunnelState disconnect();
  TunnelState state();
  TunnelStats stats();
  String lastError();
}

/// Default implementation backed by [SlipnetBindings].
class FfiSlipnetBridge implements SlipnetBridge {
  FfiSlipnetBridge(this._bindings);

  factory FfiSlipnetBridge.open() => FfiSlipnetBridge(SlipnetBindings.open());

  final SlipnetBindings _bindings;

  @override
  int abiVersion() => _bindings.abiVersion();

  @override
  String version() => _bindings.version();

  @override
  TunnelState connect(String configJson) {
    _bindings.connect(configJson);
    return state();
  }

  @override
  TunnelState disconnect() {
    _bindings.disconnect();
    return state();
  }

  @override
  TunnelState state() => TunnelState.fromCode(_bindings.state());

  @override
  TunnelStats stats() => TunnelStats.parse(_bindings.statsJson());

  @override
  String lastError() => _bindings.lastError();
}
