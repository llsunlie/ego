import 'dart:math';
import 'dart:ui' as ui;
import 'package:flutter/material.dart';
import '../providers/now_page_provider.dart';

class BreathingLight extends StatefulWidget {
  final NowPageStatus status;

  const BreathingLight({super.key, required this.status});

  @override
  State<BreathingLight> createState() => _BreathingLightState();
}

class _BreathingLightState extends State<BreathingLight>
    with TickerProviderStateMixin {
  late final AnimationController _breathCtrl;
  late final AnimationController _morphCtrl;
  late final AnimationController _coreCtrl;

  @override
  void initState() {
    super.initState();
    _breathCtrl = AnimationController(
      vsync: this,
      duration: const Duration(seconds: 5),
    )..repeat(reverse: true);
    _morphCtrl = AnimationController(
      vsync: this,
      duration: const Duration(seconds: 15),
    )..repeat();
    _coreCtrl = AnimationController(
      vsync: this,
      duration: const Duration(seconds: 3),
    )..repeat(reverse: true);
  }

  @override
  void dispose() {
    _breathCtrl.dispose();
    _morphCtrl.dispose();
    _coreCtrl.dispose();
    super.dispose();
  }

  bool get _isIdle => widget.status == NowPageStatus.idle;

  @override
  Widget build(BuildContext context) {
    return AnimatedScale(
      scale: _isIdle ? 1.0 : 0.35,
      duration: const Duration(milliseconds: 800),
      curve: Curves.easeInOut,
      child: AnimatedSlide(
        offset: _isIdle ? Offset.zero : const Offset(0, -0.25),
        duration: const Duration(milliseconds: 800),
        curve: Curves.easeInOut,
        child: LayoutBuilder(
          builder: (context, constraints) {
            final orbSize = (constraints.maxWidth * 0.3).clamp(120.0, 200.0);
            return _LightMain(
              size: orbSize,
              breathCtrl: _breathCtrl,
              morphCtrl: _morphCtrl,
              coreCtrl: _coreCtrl,
            );
          },
        ),
      ),
    );
  }
}

class _LightMain extends StatelessWidget {
  final double size;
  final AnimationController breathCtrl;
  final AnimationController morphCtrl;
  final AnimationController coreCtrl;

  const _LightMain({
    required this.size,
    required this.breathCtrl,
    required this.morphCtrl,
    required this.coreCtrl,
  });

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: Listenable.merge([breathCtrl, morphCtrl, coreCtrl]),
      builder: (_, __) {
        final breath = breathCtrl.value; // 0→1→0
        final morph = morphCtrl.value;
        final core = coreCtrl.value;

        // Prototype: egoBreath scale 1.0 → 1.12, opacity 0.92 → 1.0
        final breathScale = 1.0 + breath * 0.12;
        final opacity = 0.92 + breath * 0.08;

        // Container sized for largest halo (1.8× orb), orb centered within
        final containerSize = size * 1.8;
        final orbOffset = (containerSize - size) / 2; // = size * 0.4

        return Opacity(
          opacity: opacity,
          child: Transform.scale(
            scale: breathScale,
            child: SizedBox(
              width: containerSize,
              height: containerSize,
              child: Stack(
                clipBehavior: Clip.none,
                children: [
                  // Halo contour 2 — matches .light-contour.c2: rgba(180,210,255,0.1), ellipse at 52% 50%, blur 18px
                  _HaloContour(
                    size: containerSize,
                    morph: morph + 0.5,
                    alpha: 0.10,
                    blurSigma: 9,
                    color: const Color(0xFFB4D2FF),
                    center: const Alignment(0.04, 0.0),
                  ),
                  // Halo contour 1 — matches .light-contour: rgba(160,200,255,0.2), ellipse at 48% 48%, blur 10px
                  Positioned(
                    top: orbOffset * 0.375,
                    left: orbOffset * 0.375,
                    child: _HaloContour(
                      size: size * 1.5,
                      morph: morph,
                      alpha: 0.20,
                      blurSigma: 5,
                      color: const Color(0xFFA0C8FF),
                      center: const Alignment(-0.04, -0.04),
                    ),
                  ),
                  // Main orb — matches .light-main (blur 3px) with inner core
                  Positioned(
                    top: orbOffset,
                    left: orbOffset,
                    child: ImageFiltered(
                      imageFilter: ui.ImageFilter.blur(sigmaX: 1.5, sigmaY: 1.5),
                      child: _OrbBody(size: size, morph: morph, core: core),
                    ),
                  ),
                ],
              ),
            ),
          ),
        );
      },
    );
  }
}

