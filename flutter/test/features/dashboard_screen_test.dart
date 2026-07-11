import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:slipnet_flutter/src/bridge/connection_controller.dart';
import 'package:slipnet_flutter/src/features/dashboard/dashboard_screen.dart';
import 'package:slipnet_flutter/src/features/dashboard/widgets/connect_button.dart';

import '../support/fake_bridge.dart';

Widget _wrap() {
  return ProviderScope(
    overrides: [
      slipnetBridgeProvider.overrideWithValue(FakeSlipnetBridge()),
    ],
    child: const MaterialApp(home: DashboardScreen()),
  );
}

void main() {
  testWidgets('renders One-Tap Connect in idle state', (tester) async {
    await tester.pumpWidget(_wrap());

    expect(find.text('SlipNet'), findsOneWidget);
    expect(find.text('TAP TO CONNECT'), findsOneWidget);
    expect(find.byType(ConnectButton), findsOneWidget);
  });

  testWidgets('tapping the connect button transitions to connected', (tester) async {
    await tester.pumpWidget(_wrap());

    await tester.tap(find.byType(ConnectButton));
    await tester.pump(); // process state change
    await tester.pump(const Duration(milliseconds: 50));

    expect(find.text('CONNECTED'), findsOneWidget);

    // stop the polling timer before the test ends
    await tester.tap(find.byType(ConnectButton));
    await tester.pump();
  });
}
