import 'package:flutter_test/flutter_test.dart';
import 'package:slipnet_flutter/src/bridge/connection_controller.dart';
import 'package:slipnet_flutter/src/bridge/models.dart';

import '../support/fake_bridge.dart';

void main() {
  test('connect transitions to connected and polls stats', () async {
    final controller = ConnectionController(FakeSlipnetBridge());
    addTearDown(controller.dispose);

    controller.connect();
    expect(controller.state.state, TunnelState.connected);

    // Polling should populate stats within a couple of ticks.
    await Future<void>.delayed(const Duration(milliseconds: 1100));
    expect(controller.state.stats.bytesDown, greaterThan(0));
    expect(controller.state.stats.protocol, 'slipstream');
  });

  test('toggle connects then disconnects', () {
    final controller = ConnectionController(FakeSlipnetBridge());
    addTearDown(controller.dispose);

    controller.toggle();
    expect(controller.state.state, TunnelState.connected);

    controller.toggle();
    expect(controller.state.state, TunnelState.disconnected);
    expect(controller.state.stats.bytesUp, 0);
  });

  test('failed connect surfaces error message', () {
    final controller = ConnectionController(FakeSlipnetBridge(failConnect: true));
    addTearDown(controller.dispose);

    controller.connect();
    expect(controller.state.state, TunnelState.error);
    expect(controller.state.error, isNotNull);
  });
}
