import 'package:flutter/material.dart';

import '../../../bridge/models.dart';
import '../../../theme/app_theme.dart';

/// Hyper-minimalist "One-Tap Connect" control. A single large circular button
/// whose colour/label reflect the current [TunnelState].
class ConnectButton extends StatelessWidget {
  const ConnectButton({
    super.key,
    required this.state,
    required this.onTap,
  });

  final TunnelState state;
  final VoidCallback onTap;

  Color get _color {
    switch (state) {
      case TunnelState.connected:
        return AppTheme.accentAlt;
      case TunnelState.connecting:
        return AppTheme.accent;
      case TunnelState.error:
        return AppTheme.danger;
      case TunnelState.disconnected:
        return AppTheme.accent;
    }
  }

  String get _label {
    switch (state) {
      case TunnelState.connected:
        return 'CONNECTED';
      case TunnelState.connecting:
        return 'CONNECTING';
      case TunnelState.error:
        return 'RETRY';
      case TunnelState.disconnected:
        return 'TAP TO CONNECT';
    }
  }

  @override
  Widget build(BuildContext context) {
    return Semantics(
      button: true,
      label: _label,
      child: GestureDetector(
        onTap: onTap,
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 350),
          curve: Curves.easeOutCubic,
          width: 220,
          height: 220,
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            gradient: RadialGradient(
              colors: [_color.withOpacity(0.35), Colors.transparent],
              radius: 0.9,
            ),
            border: Border.all(color: _color, width: 3),
            boxShadow: [
              BoxShadow(
                color: _color.withOpacity(0.4),
                blurRadius: 40,
                spreadRadius: 4,
              ),
            ],
          ),
          child: Center(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Icon(
                  state == TunnelState.connected
                      ? Icons.shield_rounded
                      : Icons.power_settings_new_rounded,
                  size: 64,
                  color: _color,
                ),
                const SizedBox(height: 12),
                Text(
                  _label,
                  style: TextStyle(
                    color: _color,
                    fontWeight: FontWeight.bold,
                    letterSpacing: 1.5,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
