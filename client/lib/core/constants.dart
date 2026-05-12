class AppConstants {
  // 编译时注入：本地默认 localhost，CD 工作流通过 --dart-define=SERVER_HOST=IP 传入
  static const String serverHost = String.fromEnvironment('SERVER_HOST', defaultValue: 'localhost');
  static const int serverPort = 9443;
  static const int serverWebPort = 9080;

  // Animation durations
  static const Duration animationFast = Duration(milliseconds: 200);
  static const Duration animationDefault = Duration(milliseconds: 300);
  static const Duration animationSlow = Duration(milliseconds: 500);

  // Breathing light
  static const Duration breathDuration = Duration(milliseconds: 3000);
  static const Duration haloMorphDuration = Duration(milliseconds: 14000);

  // Writing area
  static const Duration writingSlideDuration = Duration(milliseconds: 350);

  // Memory orbs
  static const Duration orbFloat1Duration = Duration(seconds: 20);
  static const Duration orbFloat2Duration = Duration(seconds: 25);
  static const Duration orbFloat3Duration = Duration(seconds: 22);
  static const Duration orbPulseDuration = Duration(seconds: 3);

  // Envelope
  static const Duration envelopeOpenDuration = Duration(milliseconds: 800);

  // Stash animation stages
  static const Duration cardGlowDuration = Duration(milliseconds: 300);
  static const Duration rippleDuration = Duration(milliseconds: 200);
  static const Duration rippleGap = Duration(milliseconds: 100);
  static const Duration flyDuration = Duration(milliseconds: 600);
  static const Duration starburstDuration = Duration(milliseconds: 300);
  static const Duration tabPulseDuration = Duration(milliseconds: 500);

  // Starfield
  static const int starCount = 80;

  // Chat
  static const double chatSheetMaxHeight = 0.85;
}

