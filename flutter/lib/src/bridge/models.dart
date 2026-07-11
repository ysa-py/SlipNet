import 'dart:convert';

/// Mirrors the SLIPNET_STATE_* enum in `slipnet_bridge.h`.
enum TunnelState {
  disconnected,
  connecting,
  connected,
  error;

  static TunnelState fromCode(int code) {
    switch (code) {
      case 1:
        return TunnelState.connecting;
      case 2:
        return TunnelState.connected;
      case 3:
        return TunnelState.error;
      case 0:
      default:
        return TunnelState.disconnected;
    }
  }

  bool get isActive =>
      this == TunnelState.connecting || this == TunnelState.connected;
}

/// Live tunnel statistics decoded from `slipnet_bridge_stats_json`.
class TunnelStats {
  const TunnelStats({
    this.bytesUp = 0,
    this.bytesDown = 0,
    this.rttMs = 0,
    this.packetLossPct = 0,
    this.protocol = 'none',
  });

  final int bytesUp;
  final int bytesDown;
  final int rttMs;
  final double packetLossPct;
  final String protocol;

  static const TunnelStats empty = TunnelStats();

  factory TunnelStats.fromJson(Map<String, dynamic> json) {
    return TunnelStats(
      bytesUp: (json['bytesUp'] as num?)?.toInt() ?? 0,
      bytesDown: (json['bytesDown'] as num?)?.toInt() ?? 0,
      rttMs: (json['rttMs'] as num?)?.toInt() ?? 0,
      packetLossPct: (json['packetLossPct'] as num?)?.toDouble() ?? 0,
      protocol: json['protocol'] as String? ?? 'none',
    );
  }

  /// Parses a stats JSON string, tolerating malformed input.
  static TunnelStats parse(String raw) {
    try {
      final decoded = jsonDecode(raw);
      if (decoded is Map<String, dynamic>) {
        return TunnelStats.fromJson(decoded);
      }
    } on FormatException {
      // fall through to empty
    }
    return TunnelStats.empty;
  }
}
