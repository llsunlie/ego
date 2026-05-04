import 'package:flutter/material.dart';

class PastPage extends StatelessWidget {
  const PastPage({super.key});

  @override
  Widget build(BuildContext context) {
    return const Scaffold(
      body: Center(
        child: Text(
          '每一次说出口的，都留在这里',
          style: TextStyle(fontSize: 18),
        ),
      ),
    );
  }
}
