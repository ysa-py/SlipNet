import 'package:flutter_test/flutter_test.dart';
import 'package:slipnet_flutter/src/bridge/models.dart';

void main() {
  group('TunnelState.fromCode', () {
    test('maps known codes', () {
      expect(TunnelState.fromCode(0), TunnelState.disconnected);
      expect(TunnelState.fromCode(1), TunnelState.connecting);
      expect(TunnelState.fromCode(2), TunnelState.connected);
      expect(TunnelState.fromCode(3), TunnelState.error);
    });

    test('maps unknown code to disconnected', () {
      expect(TunnelState.fromCode(99), TunnelState.disconnected);
    });

    test('isActive reflects connecting/connected', () {
      expect(TunnelState.connecting.isActive, isTrue);
      expect(TunnelState.connected.isActive, isTrue);
      expect(TunnelState.disconnected.isActive, isFalse);
      expect(TunnelState.error.isActive, isFalse);
    });
  });

  group('TunnelStats.parse', () {
    test('parses a well-formed payload', () {
      final stats = TunnelStats.parse(
        '{"bytesUp":10,"bytesDown":20,"rttMs":30,"packetLossPct":1.5,"protocol":"slipstream"}',
      );
      expect(stats.bytesUp, 10);
      expect(stats.bytesDown, 20);
      expect(stats.rttMs, 30);
      expect(stats.packetLossPct, 1.5);
      expect(stats.protocol, 'slipstream');
    });

    test('returns empty on malformed json', () {
      expect(TunnelStats.parse('not json').bytesUp, 0);
      expect(TunnelStats.parse('[1,2,3]').protocol, 'none');
    });

    test('tolerates missing fields', () {
      final stats = TunnelStats.parse('{"bytesUp":5}');
      expect(stats.bytesUp, 5);
      expect(stats.bytesDown, 0);
      expect(stats.protocol, 'none');
    });
  });
}
