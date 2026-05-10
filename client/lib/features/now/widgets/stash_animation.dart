import 'dart:math';
import 'package:flutter/material.dart';

class StashAnimation extends StatefulWidget {
  final VoidCallback onComplete;

  const StashAnimation({super.key, required this.onComplete});

  @override
  State<StashAnimation> createState() => _StashAnimationState();
}

class _StashAnimationState extends State<StashAnimation>
    with TickerProviderStateMixin {
  late final AnimationController _ctrl;

  // Prototype timing:
  //   glow: 1200ms
  //   ripples: 0 / 250 / 500ms delay
  //   star rise: 200ms delay, 500ms animation
  //   star hover: 500ms
  //   star flight: 1600ms bezier
  //   starburst + tab pulse: 1400ms
  // Total: ~4200ms
  static const _totalMs = 4200;
  static const _rippleDelays = [0, 250, 500];
  static const _starRiseStart = 200;
  static const _starRiseEnd = 700;   // 200 + 500
  static const _hoverEnd = 1200;     // 700 + 500
  static const _flightEnd = 2800;    // 1200 + 1600
  static const _cleanupEnd = 4200;   // 2800 + 1400

  @override
  void initState() {
    super.initState();
    _ctrl = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: _totalMs),
    )..forward().then((_) => widget.onComplete());
  }

  @override
  void dispose() {
    _ctrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: _ctrl,
      builder: (_, __) {
        final t = _ctrl.value;
        final ms = t * _totalMs;
        final size = MediaQuery.of(context).size;
        final startPos = Offset(size.width / 2, size.height * 0.35);
        final endPos = Offset(size.width * 0.83, size.height - 30);

        // Glow phase (0 to _starRiseStart)
        final glowAlpha = ms < _starRiseStart ? 0.5 : (ms < _starRiseEnd ? (1 - (ms - _starRiseStart) / 500) * 0.5 : 0.0);

        return Stack(
          children: [
            // Card glow
            if (ms < _starRiseEnd)
              Positioned.fill(
                child: IgnorePointer(
                  child: CustomPaint(
                    painter: _GlowPainter(
                      center: startPos,
                      alpha: glowAlpha.clamp(0.0, 0.5),
                    ),
                  ),
                ),
              ),

            // 3 ripples from card center
            ...List.generate(3, (i) {
              final rippleStart = _rippleDelays[i];
              final rippleProgress =
                  ((ms - rippleStart) / 1200).clamp(0.0, 1.0);
              if (rippleProgress <= 0 || rippleProgress >= 1) {
                return const SizedBox.shrink();
              }
              return Positioned(
                left: startPos.dx - rippleProgress * 100,
                top: startPos.dy - rippleProgress * 100,
                child: IgnorePointer(
                  child: Container(
                    width: rippleProgress * 200,
                    height: rippleProgress * 200,
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      border: Border.all(
                        color: const Color(0xCCFFDC96)
                            .withValues(alpha: 1 - rippleProgress),
                        width: 2,
                      ),
                      boxShadow: [
                        BoxShadow(
                          color: const Color(0x99FFDC96)
                              .withValues(alpha: 1 - rippleProgress),
                          blurRadius: 30,
                          spreadRadius: 20 * rippleProgress,
                        ),
                      ],
                    ),
                  ),
                ),
              );
            }),

            // Flying star
            if (ms >= _starRiseStart && ms < _flightEnd)
              _FlyingStar(
                startPos: startPos,
                endPos: endPos,
                ms: ms,
                riseStart: _starRiseStart,
                riseEnd: _starRiseEnd,
                hoverEnd: _hoverEnd,
                flightEnd: _flightEnd,
              ),

            // Trail particles
            if (ms > _hoverEnd && ms < _flightEnd)
              _TrailParticles(
                startPos: startPos,
                endPos: endPos,
                progress: (ms - _hoverEnd) / (_flightEnd - _hoverEnd),
              ),

            // Starburst particles
            if (ms >= _flightEnd && ms < _cleanupEnd)
              _StarburstParticles(
                center: endPos,
                progress: (ms - _flightEnd) / (_cleanupEnd - _flightEnd),
              ),

            // Tab pulse
            if (ms >= _flightEnd && ms < _cleanupEnd)
              Positioned(
                left: endPos.dx - 15,
                top: endPos.dy - 15,
                child: IgnorePointer(
                  child: Container(
                    width: 30,
                    height: 30,
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      color: const Color(0x59FFDC78).withValues(
                        alpha: ms < _flightEnd + 400
                            ? (ms - _flightEnd) / 400
                            : (1 - (ms - _flightEnd - 400) / 1000)
                                .clamp(0.0, 1.0),
                      ),
                      boxShadow: [
                        BoxShadow(
                          color: const Color(0x99FFDC78).withValues(
                            alpha: ms < _flightEnd + 400
                                ? (ms - _flightEnd) / 400
                                : (1 - (ms - _flightEnd - 400) / 1000)
                                    .clamp(0.0, 1.0),
                          ),
                          blurRadius: 30,
                        ),
                      ],
                    ),
                  ),
                ),
              ),
          ],
        );
      },
    );
  }
}

