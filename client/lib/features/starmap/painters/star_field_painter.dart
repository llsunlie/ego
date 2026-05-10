import 'dart:math';
import 'dart:ui' as ui;
import 'package:flutter/material.dart';
import '../../../core/theme/colors.dart';
import '../data/star_position.dart';

class StarFieldPainter extends CustomPainter {
  final List<PlacedConstellation> constellations;
  final double time;
  final double zoomScale;
  final Size? canvasSize; // if set, scales world coords to fit

  StarFieldPainter({
    required this.constellations,
    required this.time,
    required this.zoomScale,
    this.canvasSize,
  });

  static const _formedR = 5.0;
  static const _formingR = 4.0;
  static const _glowBlur = 4.0;
  static const _lineWidth = 1.0;
  static const _dashLen = 5.0;
  static const _gapLen = 4.0;

  static const _palette = [
    Color(0xFF8AB4FF),
    Color(0xFFFFD98A),
    Color(0xFFC8A0FF),
    Color(0xFFA0E6C8),
    Color(0xFFFFB4D8),
  ];

  Color _color(PlacedConstellation pc) =>
      _palette[(pc.constellation.id.hashCode & 0x7FFFFFFF) % _palette.length];

  double get _scale => canvasSize == null
      ? 1.0
      : min(canvasSize!.width / StarPositionEngine.worldWidth,
            canvasSize!.height / StarPositionEngine.worldHeight);

  Offset _s(Offset w) => _scale == 1.0 ? w : w * _scale;
  double _sd(double v) => v * _scale;

  double get _sr => (_formedR / zoomScale).clamp(2.0, _formedR);
  double get _fr => (_formingR / zoomScale).clamp(1.8, _formingR);
  double get _lw => (_lineWidth / zoomScale).clamp(0.5, _lineWidth);

  @override
  void paint(Canvas canvas, Size size) {
    _drawNebula(canvas, size);
    for (final pc in constellations) {
      final sc = pc.constellation.starCount;
      if (sc >= 3) {
        _drawFormed(canvas, pc);
      } else if (sc == 2) {
        _drawForming(canvas, pc);
      } else {
        _drawLone(canvas, pc);
      }
    }
  }

  // ─── Nebula (screen-space) ───────────────────────

  void _drawNebula(Canvas canvas, Size size) {
    final center = Offset(size.width / 2, size.height / 2);
    final r = size.shortestSide / 2;

    final centerPaint = Paint()
      ..shader = ui.Gradient.radial(
        center, r,
        [const Color(0x0A7B9EC7), Colors.transparent],
      );
    canvas.drawCircle(center, r, centerPaint);

    final purplePaint = Paint()
      ..shader = ui.Gradient.radial(
        Offset(size.width * 0.3, size.height * 0.6), r * 0.8,
        [const Color(0x089B8EC4), Colors.transparent],
      );
    canvas.drawCircle(Offset(size.width * 0.3, size.height * 0.6), r * 0.8, purplePaint);

    final rng = Random(42);
    final dotPaint = Paint()..color = Colors.white.withValues(alpha: 0.03);
    for (int i = 0; i < 120; i++) {
      final x = rng.nextDouble() * size.width;
      final y = rng.nextDouble() * size.height;
      canvas.drawCircle(Offset(x, y), rng.nextDouble() * 1.2, dotPaint);
    }
  }

  // ─── Lone star (1 star) ──────────────────────────

  void _drawLone(Canvas canvas, PlacedConstellation pc) {
    final color = _color(pc);
    final pulse = 0.5 + 0.5 * sin(time * 2.1 + pc.twinklePhase);
    final alpha = 0.2 + 0.35 * pulse;
    final scale = 0.85 + 0.25 * pulse;
    final pos = _s(pc.starPositions.isNotEmpty ? pc.starPositions[0] : pc.center);

    final glowPaint = Paint()
      ..color = color.withValues(alpha: alpha * 0.5)
      ..maskFilter = ui.MaskFilter.blur(ui.BlurStyle.normal, _glowBlur);
    canvas.drawCircle(pos, _fr * scale + 3, glowPaint);

    final corePaint = Paint()..color = color.withValues(alpha: alpha);
    canvas.drawCircle(pos, _fr * scale, corePaint);

    final brightPaint = Paint()
      ..color = Colors.white.withValues(alpha: alpha * 0.6);
    canvas.drawCircle(pos, _fr * scale * 0.4, brightPaint);
  }

  // ─── Forming (2 stars) ───────────────────────────

  void _drawForming(Canvas canvas, PlacedConstellation pc) {
    final color = _color(pc);
    final pulse = 0.5 + 0.5 * sin(time * 2.1 + pc.twinklePhase);
    final alpha = 0.2 + 0.35 * pulse;
    final scale = 0.85 + 0.25 * pulse;

    for (final pos in pc.starPositions) {
      final sp = _s(pos);
      final glowPaint = Paint()
        ..color = color.withValues(alpha: alpha * 0.5)
        ..maskFilter = ui.MaskFilter.blur(ui.BlurStyle.normal, _glowBlur);
      canvas.drawCircle(sp, _fr * scale + 3, glowPaint);

      final corePaint = Paint()..color = color.withValues(alpha: alpha);
      canvas.drawCircle(sp, _fr * scale, corePaint);

      final brightPaint = Paint()
        ..color = Colors.white.withValues(alpha: alpha * 0.6);
      canvas.drawCircle(sp, _fr * scale * 0.4, brightPaint);
    }

    // Connection line between the two stars — tap to open detail.
    _drawDirectLine(canvas, pc);
  }

