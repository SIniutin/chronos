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
  static Color get textPrimary =>
      isLight ? _lightTextPrimary : _darkTextPrimary;
  static Color get textSecondary =>
      isLight ? _lightTextSecondary : _darkTextSecondary;
  static Color get onAccent => const Color(0xFF1F2933);
  static Color get elevatedSurface =>
      isLight ? const Color(0xFFFFFBF5) : _darkSurface;
  static Color get softShadow => isLight
      ? const Color(0xFFD8C7AE).withValues(alpha: 0.28)
      : Colors.black.withValues(alpha: 0.35);
  static Color get overlayOnImage => isLight
      ? Colors.white.withValues(alpha: 0.78)
      : Colors.black.withValues(alpha: 0.38);

  static List<Color> get dailyGoalGradient => isLight
      ? const [Color(0xFFFFFBF5), Color(0xFFF0E4D1)]
      : const [Color(0xFF1A3A5C), Color(0xFF0F3460)];

  static List<Color> get featuredLessonGradient => isLight
      ? const [Color(0xFFFFF3DD), Color(0xFFE8C58C)]
      : const [Color(0xFF8B4513), Color(0xFFD2691E)];

  static List<Color> get lessonHeroGradient => isLight
      ? const [Color(0xFFFFF0D2), Color(0xFFE2B978), Color(0xFFF8F5EF)]
      : const [Color(0xFF8B4513), Color(0xFF2C1810), Color(0xFF1A1A2E)];

  static ThemeData get theme => isLight ? lightTheme : darkTheme;

  static ThemeData get darkTheme {
    return ThemeData(
      useMaterial3: true,
      scaffoldBackgroundColor: _darkPrimary,
      colorScheme: const ColorScheme.dark(
        primary: accent,
        secondary: accentLight,
        surface: _darkSurface,
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
      iconTheme: const IconThemeData(color: _darkTextSecondary),
      listTileTheme: ListTileThemeData(
        iconColor: _darkTextSecondary,
        textColor: _darkTextPrimary,
        titleTextStyle: GoogleFonts.lato(color: _darkTextPrimary, fontSize: 15),
        subtitleTextStyle:
            GoogleFonts.lato(color: _darkTextSecondary, fontSize: 13),
      ),
      dialogTheme: DialogThemeData(
        backgroundColor: _darkSurface,
        titleTextStyle: GoogleFonts.playfairDisplay(
            color: _darkTextPrimary, fontSize: 20, fontWeight: FontWeight.bold),
        contentTextStyle:
            GoogleFonts.lato(color: _darkTextSecondary, fontSize: 14),
      ),
      popupMenuTheme: PopupMenuThemeData(
        color: _darkSurface,
        textStyle: GoogleFonts.lato(color: _darkTextPrimary, fontSize: 14),
        iconColor: _darkTextSecondary,
      ),
      dropdownMenuTheme: DropdownMenuThemeData(
        textStyle: GoogleFonts.lato(color: _darkTextPrimary, fontSize: 14),
        inputDecorationTheme:
            _inputDecoration(_darkCardBg, _darkPrimary, _darkTextSecondary),
      ),
      // dropdownButtonTheme: DropdownButtonThemeData(
      //   dropdownColor: _darkSurface,
      //   style: GoogleFonts.lato(color: _darkTextPrimary, fontSize: 14),
      // ),
      inputDecorationTheme:
          _inputDecoration(_darkCardBg, _darkPrimary, _darkTextSecondary),
      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ElevatedButton.styleFrom(
          backgroundColor: accent,
          foregroundColor: onAccent,
          textStyle: GoogleFonts.lato(fontWeight: FontWeight.bold),
        ),
      ),
      outlinedButtonTheme: OutlinedButtonThemeData(
        style: OutlinedButton.styleFrom(
          foregroundColor: accent,
          side: const BorderSide(color: accent),
          textStyle: GoogleFonts.lato(fontWeight: FontWeight.bold),
        ),
      ),
      textButtonTheme: TextButtonThemeData(
        style: TextButton.styleFrom(
          foregroundColor: accent,
          textStyle: GoogleFonts.lato(fontWeight: FontWeight.bold),
        ),
      ),
      chipTheme: ChipThemeData(
        backgroundColor: _darkSurface,
        selectedColor: accent.withValues(alpha: 0.2),
        labelStyle: GoogleFonts.lato(color: _darkTextPrimary),
        secondaryLabelStyle: GoogleFonts.lato(color: _darkTextPrimary),
        side: const BorderSide(color: _darkCardBg),
      ),
      textSelectionTheme: TextSelectionThemeData(
        cursorColor: accent,
        selectionColor: accent.withValues(alpha: 0.28),
        selectionHandleColor: accent,
      ),
      dividerTheme: const DividerThemeData(color: _darkCardBg),
      snackBarTheme: SnackBarThemeData(
        backgroundColor: _darkSurface,
        contentTextStyle: GoogleFonts.lato(color: _darkTextPrimary),
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
        onPrimary: Color(0xFF1F2933),
        onSurface: lightText,
      ),
      textTheme: GoogleFonts.playfairDisplayTextTheme().copyWith(
        displayLarge: GoogleFonts.playfairDisplay(
            color: lightText, fontSize: 32, fontWeight: FontWeight.bold),
        displayMedium: GoogleFonts.playfairDisplay(
            color: lightText, fontSize: 24, fontWeight: FontWeight.bold),
        headlineMedium: GoogleFonts.playfairDisplay(
            color: lightText, fontSize: 20, fontWeight: FontWeight.w600),
        bodyLarge: GoogleFonts.lato(color: lightText, fontSize: 16),
        bodyMedium: GoogleFonts.lato(color: lightMuted, fontSize: 14),
        labelLarge: GoogleFonts.lato(
            color: lightText, fontSize: 16, fontWeight: FontWeight.bold),
      ),
      appBarTheme: AppBarTheme(
        backgroundColor: lightPrimary,
        elevation: 0,
        titleTextStyle: GoogleFonts.playfairDisplay(
            color: lightText, fontSize: 20, fontWeight: FontWeight.bold),
        iconTheme: const IconThemeData(color: accent),
      ),
      bottomNavigationBarTheme: const BottomNavigationBarThemeData(
        backgroundColor: lightSecondary,
        selectedItemColor: accent,
        unselectedItemColor: lightMuted,
        type: BottomNavigationBarType.fixed,
        elevation: 20,
      ),
      iconTheme: const IconThemeData(color: lightMuted),
      listTileTheme: ListTileThemeData(
        iconColor: lightMuted,
        textColor: lightText,
        titleTextStyle: GoogleFonts.lato(color: lightText, fontSize: 15),
        subtitleTextStyle: GoogleFonts.lato(color: lightMuted, fontSize: 13),
      ),
      dialogTheme: DialogThemeData(
        backgroundColor: lightSurface,
        titleTextStyle: GoogleFonts.playfairDisplay(
            color: lightText, fontSize: 20, fontWeight: FontWeight.bold),
        contentTextStyle: GoogleFonts.lato(color: lightMuted, fontSize: 14),
      ),
      popupMenuTheme: PopupMenuThemeData(
        color: lightSurface,
        textStyle: GoogleFonts.lato(color: lightText, fontSize: 14),
        iconColor: lightMuted,
      ),
      dropdownMenuTheme: DropdownMenuThemeData(
        textStyle: GoogleFonts.lato(color: lightText, fontSize: 14),
        inputDecorationTheme:
            _inputDecoration(lightCard, lightSurface, lightMuted),
      ),
      // dropdownButtonTheme: DropdownButtonThemeData(
      //   dropdownColor: lightSurface,
      //   style: GoogleFonts.lato(color: lightText, fontSize: 14),
      // ),
      cardTheme: const CardThemeData(color: lightSurface),
      inputDecorationTheme:
          _inputDecoration(lightCard, lightSurface, lightMuted),
      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ElevatedButton.styleFrom(
          backgroundColor: accent,
          foregroundColor: onAccent,
          textStyle: GoogleFonts.lato(fontWeight: FontWeight.bold),
        ),
      ),
      outlinedButtonTheme: OutlinedButtonThemeData(
        style: OutlinedButton.styleFrom(
          foregroundColor: accent,
          side: const BorderSide(color: accent),
          textStyle: GoogleFonts.lato(fontWeight: FontWeight.bold),
        ),
      ),
      textButtonTheme: TextButtonThemeData(
        style: TextButton.styleFrom(
          foregroundColor: accent,
          textStyle: GoogleFonts.lato(fontWeight: FontWeight.bold),
        ),
      ),
      chipTheme: ChipThemeData(
        backgroundColor: lightSurface,
        selectedColor: accent.withValues(alpha: 0.18),
        labelStyle: GoogleFonts.lato(color: lightText),
        secondaryLabelStyle: GoogleFonts.lato(color: lightText),
        side: const BorderSide(color: lightCard),
      ),
      textSelectionTheme: TextSelectionThemeData(
        cursorColor: accent,
        selectionColor: accent.withValues(alpha: 0.24),
        selectionHandleColor: accent,
      ),
      dividerTheme: const DividerThemeData(color: lightCard),
      snackBarTheme: SnackBarThemeData(
        backgroundColor: lightSurface,
        contentTextStyle: GoogleFonts.lato(color: lightText),
      ),
    );
  }

  static InputDecorationTheme _inputDecoration(
      Color borderColor, Color fillColor, Color hintColor) {
    final border = OutlineInputBorder(
      borderSide: BorderSide(color: borderColor),
      borderRadius: BorderRadius.circular(8),
    );
    return InputDecorationTheme(
      filled: true,
      fillColor: fillColor,
      hintStyle: GoogleFonts.lato(color: hintColor),
      labelStyle: GoogleFonts.lato(color: hintColor),
      enabledBorder: border,
      focusedBorder:
          border.copyWith(borderSide: const BorderSide(color: accent)),
      errorBorder: border.copyWith(borderSide: const BorderSide(color: error)),
      focusedErrorBorder:
          border.copyWith(borderSide: const BorderSide(color: error)),
    );
  }
}
