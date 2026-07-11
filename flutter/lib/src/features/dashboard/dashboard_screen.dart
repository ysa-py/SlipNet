import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../bridge/connection_controller.dart';
import '../../bridge/models.dart';
import '../../theme/app_theme.dart';
import 'widgets/connect_button.dart';
import 'widgets/glass_card.dart';
import 'widgets/stat_tile.dart';

/// One-Tap Connect dashboard. Binds to [connectionControllerProvider], which is
/// driven by the native `dart:ffi` bridge.
class DashboardScreen extends ConsumerWidget {
  const DashboardScreen({super.key});

  static String _formatBytes(int bytes) {
    if (bytes < 1024) return '$bytes B';
    if (bytes < 1024 * 1024) return '${(bytes / 1024).toStringAsFixed(1)} KB';
    return '${(bytes / (1024 * 1024)).toStringAsFixed(1)} MB';
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final snapshot = ref.watch(connectionControllerProvider);
    final controller = ref.read(connectionControllerProvider.notifier);
    final stats = snapshot.stats;

    return Scaffold(
      body: Container(
        decoration: const BoxDecoration(gradient: AppTheme.backgroundGradient),
        child: SafeArea(
          child: Padding(
            padding: const EdgeInsets.all(20),
            child: Column(
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    const Text(
                      'SlipNet',
                      style: TextStyle(
                        fontSize: 24,
                        fontWeight: FontWeight.w700,
                        letterSpacing: 1.2,
                      ),
                    ),
                    Text(
                      controller.bridgeVersion(),
                      style: const TextStyle(fontSize: 11, color: Colors.white38),
                    ),
                  ],
                ),
                const Spacer(),
                ConnectButton(
                  state: snapshot.state,
                  onTap: controller.toggle,
                ),
                if (snapshot.error != null) ...[
                  const SizedBox(height: 12),
                  Text(
                    snapshot.error!,
                    style: const TextStyle(color: AppTheme.danger),
                  ),
                ],
                const Spacer(),
                GlassCard(
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      _kv('Protocol', stats.protocol),
                      _kv('Up', _formatBytes(stats.bytesUp)),
                      _kv('Down', _formatBytes(stats.bytesDown)),
                    ],
                  ),
                ),
                const SizedBox(height: 12),
                Row(
                  children: [
                    Expanded(
                      child: StatTile(
                        label: 'Ping',
                        value: '${stats.rttMs} ms',
                        icon: Icons.speed_rounded,
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: StatTile(
                        label: 'Packet loss',
                        value: '${stats.packetLossPct.toStringAsFixed(1)}%',
                        icon: Icons.stacked_line_chart_rounded,
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: StatTile(
                        label: 'Status',
                        value: _statusText(snapshot.state),
                        icon: Icons.wifi_tethering_rounded,
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  static String _statusText(TunnelState state) {
    switch (state) {
      case TunnelState.connected:
        return 'Active';
      case TunnelState.connecting:
        return 'Linking';
      case TunnelState.error:
        return 'Error';
      case TunnelState.disconnected:
        return 'Idle';
    }
  }

  Widget _kv(String k, String v) {
    return Column(
      children: [
        Text(v, style: const TextStyle(fontWeight: FontWeight.bold)),
        const SizedBox(height: 4),
        Text(k, style: const TextStyle(fontSize: 11, color: Colors.white54)),
      ],
    );
  }
}
