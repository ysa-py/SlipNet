import 'package:slipnet_flutter/src/bridge/models.dart';
import 'package:slipnet_flutter/src/bridge/slipnet_bridge.dart';

/// In-memory [SlipnetBridge] used by tests so no native `.so` is required.
class FakeSlipnetBridge implements SlipnetBridge {
  FakeSlipnetBridge({this.failConnect = false});

  final bool failConnect;
  TunnelState _state = TunnelState.disconnected;
  int _tick = 0;

  @override
  int abiVersion() => 1;

  @override
  String version() => 'fake-bridge/test';

  @override
  TunnelState connect(String configJson) {
    if (failConnect || configJson.isEmpty) {
      _state = TunnelState.error;
    } else {
      _state = TunnelState.connected;
    }
    return _state;
  }

  @override
  TunnelState disconnect() {
    _state = TunnelState.disconnected;
    _tick = 0;
    return _state;
  }

  @override
  TunnelState state() => _state;

  @override
  TunnelStats stats() {
    if (_state != TunnelState.connected) return TunnelStats.empty;
    _tick++;
    return TunnelStats(
      bytesUp: 1024 * _tick,
      bytesDown: 4096 * _tick,
      rttMs: 42,
      protocol: 'slipstream',
    );
  }

  @override
  String lastError() =>
      _state == TunnelState.error ? 'fake connect failure' : '';
}