class _OrbBody extends StatelessWidget {
  final double size;
  final double morph;
  final double core;

  const _OrbBody({
    required this.size,
    required this.morph,
    required this.core,
  });

  @override
  Widget build(BuildContext context) {
    // Prototype: egoSoft border-radius morphing across 3 keyframes
    final brTL = _animRadius(morph, 0, [46, 54], [52, 48], [48, 52]);
    final brTR = _animRadius(morph, 1, [54, 48], [48, 52], [52, 48]);
    final brBR = _animRadius(morph, 2, [52, 48], [50, 50], [48, 52]);
    final brBL = _animRadius(morph, 3, [48, 52], [50, 50], [52, 48]);

    final hTL = _animRadius(morph, 0, [50, 48], [48, 52], [52, 48]);
    final hTR = _animRadius(morph, 1, [47, 52], [52, 48], [48, 52]);
    final hBR = _animRadius(morph, 3, [53, 50], [50, 52], [52, 48]);
    final hBL = _animRadius(morph, 0, [50, 52], [52, 48], [48, 52]);

    // Inner core pulse
    final coreScale = 0.9 + core * 0.25;
    final coreOpacity = 0.7 + core * 0.3;

    return Container(
      width: size,
      height: size,
      decoration: BoxDecoration(
        borderRadius: BorderRadius.only(
          topLeft: Radius.elliptical(brTL * size / 100, hTL * size / 100),
          topRight: Radius.elliptical(brTR * size / 100, hTR * size / 100),
          bottomRight: Radius.elliptical(brBR * size / 100, hBR * size / 100),
          bottomLeft: Radius.elliptical(brBL * size / 100, hBL * size / 100),
        ),
        gradient: const RadialGradient(
          center: Alignment(-0.04, -0.08),
          colors: [
            Colors.white,
            Color(0xE6DCF0FF),
            Color(0x99A0CDFF),
            Color(0x4078AFFA),
            Colors.transparent,
          ],
          stops: [0.0, 0.12, 0.30, 0.55, 0.80],
        ),
      ),
      child: Stack(
        children: [
          // Inner core — matches .light-main::before (coreGlow)
          Positioned(
            top: size * 0.25,
            left: size * 0.25,
            child: Opacity(
              opacity: coreOpacity,
              child: Transform.scale(
                scale: coreScale,
                child: Container(
                  width: size * 0.5,
                  height: size * 0.5,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    gradient: RadialGradient(
                      colors: [
                        Colors.white.withValues(alpha: 0.9),
                        const Color(0x66C8E1FF),
                        Colors.transparent,
                      ],
                      stops: const [0.0, 0.5, 1.0],
                    ),
                  ),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _HaloContour extends StatelessWidget {
  final double size;
  final double morph;
  final double alpha;
  final double blurSigma;
  final Color color;
  final Alignment center;

  const _HaloContour({
    required this.size,
    required this.morph,
    required this.alpha,
    required this.blurSigma,
    required this.color,
    required this.center,
  });

  @override
  Widget build(BuildContext context) {
    final rx = size * 0.45 + sin(morph * 2 * pi) * size * 0.04;
    final ry = size * 0.50 + cos(morph * 2 * pi + 0.3) * size * 0.03;

    return ImageFiltered(
      imageFilter: ui.ImageFilter.blur(sigmaX: blurSigma, sigmaY: blurSigma),
      child: Container(
        width: size,
        height: size,
        decoration: BoxDecoration(
          borderRadius: BorderRadius.all(Radius.elliptical(rx, ry)),
          gradient: RadialGradient(
            center: center,
            colors: [
              color.withValues(alpha: alpha),
              color.withValues(alpha: alpha * 0.4),
              Colors.transparent,
            ],
            stops: const [0.0, 0.55, 0.8],
          ),
        ),
      ),
    );
  }
}

double _animRadius(double t, int phase, List<double> a, List<double> b, List<double> c) {
  t = (t + phase * 0.25) % 1.0;
  if (t < 0.33) {
    return _lerp(a[0], b[0], t / 0.33);
  } else if (t < 0.66) {
    return _lerp(b[0], c[0], (t - 0.33) / 0.33);
  } else {
    return _lerp(c[0], a[0], (t - 0.66) / 0.34);
  }
}

double _lerp(double a, double b, double t) => a + (b - a) * t;