  /// Simple straight line between two star positions (for 2-star forming state).
  void _drawDirectLine(Canvas canvas, PlacedConstellation pc) {
    if (pc.starPositions.length < 2) return;
    final lineAlpha = 0.25 + 0.2 * (0.5 + 0.5 * sin(time * 0.8 + pc.twinklePhase));
    final paint = Paint()
      ..color = AppColors.gold.withValues(alpha: lineAlpha)
      ..strokeWidth = _lw
      ..style = PaintingStyle.stroke
      ..strokeCap = StrokeCap.round;

    final p0 = _s(pc.starPositions[0]);
    final p1 = _s(pc.starPositions[1]);
    canvas.drawLine(p0, p1, paint);
  }

  // ─── Formed (3+ stars) ───────────────────────────

  void _drawFormed(Canvas canvas, PlacedConstellation pc) {
    _drawConnectionLines(canvas, pc);

    final color = _color(pc);

    for (int i = 0; i < pc.starPositions.length; i++) {
      final starPhase = _starPhase(pc.constellation.starIds[i]);
      final pulse = 0.5 + 0.5 * sin(time * 1.57 + starPhase);
      final alpha = 0.6 + 0.4 * pulse;
      final scale = 1.0 + 0.2 * pulse;
      final pos = _s(pc.starPositions[i]);

      final glowPaint = Paint()
        ..color = color.withValues(alpha: alpha * 0.5)
        ..maskFilter = ui.MaskFilter.blur(ui.BlurStyle.normal, _glowBlur);
      canvas.drawCircle(pos, _sr * scale + 3, glowPaint);

      final starPaint = Paint()..color = color.withValues(alpha: alpha);
      canvas.drawCircle(pos, _sr * scale, starPaint);

      final brightPaint = Paint()
        ..color = Colors.white.withValues(alpha: alpha * 0.5);
      canvas.drawCircle(pos, _sr * scale * 0.35, brightPaint);
    }

    final nameAlpha = 0.6 + 0.15 * sin(time * 0.8 + pc.twinklePhase);
    final labelBgPaint = Paint()
      ..color = const Color(0xBF19160F);
    final labelBorderPaint = Paint()
      ..color = const Color(0x1FFDC96)
      ..style = PaintingStyle.stroke
      ..strokeWidth = 0.5;

    final tp = TextPainter(
      text: TextSpan(
        text: pc.constellation.name,
        style: TextStyle(
          color: const Color(0xFFD4B88A).withValues(alpha: nameAlpha),
          fontSize: 11 / zoomScale.clamp(0.6, 1.5),
          fontWeight: FontWeight.w400,
        ),
      ),
      textDirection: ui.TextDirection.ltr,
    )..layout(maxWidth: 200);

    final sc = _s(pc.center);
    final labelRect = RRect.fromRectAndRadius(
      Rect.fromCenter(
        center: Offset(sc.dx, sc.dy + _sd(22 / zoomScale)),
        width: tp.width + 24,
        height: tp.height + 10,
      ),
      const Radius.circular(14),
    );
    canvas.drawRRect(labelRect, labelBgPaint);
    canvas.drawRRect(labelRect, labelBorderPaint);
    tp.paint(
      canvas,
      Offset(sc.dx - tp.width / 2, sc.dy + _sd(22 / zoomScale) - tp.height / 2),
    );
  }

  void _drawConnectionLines(Canvas canvas, PlacedConstellation pc) {
    if (pc.starPositions.length < 2) return;

    final path = Path();
    final first = _s(pc.starPositions.first);
    path.moveTo(first.dx, first.dy);
    for (int i = 1; i < pc.starPositions.length; i++) {
      final sp = _s(pc.starPositions[i]);
      path.lineTo(sp.dx, sp.dy);
    }

    final lineAlpha = 0.3 + 0.45 * (0.5 + 0.5 * sin(time * 1.05 + pc.twinklePhase));
    final linePaint = Paint()
      ..color = AppColors.gold.withValues(alpha: lineAlpha)
      ..strokeWidth = _lw
      ..style = PaintingStyle.stroke
      ..strokeCap = StrokeCap.round;

    final dashTotal = _dashLen + _gapLen;
    for (final metric in path.computeMetrics()) {
      double distance = 0;
      while (distance < metric.length) {
        final end = (distance + _dashLen).clamp(0, metric.length).toDouble();
        final segment = metric.extractPath(distance, end);
        canvas.drawPath(segment, linePaint);
        distance += dashTotal;
      }
    }
  }

  double _starPhase(String starId) {
    return ((starId.hashCode & 0xFFFF) / 0xFFFF) * 2 * pi;
  }

  @override
  bool shouldRepaint(covariant StarFieldPainter old) =>
      old.time != time ||
      old.zoomScale != zoomScale ||
      old.canvasSize != canvasSize ||
      old.constellations != constellations;
}
