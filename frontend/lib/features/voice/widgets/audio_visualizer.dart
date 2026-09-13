import 'dart:math';
import 'package:flutter/material.dart';

class AudioVisualizer extends StatefulWidget {
  final bool isActive;
  final Color color;
  final double height;
  final int barCount;

  const AudioVisualizer({
    super.key,
    required this.isActive,
    this.color = Colors.green,
    this.height = 60,
    this.barCount = 5,
  });

  @override
  State<AudioVisualizer> createState() => _AudioVisualizerState();
}

class _AudioVisualizerState extends State<AudioVisualizer>
    with TickerProviderStateMixin {
  late AnimationController _controller;
  late List<double> _barHeights;
  final Random _random = Random();

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      duration: const Duration(milliseconds: 150),
      vsync: this,
    );
    _barHeights = List.generate(widget.barCount, (_) => 0.1);

    if (widget.isActive) {
      _startAnimation();
    }
  }

  @override
  void didUpdateWidget(AudioVisualizer oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.isActive && !oldWidget.isActive) {
      _startAnimation();
    } else if (!widget.isActive && oldWidget.isActive) {
      _stopAnimation();
    }
  }

  void _startAnimation() {
    _controller.repeat();
    _controller.addListener(_updateBars);
  }

  void _stopAnimation() {
    _controller.removeListener(_updateBars);
    _controller.stop();
    setState(() {
      _barHeights = List.generate(widget.barCount, (_) => 0.1);
    });
  }

  void _updateBars() {
    setState(() {
      for (int i = 0; i < widget.barCount; i++) {
        _barHeights[i] = 0.1 + _random.nextDouble() * 0.9;
      }
    });
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: widget.height,
      child: Row(
        mainAxisAlignment: MainAxisAlignment.center,
        crossAxisAlignment: CrossAxisAlignment.end,
        children: List.generate(widget.barCount, (index) {
          return AnimatedContainer(
            duration: const Duration(milliseconds: 100),
            width: 8,
            height: widget.height * _barHeights[index],
            margin: const EdgeInsets.symmetric(horizontal: 2),
            decoration: BoxDecoration(
              color: widget.isActive
                  ? widget.color.withOpacity(0.5 + _barHeights[index] * 0.5)
                  : widget.color.withOpacity(0.2),
              borderRadius: BorderRadius.circular(4),
            ),
          );
        }),
      ),
    );
  }
}
