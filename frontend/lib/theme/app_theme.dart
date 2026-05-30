import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

class AppTheme {
  static ThemeMode currentMode = ThemeMode.dark;

  static const Color _darkPrimary = Color(0xFF1A1A2E);
  static const Color _darkSecondary = Color(0xFF16213E);
  static const Color _darkCardBg = Color(0xFF0F3460);
  static const Color _darkSurface = Color(0xFF1E2A45);
  static const Color _darkTextPrimary = Color(0xFFF0E6D3);
  static const Color _darkTextSecondary = Color(0xFFB8A99A);

  static const Color _lightPrimary = Color(0xFFF8F5EF);
  static const Color _lightSecondary = Color(0xFFECE2D3);
  static const Color _lightCardBg = Color(0xFFD9C6AA);
  static const Color _lightSurface = Color(0xFFFFFFFF);
  static const Color _lightTextPrimary = Color(0xFF1F2933);
  static const Color _lightTextSecondary = Color(0xFF667085);

  static const Color accent = Color(0xFFE8A838);
  static const Color accentLight = Color(0xFFF5C842);
  static const Color success = Color(0xFF4CAF50);
  static const Color error = Color(0xFFE53935);
  static const Color correct = Color(0xFF2ECC71);
  static const Color wrong = Color(0xFFE74C3C);

  static bool get isLight => currentMode == ThemeMode.light;
  static Color get primary => isLight ? _lightPrimary : _darkPrimary;
  static Color get secondary => isLight ? _lightSecondary : _darkSecondary;
  static Color get cardBg => isLight ? _lightCardBg : _darkCardBg;
  static Color get surface => isLight ? _lightSurface : _darkSurface;
  static Color get textPrimary => isLight ? _lightTextPrimary : _darkTextPrimary;
  static Color get textSecondary => isLight ? _lightTextSecondary : _darkTextSecondary;
  static Color get onAccent => const Color(0xFF1F2933);
  static Color get elevatedSurface => isLight ? const Color(0xFFFFFBF5) : _darkSurface;
  static Color get softShadow => isLight ? const Color(0xFFD8C7AE).withOpacity(0.28) : Colors.black.withOpacity(0.35);
  static Color get overlayOnImage => isLight ? Colors.white.withOpacity(0.78) : Colors.black.withOpacity(0.38);

  static List<Color> get dailyGoalGradient => isLight
      ? const [Color(0xFFFFFBF5), Color(0xFFF0E4D1)]
      : const [Color(0xFF1A3A5C), Color(0xFF0F3460)];

  static List<Color> get featuredLessonGradient => isLight
      ? const [Color(0xFFFFF3DD), Color(0xFFE8C58C)]
      : const [Color(0xFF8B4513), Color(0xFFD2691E)];

  static List<Color> get lessonHeroGradient => isLight
      ? const [Color(0xFFFFF0D2), Color(0xFFE2B978), Color(0xFFF8F5EF)]
      : const [Color(0xFF8B4513), Color(0xFF2C1810), Color(0xFF1A1A2E)];

  static ThemeData get theme => darkTheme;

