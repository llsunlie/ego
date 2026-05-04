import 'package:flutter/material.dart';

class StarmapPage extends StatelessWidget {
  const StarmapPage({super.key});

  @override
  Widget build(BuildContext context) {
    return const Scaffold(
      body: Center(
        child: Text(
          '已有 0 颗星',
          style: TextStyle(fontSize: 18),
        ),
      ),
    );
  }
}
