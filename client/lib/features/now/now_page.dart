import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/theme/colors.dart';
import '../starmap/providers/starmap_provider.dart';
import 'providers/now_page_provider.dart';
import 'widgets/starry_background.dart';
import 'widgets/breathing_light.dart';
import 'widgets/memory_dot.dart';
import 'widgets/guide_section.dart';
import 'widgets/writing_input.dart';
import 'widgets/echo_card.dart';
import 'widgets/orbiting_satellites.dart';
import 'widgets/stash_animation.dart';

class NowPage extends ConsumerStatefulWidget {
  const NowPage({super.key});

  @override
  ConsumerState<NowPage> createState() => _NowPageState();
}

class _NowPageState extends ConsumerState<NowPage>
    with TickerProviderStateMixin {
  OverlayEntry? _stashOverlay;

  @override
  void dispose() {
    _removeStashOverlay();
    super.dispose();
  }

  void _removeStashOverlay() {
    _stashOverlay?.remove();
    _stashOverlay = null;
  }

  void _showStashOverlay() {
    if (_stashOverlay != null) return;
    _stashOverlay = OverlayEntry(
      builder: (_) => StashAnimation(
        onComplete: () {
          _removeStashOverlay();
          ref.read(nowPageProvider.notifier).completeStash();
        },
      ),
    );
    Overlay.of(context).insert(_stashOverlay!);
  }

  @override
  Widget build(BuildContext context) {
    ref.listen(nowPageProvider, (prev, next) {
      if (next.status == NowPageStatus.stashing &&
          prev?.status != NowPageStatus.stashing) {
        WidgetsBinding.instance.addPostFrameCallback(
          (_) => _showStashOverlay(),
        );
      }
    });

    ref.listen(pendingTopicPromptProvider, (prev, next) {
      if (next != null && prev != next) {
        ref.read(nowPageProvider.notifier).startWriting();
      }
    });

    final state = ref.watch(nowPageProvider);
    final isIdle = state.status == NowPageStatus.idle;

    return Scaffold(
      backgroundColor: AppColors.darkBg,
      body: SafeArea(
        child: Column(
          children: [
            Expanded(
              child: Stack(
                alignment: Alignment.center,
                children: [
                  const StarryBackground(),
                  // Breathing light
                  Center(
                    child: BreathingLight(status: state.status),
                  ),
                  // Orbiting light balls around the main light
                  Center(
                    child: OrbitingSatellites(dimmed: !isIdle),
                  ),
                  // Memory dots floating around the light
                  MemoryDotGroup(dimmed: !isIdle),
                  // Guide text below the light (idle only)
                  if (isIdle)
                    const Positioned(
                      top: 0,
                      left: 0,
                      right: 0,
                      bottom: 0,
                      child: Align(
                        alignment: Alignment(0, 0.55),
                        child: AnimatedOpacity(
                          duration: Duration(milliseconds: 400),
                          opacity: 1.0,
                          child: GuideText(),
                        ),
                      ),
                    ),
                  // Writing input (writing/echoing)
                  const WritingInput(),
                  // Echo section (echoing: echo + insight + actions)
                  const EchoSection(),
                ],
              ),
            ),
            // Write button at bottom (idle only)
            if (isIdle)
              WriteButton(
                onTap: () =>
                    ref.read(nowPageProvider.notifier).startWriting(),
              ),
          ],
        ),
      ),
    );
  }
}
