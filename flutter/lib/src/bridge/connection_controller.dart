import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'models.dart';
import 'slipnet_bridge.dart';

/// Provides the native bridge. Overridden in tests with a fake.
final slipnetBridgeProvider = Provider<SlipnetBridge>((ref) {
  return FfiSlipnetBridge.open();
});

/// Aggregate UI-facing connection state.
class ConnectionSnapshot {
  const ConnectionSnapshot({
    required this.state,
    required this.stats,
    this.error,
  });

  final TunnelState state;
  final TunnelStats stats;
  final String? error;

  static const ConnectionSnapshot initial = ConnectionSnapshot(
    state: TunnelState.disconnected,
    stats: TunnelStats.empty,
  );

  ConnectionSnapshot copyWith({
    TunnelState? state,
    TunnelStats? stats,
    String? error,
  }) {
    return ConnectionSnapshot(
      state: state ?? this.state,
      stats: stats ?? this.stats,
      error: error,
    );
  }
}

/// Drives connect/disconnect and polls the bridge for live stats.
class ConnectionController extends StateNotifier<ConnectionSnapshot> {
  ConnectionController(this._bridge) : super(ConnectionSnapshot.initial);

  final SlipnetBridge _bridge;
  Timer? _pollTimer;

  /// Placeholder profile config. Later phases pass the selected profile JSON.
  static const String _demoConfig =
      '{"tunnel":"slipstream","domain":"example.com"}';

  String bridgeVersion() => _bridge.version();

  void toggle() {
    if (state.state.isActive) {
      disconnect();
    } else {
      connect();
    }
  }

  void connect([String? configJson]) {
    final result = _bridge.connect(configJson ?? _demoConfig);
    state = state.copyWith(
      state: result,
      error: result == TunnelState.error ? _bridge.lastError() : null,
    );
    if (result == TunnelState.connected) {
      _startPolling();
    }
  }

  void disconnect() {
    _stopPolling();
    final result = _bridge.disconnect();
    state = ConnectionSnapshot(state: result, stats: TunnelStats.empty);
  }

  void _startPolling() {
    _pollTimer?.cancel();
    _pollTimer = Timer.periodic(const Duration(seconds: 1), (_) => _poll());
    _poll();
  }

  void _stopPolling() {
    _pollTimer?.cancel();
    _pollTimer = null;
  }

  void _poll() {
    state = state.copyWith(
      state: _bridge.state(),
      stats: _bridge.stats(),
    );
  }

  @override
  void dispose() {
    _stopPolling();
    super.dispose();
  }
}

final connectionControllerProvider =
    StateNotifierProvider<ConnectionController, ConnectionSnapshot>((ref) {
  return ConnectionController(ref.watch(slipnetBridgeProvider));
});
