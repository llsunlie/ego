import 'package:flutter/material.dart';
import 'colors.dart';

ThemeData darkTheme() {
  // Get Material 3 dark defaults and apply local fonts to ALL text styles.
  // This ensures every TextStyle (bodyLarge, headlineLarge, titleMedium, etc.)
  // inherits NotoSansSC + NotoSansSymbols2 without per-widget declarations.
  final baseTextTheme = ThemeData(brightness: Brightness.dark).textTheme.apply(
    fontFamily: 'NotoSansSC',
    fontFamilyFallback: const ['NotoSansSymbols2'],
  );

  return ThemeData(
    fontFamily: 'NotoSansSC',
    fontFamilyFallback: const ['NotoSansSymbols2'],
    brightness: Brightness.dark,
    scaffoldBackgroundColor: AppColors.darkBg,
    colorScheme: ColorScheme.dark(
      primary: AppColors.gold,
      secondary: AppColors.coldBlue,
      surface: AppColors.surface,
    ),
    textTheme: baseTextTheme.copyWith(
      bodyLarge: const TextStyle(color: AppColors.textPrimary, fontSize: 16),
      bodyMedium: const TextStyle(color: AppColors.textSecondary, fontSize: 14),
    ),
    inputDecorationTheme: InputDecorationTheme(
      filled: true,
      fillColor: AppColors.surface,
      border: OutlineInputBorder(
        borderRadius: BorderRadius.circular(12),
        borderSide: BorderSide.none,
      ),
      hintStyle: const TextStyle(color: AppColors.textSecondary),
    ),
    elevatedButtonTheme: ElevatedButtonThemeData(
      style: ElevatedButton.styleFrom(
        backgroundColor: AppColors.gold,
        foregroundColor: Colors.black,
        minimumSize: const Size(double.infinity, 52),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(12),
        ),
      ),
    ),
  );
}
