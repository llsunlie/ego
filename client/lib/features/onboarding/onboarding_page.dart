import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/providers/onboarding_provider.dart';
import '../../core/theme/colors.dart';
import 'onboarding_data.dart';

class OnboardingPage extends ConsumerStatefulWidget {
  const OnboardingPage({super.key});

  @override
  ConsumerState<OnboardingPage> createState() => _OnboardingPageState();
}

class _OnboardingPageState extends ConsumerState<OnboardingPage>
    with TickerProviderStateMixin {
  int _step = 0; // 0=intro, 1=select, 2=diary, 3=insight, 4=preview
  int _feelingIdx = 0;
  int _diaryIdx = 0;

  late final AnimationController _breathCtrl;
  late final AnimationController _previewCtrl;

  @override
  void initState() {
    super.initState();
    _breathCtrl = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 3000),
    )..repeat(reverse: true);
    _previewCtrl = AnimationController(
      vsync: this,
      duration: const Duration(seconds: 16),
    );
  }

  @override
  void dispose() {
    _breathCtrl.dispose();
    _previewCtrl.dispose();
    super.dispose();
  }

  void _goToStep(int step) => setState(() => _step = step);

  void _selectFeeling(int idx) {
    _feelingIdx = idx;
    _diaryIdx = 0;
    _goToStep(2);
  }

  void _nextDiary() {
    final count = onboardingData[_feelingIdx].diary.length;
    setState(() => _diaryIdx = (_diaryIdx + 1) % count);
  }

  void _likeDiary() => _goToStep(3);

  void _toPreview() {
    _goToStep(4);
    _previewCtrl.forward(from: 0);
  }

  void _finish() {
    ref.read(onboardingCompleteProvider.notifier).complete();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.darkBg,
      body: SafeArea(
        child: AnimatedSwitcher(
          duration: const Duration(milliseconds: 500),
          switchInCurve: Curves.easeOut,
          switchOutCurve: Curves.easeIn,
          child: _buildStep(_step),
        ),
      ),
    );
  }

  Widget _buildStep(int step) {
    return switch (step) {
      0 => _StepIntro(onStart: () => _goToStep(1), breathCtrl: _breathCtrl),
      1 => _StepSelectFeeling(onSelect: _selectFeeling),
      2 => _StepDiary(
          data: onboardingData[_feelingIdx],
          diaryIdx: _diaryIdx,
          onAnother: _nextDiary,
          onLike: _likeDiary,
        ),
      3 => _StepInsight(
          data: onboardingData[_feelingIdx],
          onContinue: _toPreview,
        ),
      4 => _StepPreview(
          controller: _previewCtrl,
          onFinish: _finish,
        ),
      _ => const SizedBox.shrink(),
    };
  }
}

// ─── Step 0: Intro ─────────────────────────────────────────────

class _StepIntro extends StatelessWidget {
  final VoidCallback onStart;
  final AnimationController breathCtrl;

  const _StepIntro({required this.onStart, required this.breathCtrl});

  @override
  Widget build(BuildContext context) {
    return Container(
      key: const ValueKey('intro'),
      padding: const EdgeInsets.symmetric(horizontal: 32),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          AnimatedBuilder(
            animation: breathCtrl,
            builder: (_, child) => Transform.scale(
              scale: 1.0 + breathCtrl.value * 0.12,
              child: Container(
                width: 70,
                height: 70,
                decoration: BoxDecoration(
                  borderRadius: BorderRadius.circular(
                    32.2 + breathCtrl.value * 2.8,
                  ),
                  gradient: const RadialGradient(
                    center: Alignment(-0.04, -0.08),
                    colors: [
                      Colors.white,
                      Color(0x99A0CDFF),
                      Color(0x00A0CDFF),
                    ],
                    stops: [0.0, 0.3, 0.8],
                  ),
                  boxShadow: [
                    BoxShadow(
                      color: AppColors.obLightBlue.withValues(alpha: 0.3),
                      blurRadius: 20,
                    ),
                  ],
                ),
              ),
            ),
          ),
          const SizedBox(height: 28),
          const Text(
            'ego',
            style: TextStyle(
              fontSize: 22,
              color: AppColors.textPrimary,
              letterSpacing: 6,
              fontWeight: FontWeight.w200,
            ),
          ),
          const SizedBox(height: 20),
          const Text(
            '一个记得你说过什么的地方。\n'
            '你写下此刻的想法，它会在你过去的话里，\n'
            '找到一些你自己都忘了的东西还给你。',
            textAlign: TextAlign.center,
            style: TextStyle(
              fontSize: 13,
              color: Color(0xFFA8A8C0),
              height: 2,
              fontWeight: FontWeight.w300,
            ),
          ),
          const SizedBox(height: 16),
          const Text(
            '接下来是一段简短的模拟体验。\n你会感受到 ego 在做什么——只需要 1 分钟。',
            textAlign: TextAlign.center,
            style: TextStyle(
              fontSize: 11,
              color: Color(0xFF6A6A80),
              height: 1.8,
              fontWeight: FontWeight.w200,
              letterSpacing: 0.5,
            ),
          ),
          const SizedBox(height: 36),
          _PrimaryButton(label: '开始体验', onTap: onStart),
        ],
      ),
    );
  }
}

