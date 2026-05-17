import 'dart:async';
import 'dart:math';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../core/theme/colors.dart';
import '../../core/providers/tab_provider.dart';
import '../../data/repositories/local_store.dart';
import 'data/star_position.dart';
import 'painters/star_field_painter.dart';
import 'providers/starmap_provider.dart';

class StarmapPage extends ConsumerStatefulWidget {
  const StarmapPage({super.key});

  @override
  ConsumerState<StarmapPage> createState() => _StarmapPageState();
}

class _StarmapPageState extends ConsumerState<StarmapPage>
    with TickerProviderStateMixin {
  late final AnimationController _twinkleCtrl;
  List<PlacedConstellation> _placed = [];
  bool _hasLoaded = false;
  bool _showTapGuide = false;
  Timer? _tapGuideTimer;

  @override
  void initState() {
    super.initState();
    _twinkleCtrl = AnimationController(
      vsync: this,
      duration: const Duration(seconds: 5),
    )..repeat();
  }

  @override
  void dispose() {
    _twinkleCtrl.dispose();
    _tapGuideTimer?.cancel();
    super.dispose();
  }

  void _dismissTapGuide() {
    _tapGuideTimer?.cancel();
    if (_showTapGuide) {
      setState(() => _showTapGuide = false);
      LocalStore.setStarmapTapGuideShown(true);
    }
  }

  void _activateTapGuide() {
    if (_showTapGuide || LocalStore.getStarmapTapGuideShown()) return;
    setState(() => _showTapGuide = true);
    _tapGuideTimer = Timer(const Duration(seconds: 3), () {
      if (mounted) _dismissTapGuide();
    });
  }

  void _tryTap(Offset screenPos, Size canvasSize) {
    final scale = min(
      canvasSize.width / StarPositionEngine.worldWidth,
      canvasSize.height / StarPositionEngine.worldHeight,
    );
    final hitR = StarPositionEngine.hitRadius * scale;

    PlacedConstellation? best;
    double bestDist = double.infinity;
    for (final pc in _placed) {
      // Check star positions (scaled to screen space)
      for (final sp in pc.starPositions) {
        final screenPos2 = Offset(sp.dx * scale, sp.dy * scale);
        final dist = (screenPos - screenPos2).distance;
        if (dist < hitR && dist < bestDist) {
          best = pc;
          bestDist = dist;
        }
      }
      // Also check label area below center (for 3+ star constellations)
      if (pc.constellation.starCount >= 3) {
        final labelPos = Offset(pc.center.dx * scale, (pc.center.dy + 30) * scale);
        final dist = (screenPos - labelPos).distance;
        if (dist < hitR * 2 && dist < bestDist) {
          best = pc;
          bestDist = dist;
        }
      }
    }
    if (best != null) {
      context.push('/starmap/detail/${best.constellation.id}');
    }
  }

  @override
  Widget build(BuildContext context) {
    final tabIndex = ref.watch(tabProvider);

    if (tabIndex == 2 && !_hasLoaded) {
      _hasLoaded = true;
      WidgetsBinding.instance.addPostFrameCallback((_) {
        ref.read(starmapProvider.notifier).loadConstellations();
      });
    }
    if (tabIndex != 2) {
      _hasLoaded = false;
    }

    final state = ref.watch(starmapProvider);
    _placed = StarPositionEngine.placeAll(state.constellations);

    return Scaffold(
      backgroundColor: AppColors.darkBg,
      body: SafeArea(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Page header — matches past_page style
            Padding(
              padding: const EdgeInsets.fromLTRB(24, 20, 24, 8),
              child: Text(
                '已有 ${state.totalStarCount} 颗星',
                style: TextStyle(
                  fontSize: 14,
                  color: Colors.white.withValues(alpha: 0.5),
                  fontWeight: FontWeight.w300,
                  letterSpacing: 1.5,
                ),
              ),
            ),

            // Starfield
            Expanded(
              child: _buildStarfield(state),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildStarfield(StarmapState state) {
    if (state.isLoading && _placed.isEmpty) {
      return const Center(
        child: CircularProgressIndicator(color: AppColors.gold),
      );
    }

    if (state.error != null && _placed.isEmpty) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(Icons.cloud_off, color: AppColors.textHint, size: 40),
            const SizedBox(height: 12),
            Text('加载失败',
                style: TextStyle(color: AppColors.textHint, fontSize: 14)),
            const SizedBox(height: 8),
            TextButton(
              onPressed: () =>
                  ref.read(starmapProvider.notifier).loadConstellations(),
              child:
                  const Text('重试', style: TextStyle(color: AppColors.gold)),
            ),
          ],
        ),
      );
    }

    if (_placed.isEmpty && !state.isLoading) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.auto_awesome,
                color: AppColors.textHint.withValues(alpha: 0.5), size: 48),
            const SizedBox(height: 16),
            if (state.totalStarCount == 0)
              const Text(
                '去「此刻」写点什么，让星星汇聚成星座',
                style: TextStyle(color: AppColors.textHint, fontSize: 13),
              ),
          ],
        ),
      );
    }

    _activateTapGuide();

    return LayoutBuilder(
      builder: (context, constraints) {
        final canvasSize = Size(constraints.maxWidth, constraints.maxHeight);
        return GestureDetector(
          onTapUp: (d) {
            if (_showTapGuide) {
              _dismissTapGuide();
            } else {
              _tryTap(d.localPosition, canvasSize);
            }
          },
          child: Stack(
            children: [
              AnimatedBuilder(
                animation: _twinkleCtrl,
                builder: (_, __) => CustomPaint(
                  size: canvasSize,
                  painter: StarFieldPainter(
                    constellations: _placed,
                    time: _twinkleCtrl.value * 2 * pi,
                    zoomScale: 1.0,
                    canvasSize: canvasSize,
                  ),
                ),
              ),
              if (_showTapGuide)
                Positioned(
                  top: 16,
                  left: 0,
                  right: 0,
                  child: IgnorePointer(
                    child: Center(
                      child: AnimatedOpacity(
                        opacity: _showTapGuide ? 1.0 : 0.0,
                        duration: const Duration(milliseconds: 400),
                        child: Container(
                          padding: const EdgeInsets.symmetric(
                              horizontal: 20, vertical: 10),
                          decoration: BoxDecoration(
                            color: Colors.black.withValues(alpha: 0.7),
                            borderRadius: BorderRadius.circular(20),
                            border: Border.all(
                              color: AppColors.gold.withValues(alpha: 0.3),
                            ),
                          ),
                          child: const Text(
                            '点击星座或星星查看详情',
                            style: TextStyle(
                              color: AppColors.gold,
                              fontSize: 13,
                              fontWeight: FontWeight.w400,
                            ),
                          ),
                        ),
                      ),
                    ),
                  ),
                ),
            ],
          ),
        );
      },
    );
  }
}
