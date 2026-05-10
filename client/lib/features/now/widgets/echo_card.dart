import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/theme/colors.dart';
import '../../../data/generated/api.pb.dart' as pb;
import '../providers/now_page_provider.dart';

class EchoSection extends ConsumerStatefulWidget {
  const EchoSection({super.key});

  @override
  ConsumerState<EchoSection> createState() => _EchoSectionState();
}

class _EchoSectionState extends ConsumerState<EchoSection> {
  bool _showConnector = false;
  bool _showInsight = false;
  Timer? _connectorTimer;
  Timer? _insightTimer;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final state = ref.read(nowPageProvider);
    if (state.status == NowPageStatus.echoing) {
      _scheduleStaggers();
    }
  }

  void _scheduleStaggers() {
    _connectorTimer?.cancel();
    _insightTimer?.cancel();
    _showConnector = false;
    _showInsight = false;
    // Match prototype: connector at 900ms, insight at 1600ms after echo appears
    _connectorTimer = Timer(const Duration(milliseconds: 900), () {
      if (mounted) setState(() => _showConnector = true);
    });
    _insightTimer = Timer(const Duration(milliseconds: 1600), () {
      if (mounted) setState(() => _showInsight = true);
    });
  }

  @override
  void dispose() {
    _connectorTimer?.cancel();
    _insightTimer?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(nowPageProvider);
    final visible = state.status == NowPageStatus.echoing && !state.isLoading;
    final echo = state.echo;

    // Start stagger timers when echo becomes visible
    if (visible && !_showConnector && !_showInsight) {
      WidgetsBinding.instance.addPostFrameCallback((_) => _scheduleStaggers());
    }

    return AnimatedPositioned(
      duration: const Duration(milliseconds: 500),
      curve: Curves.easeOut,
      left: 24,
      right: 24,
      top: visible ? 100 : -600,
      bottom: 0,
      child: AnimatedOpacity(
        duration: const Duration(milliseconds: 400),
        opacity: visible ? 1.0 : 0.0,
        child: visible
            ? SingleChildScrollView(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    _EchoBox(
                      echo: echo,
                      firstMatchedContent: state.matchedMoments.isNotEmpty
                          ? state.matchedMoments.first.content
                          : null,
                    ),
                    const SizedBox(height: 16),
                    if (echo != null && echo.matchedMomentIds.length > 1) ...[
                      _CandidateToggle(
                        echo: echo,
                        matchedMoments: state.matchedMoments,
                      ),
                      const SizedBox(height: 16),
                    ],
                    // Connector line (stagger: 900ms)
                    _ConnectorLine(show: _showConnector),
                    const SizedBox(height: 16),
                    // Insight card (stagger: 1600ms)
                    _InsightSlot(show: _showInsight, insight: state.insight),
                    const SizedBox(height: 20),
                    _EchoActions(
                      onContinue: () =>
                          ref.read(nowPageProvider.notifier).reopenWhisper(),
                      onStash: () =>
                          ref.read(nowPageProvider.notifier).beginStash(),
                      onDismiss: () =>
                          ref.read(nowPageProvider.notifier).dismissEcho(),
                    ),
                    const SizedBox(height: 40),
                  ],
                ),
              )
            : const SizedBox.shrink(),
      ),
    );
  }
}

class _EchoBox extends StatelessWidget {
  final pb.Echo? echo;
  final String? firstMatchedContent;

  const _EchoBox({required this.echo, this.firstMatchedContent});

  @override
  Widget build(BuildContext context) {
    final e = echo;
    final hasMatches = e != null && e.matchedMomentIds.isNotEmpty;

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(28),
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: Colors.white.withValues(alpha: 0.08)),
        color: Colors.white.withValues(alpha: 0.03),
      ),
      child: Column(
        children: [
          if (hasMatches) ...[
            const Text(
              '你之前也说过类似的',
              style: TextStyle(
                fontSize: 11,
                color: Color(0xFF5A5A70),
                letterSpacing: 1.5,
                fontWeight: FontWeight.w200,
              ),
            ),
            const SizedBox(height: 14),
            if (firstMatchedContent != null)
              Text(
                firstMatchedContent!,
                style: const TextStyle(
                  fontSize: 15,
                  color: Color(0xFFD8D8E8),
                  height: 1.8,
                  fontWeight: FontWeight.w300,
                  fontStyle: FontStyle.italic,
                ),
              ),
          ] else
            const Text(
              '回声在你过去的话里找到了自己',
              style: TextStyle(
                fontSize: 15,
                color: Color(0xFFD8D8E8),
                height: 1.8,
                fontWeight: FontWeight.w300,
                fontStyle: FontStyle.italic,
              ),
            ),
        ],
      ),
    );
  }
}