class _FlyingStar extends StatelessWidget {
  final Offset startPos;
  final Offset endPos;
  final double ms;
  final int riseStart;
  final int riseEnd;
  final int hoverEnd;
  final int flightEnd;

  const _FlyingStar({
    required this.startPos,
    required this.endPos,
    required this.ms,
    required this.riseStart,
    required this.riseEnd,
    required this.hoverEnd,
    required this.flightEnd,
  });

  @override
  Widget build(BuildContext context) {
    final riseProgress = ((ms - riseStart) / (riseEnd - riseStart)).clamp(0.0, 1.0);
    final flightProgress = ((ms - hoverEnd) / (flightEnd - hoverEnd)).clamp(0.0, 1.0);

    double scale;
    double opacity;
    Offset pos;

    if (ms < riseEnd) {
      // Rise phase: scale 0.2 → 1.4, rises 40px
      final eased = Curves.easeOut.transform(riseProgress);
      scale = 0.2 + 1.2 * eased;
      opacity = eased;
      pos = Offset(
        startPos.dx,
        startPos.dy - 40 * eased,
      );
    } else if (ms < hoverEnd) {
      // Hover: scale 1.4, stays up 40px, breathing slightly
      final breathe = 1.0 + sin((ms - riseEnd) / 500 * pi) * 0.05;
      scale = 1.4 * breathe;
      opacity = 1.0;
      pos = Offset(startPos.dx, startPos.dy - 40);
    } else {
      // Flight: along bezier, shrinks 1.4 → 0.5
      final midX = (startPos.dx + endPos.dx) / 2;
      final midY = min(startPos.dy - 40, endPos.dy) - 120;
      final eased = Curves.easeInOut.transform(flightProgress);
      pos = _bezier(eased, Offset(startPos.dx, startPos.dy - 40), Offset(midX, midY), endPos);
      scale = 1.4 - 0.9 * eased;
      opacity = 1.0 - eased * 0.3;
    }

    return Positioned(
      left: pos.dx - 12,
      top: pos.dy - 12,
      child: IgnorePointer(
        child: Transform.scale(
          scale: scale,
          child: Container(
            width: 24,
            height: 24,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              boxShadow: [
                BoxShadow(
                  color: const Color(0xE6FFF0B4).withValues(alpha: opacity),
                  blurRadius: 30,
                  spreadRadius: 10,
                ),
                BoxShadow(
                  color: const Color(0x99FFC878).withValues(alpha: opacity),
                  blurRadius: 60,
                  spreadRadius: 20,
                ),
              ],
              gradient: const RadialGradient(
                colors: [
                  Colors.white,
                  Color(0xFFFDE0A0),
                  Color(0xFFFAB468),
                  Color(0x33FAB468),
                ],
                stops: [0.0, 0.3, 0.6, 0.85],
              ),
            ),
          ),
        ),
      ),
    );
  }

  Offset _bezier(double t, Offset p0, Offset p1, Offset p2) {
    final mt = 1 - t;
    return Offset(
      mt * mt * p0.dx + 2 * mt * t * p1.dx + t * t * p2.dx,
      mt * mt * p0.dy + 2 * mt * t * p1.dy + t * t * p2.dy,
    );
  }
}

class _TrailParticles extends StatefulWidget {
  final Offset startPos;
  final Offset endPos;
  final double progress;

