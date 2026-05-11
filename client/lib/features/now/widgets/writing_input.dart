import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/theme/colors.dart';
import '../providers/now_page_provider.dart';
import '../../starmap/providers/starmap_provider.dart';

class WritingInput extends ConsumerStatefulWidget {
  const WritingInput({super.key});

  @override
  ConsumerState<WritingInput> createState() => _WritingInputState();
}

class _WritingInputState extends ConsumerState<WritingInput> {
  final _controller = TextEditingController();
  bool _canSubmit = false;
  bool _topicConsumed = false;

  @override
  void initState() {
    super.initState();
    _controller.addListener(_onChanged);
  }

  void _onChanged() {
    final can = _controller.text.trim().isNotEmpty;
    if (can != _canSubmit) setState(() => _canSubmit = can);
  }

  void _submit() {
    final text = _controller.text.trim();
    if (text.isEmpty) return;
    ref.read(nowPageProvider.notifier).submitMoment(text);
    _controller.clear();
    ref.read(pendingTopicPromptProvider.notifier).state = null;
    _topicConsumed = false;
  }

  void _cancel() {
    _controller.clear();
    ref.read(pendingTopicPromptProvider.notifier).state = null;
    _topicConsumed = false;
    ref.read(nowPageProvider.notifier).dismissEcho();
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(nowPageProvider);
    final pendingTopic = ref.watch(pendingTopicPromptProvider);
    final show = state.status == NowPageStatus.writing ||
        (state.status == NowPageStatus.echoing && state.isReopen);

    if (pendingTopic != null && !_topicConsumed && !state.isReopen) {
      _topicConsumed = true;
    }

    String hint;
    if (pendingTopic != null && (!state.isReopen || _topicConsumed)) {
      hint = pendingTopic;
    } else if (state.isReopen) {
      hint = '顺着刚才的，再说一句……';
    } else {
      hint = '随便说点什么，这里听着……';
    }
    final tip = state.isReopen ? '光继续听着' : '想说多久说多久，什么时候停都行';

    final screenHeight = MediaQuery.of(context).size.height;

    return AnimatedPositioned(
      duration: const Duration(milliseconds: 350),
      curve: Curves.easeOut,
      left: 24,
      right: 24,
      top: show ? 0 : -screenHeight,
      bottom: show ? 0 : screenHeight,
      child: Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            if (state.isLoading)
              const Padding(
                padding: EdgeInsets.only(bottom: 8),
                child: SizedBox(
                  width: 16,
                  height: 16,
                  child: CircularProgressIndicator(
                    strokeWidth: 1.5,
                    color: AppColors.gold,
                  ),
                ),
              ),
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: Colors.white.withValues(alpha: 0.03),
                borderRadius: BorderRadius.circular(18),
                border: Border.all(color: Colors.white.withValues(alpha: 0.08)),
              ),
              child: TextField(
                controller: _controller,
                enabled: !state.isLoading,
                maxLines: 3,
                style: const TextStyle(
                  color: Color(0xFFE8E8F0),
                  fontSize: 15,
                  height: 1.7,
                  fontWeight: FontWeight.w300,
                ),
                decoration: InputDecoration(
                  hintText: hint,
                  hintStyle: const TextStyle(
                    color: Color(0xFF4A4A60),
                    fontWeight: FontWeight.w200,
                  ),
                  border: InputBorder.none,
                  contentPadding: EdgeInsets.zero,
                ),
              ),
            ),
            const SizedBox(height: 10),
            Text(
              tip,
              style: const TextStyle(
                fontSize: 11,
                color: Color(0xFF5A5A70),
                letterSpacing: 1,
                fontWeight: FontWeight.w200,
              ),
            ),
            const SizedBox(height: 14),
            Row(
              children: [
                Expanded(
                  child: TextButton(
                    onPressed: _cancel,
                    style: TextButton.styleFrom(
                      padding: const EdgeInsets.symmetric(vertical: 11),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(20),
                        side: const BorderSide(color: Color(0x1EFFFFFF)),
                      ),
                      foregroundColor: const Color(0xFFA8A8C0),
                      textStyle: const TextStyle(
                        fontSize: 12,
                        fontWeight: FontWeight.w300,
                        letterSpacing: 1,
                      ),
                    ),
                    child: const Text('算了'),
                  ),
                ),
                const SizedBox(width: 10),
                Expanded(
                  child: TextButton(
                    onPressed: _canSubmit && !state.isLoading ? _submit : null,
                    style: TextButton.styleFrom(
                      padding: const EdgeInsets.symmetric(vertical: 11),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(20),
                        side: BorderSide(
                          color: AppColors.gold.withValues(alpha: 0.4),
                        ),
                      ),
                      foregroundColor: AppColors.warmGold,
                      disabledForegroundColor:
                          AppColors.gold.withValues(alpha: 0.3),
                      textStyle: const TextStyle(
                        fontSize: 12,
                        fontWeight: FontWeight.w300,
                        letterSpacing: 1,
                      ),
                    ),
                    child: const Text('先到这儿'),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