class _CandidateToggle extends StatefulWidget {
  final pb.Echo echo;
  final List<pb.Moment> matchedMoments;

  const _CandidateToggle({required this.echo, required this.matchedMoments});

  @override
  State<_CandidateToggle> createState() => _CandidateToggleState();
}

class _CandidateToggleState extends State<_CandidateToggle> {
  bool _expanded = false;

  @override
  Widget build(BuildContext context) {
    final ids = widget.echo.matchedMomentIds;
    final sims = widget.echo.similarities;
    final momentsMap = <String, pb.Moment>{};
    for (final m in widget.matchedMoments) {
      momentsMap[m.id] = m;
    }

    return Column(
      children: [
        GestureDetector(
          onTap: () => setState(() => _expanded = !_expanded),
          child: Padding(
            padding: const EdgeInsets.symmetric(vertical: 10),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Text(
                  _expanded ? '收起' : '之前的你还说过 ›',
                  style: const TextStyle(
                    fontSize: 11,
                    color: Color(0xFF7A7A90),
                    letterSpacing: 1,
                    fontWeight: FontWeight.w200,
                  ),
                ),
                const SizedBox(width: 4),
                AnimatedRotation(
                  turns: _expanded ? 0.5 : 0,
                  duration: const Duration(milliseconds: 200),
                  child: const Icon(
                    Icons.keyboard_arrow_down,
                    size: 14,
                    color: Color(0xFF7A7A90),
                  ),
                ),
              ],
            ),
          ),
        ),
        AnimatedSize(
          duration: const Duration(milliseconds: 300),
          curve: Curves.easeInOut,
          alignment: Alignment.topCenter,
          child: _expanded
              ? Column(
                  children: List.generate(ids.length, (i) {
                    final sim = i < sims.length
                        ? (sims[i] * 100).toStringAsFixed(0)
                        : '--';
                    final matchedMoment = momentsMap[ids[i]];
                    return Padding(
                      padding: const EdgeInsets.only(bottom: 8),
                      child: Container(
                        width: double.infinity,
                        padding: const EdgeInsets.symmetric(
                          horizontal: 14,
                          vertical: 12,
                        ),
                        decoration: BoxDecoration(
                          borderRadius: BorderRadius.circular(12),
                          border: Border.all(
                            color: Colors.white.withValues(alpha: 0.05),
                          ),
                          color: Colors.white.withValues(alpha: 0.02),
                        ),
                        child: Row(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Container(
                              width: 6,
                              height: 6,
                              margin: const EdgeInsets.only(top: 5),
                              decoration: BoxDecoration(
                                shape: BoxShape.circle,
                                color: AppColors.warmGold.withValues(
                                  alpha: 0.4 + i * 0.15,
                                ),
                              ),
                            ),
                            const SizedBox(width: 10),
                            Expanded(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  if (matchedMoment != null)
                                    Text(
                                      matchedMoment.content,
                                      style: const TextStyle(
                                        fontSize: 13,
                                        color: Color(0xFFB8B8C8),
                                        height: 1.6,
                                        fontWeight: FontWeight.w300,
                                        fontStyle: FontStyle.italic,
                                      ),
                                    ),
                                  const SizedBox(height: 6),
                                  Text(
                                    '${sim}% 相似',
                                    style: const TextStyle(
                                      fontSize: 11,
                                      color: Color(0xFF6A6A80),
                                      fontWeight: FontWeight.w200,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ],
                        ),
                      ),
                    );
                  }),
                )
              : const SizedBox.shrink(),
        ),
      ],
    );
  }
}

