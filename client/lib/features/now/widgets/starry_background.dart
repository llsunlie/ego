import 'dart:math';
import 'package:flutter/material.dart';
import '../../../core/constants.dart';

class StarryBackground extends StatefulWidget {
  const StarryBackground({super.key});

  @override
  State<StarryBackground> createState() => _StarryBackgroundState();
}

class _StarryBackgroundState extends State<StarryBackground>
    with TickerProviderStateMixin {
  late final AnimationController _ctrl;
  final _stars = <_Star>[];
  final _rng = Random(42);

  @override
  void initState() {
    super.initState();
    _ctrl = AnimationController(
      vsync: this,
      duration: const Duration(seconds: 4),
    )..repeat();

    for (int i = 0; i < AppConstants.starCount; i++) {
      _stars.add(_Star(
        x: _rng.nextDouble(),
        y: _rng.nextDouble(),
        radius: 0.5 + _rng.nextDouble() * 1.2,
        phase: _rng.nextDouble() * 2 * pi,
        speed: 0.5 + _rng.nextDouble() * 1.5,
      ));
    }
  }

  @override
  void dispose() {
    _ctrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Positioned.fill(
      child: RepaintBoundary(
        child: AnimatedBuilder(
          animation: _ctrl,
          builder: (_, __) {
            return CustomPaint(
              painter: _StarryPainter(
                stars: _stars,
                time: _ctrl.value * 2 * pi,
              ),
            );
          },
        ),
      ),
    );
  }
}

class _Star {
  final double x;
  final double y;
  final double radius;
  final double phase;
  final double speed;
  const _Star({
    required this.x,
    required this.y,
    required this.radius,
    required this.phase,
    required this.speed,
  });
}

class _StarryPainter extends CustomPainter {
  final List<_Star> stars;
  final double time;

  _StarryPainter({required this.stars, required this.time});

  @override
  void paint(Canvas canvas, Size size) {
    for (final star in stars) {
      final alpha = 0.1 + 0.3 * (0.5 + 0.5 * sin(time * star.speed + star.phase));
      final paint = Paint()
        ..color = Colors.white.withValues(alpha: alpha)
        ..maskFilter = const MaskFilter.blur(BlurStyle.normal, 1);
      canvas.drawCircle(
        Offset(star.x * size.width, star.y * size.height),
        star.radius,
        paint,
      );
    }
  }

  @override
  bool shouldRepaint(_StarryPainter old) => old.time != time;
}
