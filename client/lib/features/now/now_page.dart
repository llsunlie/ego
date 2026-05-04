import 'package:flutter/material.dart';

class NowPage extends StatelessWidget {
  const NowPage({super.key});

  @override
  Widget build(BuildContext context) {
    return const Scaffold(
      body: Center(
        child: Text(
          '有什么想说的吗',
          style: TextStyle(fontSize: 18),
        ),
      ),
    );
  }
}