class _InsightSlot extends StatelessWidget {
  final bool show;
  final pb.Insight? insight;

  const _InsightSlot({required this.show, this.insight});

  @override
  Widget build(BuildContext context) {
    final ins = insight;
    return AnimatedOpacity(
      duration: const Duration(milliseconds: 600),
      opacity: show && ins != null ? 1.0 : 0.0,
      child: AnimatedSlide(
        duration: const Duration(milliseconds: 600),
        curve: Curves.easeOut,
        offset: show && ins != null ? Offset.zero : const Offset(0, 0.1),
        child: show && ins != null
            ? Container(
                width: double.infinity,
                padding: const EdgeInsets.all(22),
                decoration: BoxDecoration(
                  borderRadius: BorderRadius.circular(18),
                  border: Border.all(
                    color: AppColors.gold.withValues(alpha: 0.25),
                  ),
                  gradient: LinearGradient(
                    colors: [
                      AppColors.gold.withValues(alpha: 0.08),
                      AppColors.softPurple.withValues(alpha: 0.05),
                    ],
                  ),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text(
                      '✦ 我发现',
                      style: TextStyle(
                        fontSize: 10,
                        color: Color(0xFFCCA880),
                        letterSpacing: 2.5,
                        fontWeight: FontWeight.w200,
                      ),
                    ),
                    const SizedBox(height: 12),
                    Text(
                      ins.text,
                      style: const TextStyle(
                        fontSize: 14,
                        color: Color(0xFFF0E6D2),
                        height: 1.8,
                        fontWeight: FontWeight.w300,
                      ),
                    ),
                  ],
                ),
              )
            : const SizedBox.shrink(),
      ),
    );
  }
}

class _ConnectorLine extends StatelessWidget {
  final bool show;

  const _ConnectorLine({required this.show});

  @override
  Widget build(BuildContext context) {
    return AnimatedOpacity(
      duration: const Duration(milliseconds: 500),
      opacity: show ? 1.0 : 0.0,
      child: SizedBox(
        width: 1,
        height: 24,
        child: Container(
          decoration: BoxDecoration(
            gradient: LinearGradient(
              begin: Alignment.topCenter,
              end: Alignment.bottomCenter,
              colors: [
                Colors.transparent,
                AppColors.gold.withValues(alpha: 0.6),
                Colors.transparent,
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _EchoActions extends StatelessWidget {
  final VoidCallback onContinue;
  final VoidCallback onStash;
  final VoidCallback onDismiss;

  const _EchoActions({
    required this.onContinue,
    required this.onStash,
    required this.onDismiss,
  });

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        SizedBox(
          width: double.infinity,
          child: ElevatedButton(
            onPressed: onContinue,
            style: ElevatedButton.styleFrom(
              backgroundColor: AppColors.gold.withValues(alpha: 0.12),
              foregroundColor: AppColors.warmGold,
              padding: const EdgeInsets.symmetric(vertical: 13),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(26),
                side: BorderSide(
                  color: AppColors.gold.withValues(alpha: 0.35),
                ),
              ),
              textStyle: const TextStyle(
                fontSize: 14,
                fontWeight: FontWeight.w300,
                letterSpacing: 1,
              ),
            ),
            child: const Text('顺着再想想'),
          ),
        ),
        const SizedBox(height: 14),
        TextButton(
          onPressed: onStash,
          style: TextButton.styleFrom(
            foregroundColor: const Color(0xFFA8A8C0),
            textStyle: const TextStyle(
              fontSize: 12,
              fontWeight: FontWeight.w300,
              letterSpacing: 1.5,
            ),
          ),
          child: const Text('✦ 收进星图'),
        ),
        TextButton(
          onPressed: onDismiss,
          style: TextButton.styleFrom(
            foregroundColor: const Color(0xFF5A5A70),
            textStyle: const TextStyle(
              fontSize: 11,
              fontWeight: FontWeight.w200,
              letterSpacing: 1.5,
            ),
          ),
          child: const Text('嗯，先这样'),
        ),
      ],
    );
  }
}
