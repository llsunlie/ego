# Now Page Guide Text Typewriter — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the static guide text "有什么想说的吗" on the now page with a 7-text typewriter carousel that types out, pauses, deletes, and cycles through prompts.

**Architecture:** Convert `GuideText` from `StatelessWidget` to `StatefulWidget` with a Timer-driven state machine (typing → pause → deleting → switch → loop). No provider or parent page changes needed — the widget is self-contained and only mounted during idle state.

**Tech Stack:** Flutter, dart:async (Timer)

---

## File Structure

| File | Action | Responsibility |
|------|--------|----------------|
| `client/lib/features/now/widgets/guide_section.dart` | Modify | `GuideText` widget — typewriter state machine + rendering |
| `client/lib/features/now/now_page.dart` | No change | Already guards `GuideText` behind `if (isIdle)` |

---

### Task 1: Implement typewriter GuideText widget

**Files:**
- Modify: `client/lib/features/now/widgets/guide_section.dart`

- [ ] **Step 1: Add dart:async import**

At the top of `guide_section.dart`, add:

```dart
import 'dart:async';
```

- [ ] **Step 2: Replace GuideText class with StatefulWidget + typewriter logic**

Replace the existing `GuideText` class:

```dart
class GuideText extends StatelessWidget {
  const GuideText({super.key});

  @override
  Widget build(BuildContext context) {
    return FractionallySizedBox(
      widthFactor: 0.5,
      child: const Text(
        '有什么想说的吗',
        textAlign: TextAlign.center,
        style: TextStyle(
          fontSize: 14,
          color: Color(0xFFA8A8C0),
          letterSpacing: 5,
          fontWeight: FontWeight.w200,
        ),
      ),
    );
  }
}
```

With:

```dart
enum _Phase { typing, pause, deleting, switching }

class GuideText extends StatefulWidget {
  const GuideText({super.key});

  @override
  State<GuideText> createState() => _GuideTextState();
}

class _GuideTextState extends State<GuideText> {
  static const _texts = [
    '有什么想说的吗？',
    '刚才做了什么？',
    '今天有什么新鲜事？',
    '此刻在想什么？',
    '心情怎么样？',
    '有什么想记住的？',
    '随便写点什么吧',
  ];

  static const _typeSpeed = Duration(milliseconds: 120);
  static const _deleteSpeed = Duration(milliseconds: 60);
  static const _pauseDuration = Duration(milliseconds: 2500);
  static const _switchDelay = Duration(milliseconds: 400);
  static const _initialDelay = Duration(seconds: 1);

  int _textIndex = 0;
  int _charCount = 0;
  _Phase _phase = _Phase.typing;
  bool _cursorVisible = true;
  Timer? _typeTimer;
  Timer? _cursorTimer;

  @override
  void initState() {
    super.initState();
    // Random start text to avoid always showing #1
    _textIndex = DateTime.now().millisecondsSinceEpoch % _texts.length;
    // Cursor blink timer
    _cursorTimer = Timer.periodic(
      const Duration(milliseconds: 500),
      (_) {
        if (mounted) setState(() => _cursorVisible = !_cursorVisible);
      },
    );
    // Initial delay before first type
    Future.delayed(_initialDelay, () {
      if (mounted) _startTyping();
    });
  }

  @override
  void dispose() {
    _typeTimer?.cancel();
    _cursorTimer?.cancel();
    super.dispose();
  }

  void _startTyping() {
    _phase = _Phase.typing;
    _charCount = 0;
    setState(() {});
    _typeNextChar();
  }

  void _typeNextChar() {
    if (!mounted) return;
    final text = _texts[_textIndex];
    if (_charCount < text.length) {
      _charCount++;
      setState(() {});
      _typeTimer = Timer(_typeSpeed, _typeNextChar);
    } else {
      _phase = _Phase.pause;
      setState(() {});
      _typeTimer = Timer(_pauseDuration, _startDeleting);
    }
  }

  void _startDeleting() {
    if (!mounted) return;
    _phase = _Phase.deleting;
    setState(() {});
    _deleteNextChar();
  }

  void _deleteNextChar() {
    if (!mounted) return;
    if (_charCount > 0) {
      _charCount--;
      setState(() {});
      _typeTimer = Timer(_deleteSpeed, _deleteNextChar);
    } else {
      _phase = _Phase.switching;
      _textIndex = (_textIndex + 1) % _texts.length;
      setState(() {});
      _typeTimer = Timer(_switchDelay, _startTyping);
    }
  }

  @override
  Widget build(BuildContext context) {
    final text = _texts[_textIndex];
    final displayed = text.substring(0, _charCount);

    return FractionallySizedBox(
      widthFactor: 0.5,
      child: Text(
        '$displayed${_cursorVisible ? '|' : ' '}',
        textAlign: TextAlign.center,
        style: const TextStyle(
          fontSize: 14,
          color: Color(0xFFA8A8C0),
          letterSpacing: 5,
          fontWeight: FontWeight.w200,
        ),
      ),
    );
  }
}
```

- [ ] **Step 3: Verify Flutter static analysis passes**

```bash
cd client && flutter analyze
```

Expected: zero issues.

---

## Verification Checklist

| # | Check | Command |
|---|-------|---------|
| 1 | Flutter analyze | `cd client && flutter analyze` |
| 2 | Widget mounts in idle only | Existing `if (isIdle)` guard unchanged |
| 3 | Timer cancels on dispose | `dispose()` calls `_typeTimer?.cancel()` and `_cursorTimer?.cancel()` |
| 4 | Mounted check before setState | All timer callbacks guard with `if (mounted)` |