// ─── Step 1: Select Feeling ────────────────────────────────────

class _StepSelectFeeling extends StatelessWidget {
  final void Function(int) onSelect;

  const _StepSelectFeeling({required this.onSelect});

  @override
  Widget build(BuildContext context) {
    return Container(
      key: const ValueKey('select'),
      padding: const EdgeInsets.symmetric(horizontal: 24),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Text(
            '选一个最接近你最近状态的',
            style: TextStyle(
              fontSize: 14,
              color: Color(0xFFC8C8D8),
              letterSpacing: 1,
              fontWeight: FontWeight.w300,
            ),
          ),
          const SizedBox(height: 24),
          ...List.generate(onboardingFeelings.length, (i) {
            return Padding(
              padding: const EdgeInsets.only(bottom: 10),
              child: SizedBox(
                width: double.infinity,
                child: TextButton(
                  onPressed: () => onSelect(i),
                  style: TextButton.styleFrom(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 20,
                      vertical: 16,
                    ),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                      side: const BorderSide(
                        color: Color(0x22FFFFFF),
                      ),
                    ),
                    backgroundColor: const Color(0x08FFFFFF),
                  ),
                  child: Text(
                    '"${onboardingFeelings[i]}"',
                    style: const TextStyle(
                      fontSize: 13,
                      color: Color(0xFFC0C0D0),
                      fontWeight: FontWeight.w300,
                    ),
                    textAlign: TextAlign.center,
                  ),
                ),
              ),
            );
          }),
        ],
      ),
    );
  }
}

// ─── Step 2: Diary ─────────────────────────────────────────────

class _StepDiary extends StatelessWidget {
  final OnboardingGroup data;
  final int diaryIdx;
  final VoidCallback onAnother;
  final VoidCallback onLike;

  const _StepDiary({
    required this.data,
    required this.diaryIdx,
    required this.onAnother,
    required this.onLike,
  });

  @override
  Widget build(BuildContext context) {
    final diary = data.diary[diaryIdx];
    return Container(
      key: const ValueKey('diary'),
      padding: const EdgeInsets.symmetric(horizontal: 24),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          _DiaryCard(date: diary.date, text: diary.text),
          const SizedBox(height: 18),
          const Text(
            '这段话让你有感觉吗？',
            style: TextStyle(
              fontSize: 12,
              color: Color(0xFF7A7A90),
              letterSpacing: 2,
            ),
          ),
          const SizedBox(height: 14),
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              _GhostButton(label: '换一条看看', onTap: onAnother),
              const SizedBox(width: 10),
              _GhostButton(label: '有点像我', primary: true, onTap: onLike),
            ],
          ),
        ],
      ),
    );
  }
}

// ─── Step 3: Insight + Respond ────────────────────────────────

class _StepInsight extends StatelessWidget {
  final OnboardingGroup data;
  final VoidCallback onContinue;

  const _StepInsight({required this.data, required this.onContinue});

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      key: const ValueKey('insight'),
      padding: const EdgeInsets.symmetric(horizontal: 24),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const SizedBox(height: 48),
          _DiaryCard(date: data.diary2.date, text: data.diary2.text, small: true),
          const SizedBox(height: 20),
          Container(
            width: double.infinity,
            padding: const EdgeInsets.all(20),
            decoration: BoxDecoration(
              borderRadius: BorderRadius.circular(16),
              border: Border.all(
                color: const Color(0x33D4A853),
              ),
              boxShadow: const [
                BoxShadow(
                  color: Color(0x1AFFDC96),
                  blurRadius: 30,
                ),
              ],
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
                  ),
                ),
                const SizedBox(height: 12),
                Text(
                  data.insightFull,
                  style: const TextStyle(
                    fontSize: 13,
                    color: Color(0xFFF0E6D2),
                    height: 1.8,
                    fontWeight: FontWeight.w300,
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 20),
          const SizedBox(
            width: double.infinity,
            child: Text(
              '想接着说点什么吗？也可以不说。',
              style: TextStyle(
                fontSize: 13,
                color: Color(0xFF6A6A80),
                fontWeight: FontWeight.w200,
              ),
              textAlign: TextAlign.start,
            ),
          ),
          const SizedBox(height: 12),
          TextField(
            maxLines: 2,
            style: const TextStyle(
              color: AppColors.textPrimary,
              fontSize: 14,
              fontWeight: FontWeight.w300,
            ),
            decoration: InputDecoration(
              hintText: '随便说点什么……',
              hintStyle: const TextStyle(
                color: Color(0xFF4A4A60),
                fontWeight: FontWeight.w200,
              ),
              filled: true,
              fillColor: const Color(0x0AFFFFFF),
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(12),
                borderSide: const BorderSide(
                  color: Color(0x15FFFFFF),
                ),
              ),
              enabledBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(12),
                borderSide: const BorderSide(
                  color: Color(0x15FFFFFF),
                ),
              ),
              focusedBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(12),
                borderSide: const BorderSide(
                  color: Color(0x30D4A853),
                ),
              ),
              contentPadding: const EdgeInsets.all(16),
            ),
          ),
          const SizedBox(height: 12),
          SizedBox(
            width: double.infinity,
            child: _GhostButton(
              label: '继续',
              primary: true,
              onTap: onContinue,
              expand: true,
            ),
          ),
        ],
      ),
    );
  }
}