  static ThemeData get darkTheme {
    return ThemeData(
      useMaterial3: true,
      scaffoldBackgroundColor: _darkPrimary,
      colorScheme: const ColorScheme.dark(
        primary: accent,
        secondary: accentLight,
        surface: _darkSurface,
        background: _darkPrimary,
        onPrimary: Color(0xFF1F2933),
        onSurface: _darkTextPrimary,
      ),
      textTheme: GoogleFonts.playfairDisplayTextTheme().copyWith(
        displayLarge: GoogleFonts.playfairDisplay(
          color: _darkTextPrimary,
          fontSize: 32,
          fontWeight: FontWeight.bold,
        ),
        displayMedium: GoogleFonts.playfairDisplay(
          color: _darkTextPrimary,
          fontSize: 24,
          fontWeight: FontWeight.bold,
        ),
        headlineMedium: GoogleFonts.playfairDisplay(
          color: _darkTextPrimary,
          fontSize: 20,
          fontWeight: FontWeight.w600,
        ),
        bodyLarge: GoogleFonts.lato(
          color: _darkTextPrimary,
          fontSize: 16,
        ),
        bodyMedium: GoogleFonts.lato(
          color: _darkTextSecondary,
          fontSize: 14,
        ),
        labelLarge: GoogleFonts.lato(
          color: _darkPrimary,
          fontSize: 16,
          fontWeight: FontWeight.bold,
        ),
      ),
      appBarTheme: AppBarTheme(
        backgroundColor: _darkPrimary,
        elevation: 0,
        titleTextStyle: GoogleFonts.playfairDisplay(
          color: _darkTextPrimary,
          fontSize: 20,
          fontWeight: FontWeight.bold,
        ),
        iconTheme: const IconThemeData(color: accent),
      ),
      bottomNavigationBarTheme: const BottomNavigationBarThemeData(
        backgroundColor: _darkSecondary,
        selectedItemColor: accent,
        unselectedItemColor: _darkTextSecondary,
        type: BottomNavigationBarType.fixed,
        elevation: 20,
      ),
    );
  }

  static ThemeData get lightTheme {
    const lightPrimary = Color(0xFFF8F5EF);
    const lightSurface = Color(0xFFFFFFFF);
    const lightSecondary = Color(0xFFECE2D3);
    const lightCard = Color(0xFFD9C6AA);
    const lightText = Color(0xFF1F2933);
    const lightMuted = Color(0xFF667085);
    return ThemeData(
      useMaterial3: true,
      scaffoldBackgroundColor: lightPrimary,
      colorScheme: const ColorScheme.light(
        primary: accent,
        secondary: accentLight,
        surface: lightSurface,
        background: lightPrimary,
        onPrimary: Color(0xFF1F2933),
        onSurface: lightText,
      ),
      textTheme: GoogleFonts.playfairDisplayTextTheme().copyWith(
        displayLarge: GoogleFonts.playfairDisplay(color: lightText, fontSize: 32, fontWeight: FontWeight.bold),
        displayMedium: GoogleFonts.playfairDisplay(color: lightText, fontSize: 24, fontWeight: FontWeight.bold),
        headlineMedium: GoogleFonts.playfairDisplay(color: lightText, fontSize: 20, fontWeight: FontWeight.w600),
        bodyLarge: GoogleFonts.lato(color: lightText, fontSize: 16),
        bodyMedium: GoogleFonts.lato(color: lightMuted, fontSize: 14),
        labelLarge: GoogleFonts.lato(color: lightText, fontSize: 16, fontWeight: FontWeight.bold),
      ),
      appBarTheme: AppBarTheme(
        backgroundColor: lightPrimary,
        elevation: 0,
        titleTextStyle: GoogleFonts.playfairDisplay(color: lightText, fontSize: 20, fontWeight: FontWeight.bold),
        iconTheme: const IconThemeData(color: accent),
      ),
      bottomNavigationBarTheme: const BottomNavigationBarThemeData(
        backgroundColor: lightSecondary,
        selectedItemColor: accent,
        unselectedItemColor: lightMuted,
        type: BottomNavigationBarType.fixed,
        elevation: 20,
      ),
      cardTheme: const CardThemeData(color: lightSurface),
      inputDecorationTheme: OutlineInputBorder(
        borderSide: const BorderSide(color: lightCard),
        borderRadius: BorderRadius.circular(8),
      ).toInputDecorationTheme(),
    );
  }
}

extension on OutlineInputBorder {
  InputDecorationTheme toInputDecorationTheme() {
    return InputDecorationTheme(
      enabledBorder: this,
      focusedBorder: copyWith(borderSide: const BorderSide(color: AppTheme.accent)),
    );
  }
}
