import 'package:flutter/material.dart';

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