// ─── Step 4: Preview ───────────────────────────────────────────

class _StepPreview extends StatefulWidget {
  final AnimationController controller;
  final VoidCallback onFinish;

  const _StepPreview({required this.controller, required this.onFinish});

  @override
  State<_StepPreview> createState() => _StepPreviewState();
}

class _StepPreviewState extends State<_StepPreview> {
  bool _showButton = false;

  @override
  void initState() {
    super.initState();
    widget.controller.addListener(_onTick);
    widget.controller.forward();
  }

  void _onTick() {
    // Show button after all 7 lines have appeared (~13.5s / 16s = 0.84)
    if (widget.controller.value > 0.84 && !_showButton) {
      setState(() => _showButton = true);
    }
  }

  @override
  void dispose() {
    widget.controller.removeListener(_onTick);
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      key: const ValueKey('preview'),
      padding: const EdgeInsets.symmetric(horizontal: 24),
      child: AnimatedBuilder(
        animation: widget.controller,
        builder: (context, _) {
          return Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              ...List.generate(onboardingPreviewLines.length, (i) {
                final start = i * 1.8 / 16.0;
                final end = start + 0.2; // 0.2 = fade-in duration
                final t = widget.controller.value;
                final appear = t < start
                    ? 0.0
                    : t > end
                        ? 1.0
                        : (t - start) / (end - start);
                return Opacity(
                  opacity: appear,
                  child: Padding(
                    padding: const EdgeInsets.only(bottom: 12),
                    child: Text(
                      onboardingPreviewLines[i],
                      textAlign: TextAlign.center,
                      style: TextStyle(
                        fontSize: 14,
                        color: AppColors.textPrimary.withValues(
                          alpha: 0.8 + i * 0.03,
                        ),
                        fontWeight: FontWeight.w300,
                        height: 1.6,
                      ),
                    ),
                  ),
                );
              }),
              const SizedBox(height: 20),
              AnimatedOpacity(
                opacity: _showButton ? 1.0 : 0.0,
                duration: const Duration(milliseconds: 600),
                child: _PrimaryButton(
                  label: '开始我的第一条',
                  onTap: widget.onFinish,
                ),
              ),
            ],
          );
        },
      ),
    );
  }
}

// ─── Shared Widgets ────────────────────────────────────────────

class _PrimaryButton extends StatelessWidget {
  final String label;
  final VoidCallback onTap;

  const _PrimaryButton({required this.label, required this.onTap});

  @override
  Widget build(BuildContext context) {
    return ElevatedButton(
      onPressed: onTap,
      style: ElevatedButton.styleFrom(
        backgroundColor: AppColors.gold,
        foregroundColor: AppColors.darkBg,
        padding: const EdgeInsets.symmetric(horizontal: 36, vertical: 14),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(24),
        ),
        textStyle: const TextStyle(
          fontSize: 15,
          fontWeight: FontWeight.w500,
          letterSpacing: 2,
        ),
      ),
      child: Text(label),
    );
  }
}

class _GhostButton extends StatelessWidget {
  final String label;
  final bool primary;
  final VoidCallback onTap;
  final bool expand;

  const _GhostButton({
    required this.label,
    this.primary = false,
    required this.onTap,
    this.expand = false,
  });

  @override
  Widget build(BuildContext context) {
    final borderColor = primary
        ? AppColors.gold.withValues(alpha: 0.4)
        : const Color(0x20FFFFFF);
    final textColor =
        primary ? AppColors.warmGold : const Color(0xFF8B8B9E);

    return SizedBox(
      width: expand ? double.infinity : null,
      child: TextButton(
        onPressed: onTap,
        style: TextButton.styleFrom(
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(20),
            side: BorderSide(color: borderColor),
          ),
        ),
        child: Text(
          label,
          style: TextStyle(
            color: textColor,
            fontSize: 13,
            fontWeight: FontWeight.w300,
            letterSpacing: 1,
          ),
        ),
      ),
    );
  }
}

class _DiaryCard extends StatelessWidget {
  final String date;
  final String text;
  final bool small;

  const _DiaryCard({
    required this.date,
    required this.text,
    this.small = false,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: EdgeInsets.all(small ? 18 : 28),
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: const Color(0x15FFFFFF)),
        color: const Color(0x08FFFFFF),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            date,
            style: const TextStyle(
              fontSize: 10,
              color: Color(0xFF6A6A80),
              letterSpacing: 2,
              fontWeight: FontWeight.w200,
            ),
          ),
          const SizedBox(height: 12),
          Text(
            text,
            style: const TextStyle(
              fontSize: 14,
              color: Color(0xFFD8D8E8),
              height: 1.8,
              fontStyle: FontStyle.italic,
              fontWeight: FontWeight.w300,
            ),
          ),
        ],
      ),
    );
  }
}