  const _TrailParticles({
    required this.startPos,
    required this.endPos,
    required this.progress,
  });

  @override
  State<_TrailParticles> createState() => _TrailParticlesState();
}

class _TrailParticlesState extends State<_TrailParticles> {
  final List<_TrailDot> _trails = [];

  @override
  void didUpdateWidget(_TrailParticles old) {
    super.didUpdateWidget(old);
    // Spawn trail particles at current position
    if (widget.progress > 0 && widget.progress < 1) {
      final midX = (widget.startPos.dx + widget.endPos.dx) / 2;
      final midY = min(widget.startPos.dy - 40, widget.endPos.dy) - 120;
      final t = widget.progress;
      final mt = 1 - t;
      final x = mt * mt * widget.startPos.dx + 2 * mt * t * midX + t * t * widget.endPos.dx;
      final y = mt * mt * (widget.startPos.dy - 40) + 2 * mt * t * midY + t * t * widget.endPos.dy;
      _trails.add(_TrailDot(x: x, y: y, opacity: 0.8 - t * 0.3, scale: 1.0 - t * 0.5));
    }
  }

  @override
  Widget build(BuildContext context) {
    return Stack(
      children: _trails.map((t) {
        return Positioned(
          left: t.x - 5,
          top: t.y - 5,
          child: TweenAnimationBuilder<double>(
            tween: Tween(begin: t.opacity, end: 0.0),
            duration: const Duration(milliseconds: 800),
            onEnd: () {
              _trails.remove(t);
            },
            builder: (_, opacity, __) {
              return IgnorePointer(
                child: Transform.scale(
                  scale: t.scale * (opacity / t.opacity).clamp(0.2, 1.0),
                  child: Container(
                    width: 10,
                    height: 10,
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      color:
                          const Color(0xCCFFDC96).withValues(alpha: opacity),
                    ),
                  ),
                ),
              );
            },
          ),
        );
      }).toList(),
    );
  }
}

class _TrailDot {
  final double x;
  final double y;
  final double opacity;
  final double scale;
  _TrailDot({
    required this.x,
    required this.y,
    required this.opacity,
    required this.scale,
  });
}

class _StarburstParticles extends StatelessWidget {
  final Offset center;
  final double progress;

  const _StarburstParticles({
    required this.center,
    required this.progress,
  });

  @override
  Widget build(BuildContext context) {
    final rng = Random(42);
    return Stack(
      children: List.generate(16, (i) {
        final angle = (pi * 2 * i) / 16 + (rng.nextDouble() - 0.5) * 0.3;
        final distance = 40 + rng.nextDouble() * 30;
        final eased = Curves.easeOut.transform(progress.clamp(0.0, 1.0));
        final tx = cos(angle) * distance * eased;
        final ty = sin(angle) * distance * eased;
        final opacity = (1 - progress).clamp(0.0, 1.0);
        final scale = progress < 0.4 ? 1.0 + progress * 1.0 : (1 - (progress - 0.4) / 0.6).clamp(0.2, 1.4);

        return Positioned(
          left: center.dx + tx - 2,
          top: center.dy + ty - 2,
          child: IgnorePointer(
            child: Transform.scale(
              scale: scale,
              child: Container(
                width: 4,
                height: 4,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  color: const Color(0xCCFFFFFF).withValues(alpha: opacity),
                  boxShadow: [
                    BoxShadow(
                      color: const Color(0xCCFFDC96).withValues(alpha: opacity),
                      blurRadius: 8,
                    ),
                  ],
                ),
              ),
            ),
          ),
        );
      }),
    );
  }
}

class _GlowPainter extends CustomPainter {
  final Offset center;
  final double alpha;

  _GlowPainter({required this.center, required this.alpha});

  @override
  void paint(Canvas canvas, Size size) {
    if (alpha <= 0) return;
    final paint = Paint()
      ..shader = RadialGradient(
        colors: [
          const Color(0x80FFDC78).withValues(alpha: alpha),
          const Color(0x00FFDC78),
        ],
      ).createShader(Rect.fromCircle(center: center, radius: 200));
    canvas.drawCircle(center, 200, paint);
  }

  @override
  bool shouldRepaint(_GlowPainter old) =>
      old.center != center || old.alpha != alpha;
}
