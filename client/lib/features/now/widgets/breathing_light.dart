import 'dart:math';
import 'package:flutter/material.dart';
import '../../../core/constants.dart';
import '../../../core/theme/colors.dart';
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
  late final AnimationController _haloCtrl;

  @override
  void initState() {
    super.initState();
    _breathCtrl = AnimationController(
      vsync: this,
      duration: AppConstants.breathDuration,
    )..repeat(reverse: true);
    _haloCtrl = AnimationController(
      vsync: this,
      duration: AppConstants.haloMorphDuration,
    )..repeat();
  }

  @override
  void dispose() {
    _breathCtrl.dispose();
    _haloCtrl.dispose();
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
        child: AnimatedBuilder(
          animation: Listenable.merge([_breathCtrl, _haloCtrl]),
          builder: (_, __) {
            return CustomPaint(
              size: const Size(200, 200),
              painter: _BreathingLightPainter(
                breathValue: _breathCtrl.value,
                haloValue: _haloCtrl.value,
              ),
            );
          },
        ),
      ),
    );
  }
}

class _BreathingLightPainter extends CustomPainter {
  final double breathValue;
  final double haloValue;

  _BreathingLightPainter({
    required this.breathValue,
    required this.haloValue,
  });

  @override
  void paint(Canvas canvas, Size size) {
    final center = Offset(size.width / 2, size.height / 2);
    final baseRadius = size.width * 0.28;

    // Outer halo
    _drawHalo(canvas, center, baseRadius * 2.0, haloValue, 0);
    _drawHalo(canvas, center, baseRadius * 2.3, haloValue + pi, pi / 3);

    // Main blob
    final breathScale = 1.0 + breathValue * 0.15;
    final radius = baseRadius * breathScale;

    final blobPaint = Paint()
      ..shader = RadialGradient(
        center: const Alignment(-0.04, -0.08),
        colors: [
          Colors.white,
          AppColors.obLightBlue.withValues(alpha: 0.6),
          AppColors.obLightBlue.withValues(alpha: 0.0),
        ],
        stops: const [0.0, 0.3, 0.8],
      ).createShader(Rect.fromCircle(center: center, radius: radius * 1.5));

    final path = _blobPath(center, radius, breathValue, haloValue);
    canvas.drawPath(path, blobPaint);

    // Inner core
    final corePaint = Paint()
      ..shader = RadialGradient(
        colors: [
          Colors.white.withValues(alpha: 0.9),
          Colors.white.withValues(alpha: 0.0),
        ],
      ).createShader(Rect.fromCircle(
        center: center,
        radius: radius * 0.5 * (0.9 + breathValue * 0.2),
      ));
    canvas.drawCircle(center, radius * 0.5, corePaint);
  }

  void _drawHalo(
    Canvas canvas,
    Offset center,
    double radius,
    double time,
    double phaseOffset,
  ) {
    final morph = sin(time + phaseOffset);
    final rx = radius * (1.0 + morph * 0.08);
    final ry = radius * (1.0 - morph * 0.06);

    final paint = Paint()
      ..shader = RadialGradient(
        colors: [
          AppColors.obLightBlue.withValues(alpha: 0.12),
          AppColors.obLightBlue.withValues(alpha: 0.03),
          Colors.transparent,
        ],
        stops: const [0.0, 0.55, 0.8],
      ).createShader(Rect.fromCircle(center: center, radius: max(rx, ry)));
    canvas.drawOval(Rect.fromCenter(center: center, width: rx * 2, height: ry * 2), paint);
  }

  Path _blobPath(Offset center, double r, double breath, double halo) {
    final points = 8;
    final path = Path();
    for (int i = 0; i < points; i++) {
      final angle = (i / points) * 2 * pi;
      final dist = r * (0.85 + 0.15 * sin(angle * 3 + halo) * (0.7 + 0.3 * breath));
      final x = center.dx + cos(angle) * dist;
      final y = center.dy + sin(angle) * dist;
      if (i == 0) {
        path.moveTo(x, y);
      } else {
        final prevAngle = ((i - 1) / points) * 2 * pi;
        final prevDist = r *
            (0.85 + 0.15 * sin(prevAngle * 3 + halo) * (0.7 + 0.3 * breath));
        final cpAngle = (angle + prevAngle) / 2;
        final cpDist = (dist + prevDist) / 2 * 1.2;
        path.quadraticBezierTo(
          center.dx + cos(cpAngle) * cpDist,
          center.dy + sin(cpAngle) * cpDist,
          x,
          y,
        );
      }
    }
    path.close();
    return path;
  }

  @override
  bool shouldRepaint(_BreathingLightPainter old) =>
      old.breathValue != breathValue || old.haloValue != haloValue;
}
