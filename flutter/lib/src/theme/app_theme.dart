import 'package:flutter/material.dart';

/// Brand palette + a dark, glassmorphism-friendly theme shared across
/// Android and iOS (Material 3 base with translucent surfaces).
class AppTheme {
  static const Color accent = Color(0xFF5B8CFF);
  static const Color accentAlt = Color(0xFF00E0C6);
  static const Color danger = Color(0xFFFF5C7A);
  static const Color bgTop = Color(0xFF0B1020);
  static const Color bgBottom = Color(0xFF141B34);

  static ThemeData dark() {
    final base = ThemeData(
      useMaterial3: true,
      brightness: Brightness.dark,
      colorScheme: const ColorScheme.dark(
        primary: accent,
        secondary: accentAlt,
        error: danger,
        surface: bgBottom,
      ),
      scaffoldBackgroundColor: bgTop,
    );
    return base.copyWith(
      textTheme: base.textTheme.apply(
        bodyColor: Colors.white,
        displayColor: Colors.white,
      ),
    );
  }

  static const LinearGradient backgroundGradient = LinearGradient(
    begin: Alignment.topLeft,
    end: Alignment.bottomRight,
    colors: [bgTop, bgBottom],
  );
}
