import 'dart:async';

import 'package:flutter/material.dart';

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
    '刚才吃了什么？',
  ];

  static const _typeSpeed = Duration(milliseconds: 120);
  static const _deleteSpeed = Duration(milliseconds: 60);
  static const _pauseDuration = Duration(milliseconds: 2500);
  static const _switchDelay = Duration(milliseconds: 400);
  static const _initialDelay = Duration(seconds: 1);

  int _textIndex = 0;
  int _charCount = 0;
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
      setState(() {});
      _typeTimer = Timer(_pauseDuration, _startDeleting);
    }
  }

  void _startDeleting() {
    if (!mounted) return;
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
      child: Text.rich(
        TextSpan(
          children: [
            TextSpan(text: displayed),
            TextSpan(
              text: '|',
              style: TextStyle(
                color: _cursorVisible
                    ? const Color(0xFFA8A8C0)
                    : const Color(0x00A8A8C0),
              ),
            ),
          ],
        ),
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

class WriteButton extends StatelessWidget {
  final VoidCallback onTap;

  const WriteButton({super.key, required this.onTap});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 60),
      child: Center(
        child: FractionallySizedBox(
          widthFactor: 0.65,
          child: ElevatedButton(
            onPressed: onTap,
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.white.withValues(alpha: 0.05),
              foregroundColor: const Color(0xFFD0D0E0),
              padding: const EdgeInsets.symmetric(horizontal: 36, vertical: 13),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(26),
                side: const BorderSide(color: Color(0x1EFFFFFF)),
              ),
              textStyle: const TextStyle(
                fontSize: 13,
                fontWeight: FontWeight.w300,
                letterSpacing: 2,
              ),
            ),
            child: const Text('写下此刻'),
          ),
        ),
      ),
    );
  }
}
