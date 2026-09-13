import 'package:flutter/material.dart';

class PttButton extends StatefulWidget {
  final bool isTransmitting;
  final bool isChannelBusy;
  final VoidCallback? onPressed;
  final VoidCallback? onReleased;

  const PttButton({
    super.key,
    required this.isTransmitting,
    required this.isChannelBusy,
    this.onPressed,
    this.onReleased,
  });

  @override
  State<PttButton> createState() => _PttButtonState();
}

class _PttButtonState extends State<PttButton>
    with SingleTickerProviderStateMixin {
  late AnimationController _controller;
  late Animation<double> _scaleAnimation;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      duration: const Duration(milliseconds: 150),
      vsync: this,
    );
    _scaleAnimation = Tween<double>(begin: 1.0, end: 0.95).animate(
      CurvedAnimation(parent: _controller, curve: Curves.easeInOut),
    );
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final bool canTransmit = !widget.isChannelBusy || widget.isTransmitting;
    final Color buttonColor = widget.isTransmitting
        ? Colors.red
        : canTransmit
            ? Colors.green
            : Colors.grey;

    return GestureDetector(
      onTapDown: canTransmit ? (_) => _handlePress() : null,
      onTapUp: canTransmit ? (_) => _handleRelease() : null,
      onTapCancel: canTransmit ? _handleRelease : null,
      child: AnimatedBuilder(
        animation: _scaleAnimation,
        builder: (context, child) {
          return Transform.scale(
            scale: _scaleAnimation.value,
            child: child,
          );
        },
        child: Container(
          width: 120,
          height: 120,
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            color: buttonColor,
            boxShadow: [
              BoxShadow(
                color: buttonColor.withOpacity(0.5),
                blurRadius: 20,
                spreadRadius: 5,
              ),
            ],
          ),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(
                widget.isTransmitting ? Icons.mic : Icons.mic_none,
                color: Colors.white,
                size: 40,
              ),
              const SizedBox(height: 4),
              Text(
                widget.isTransmitting ? 'FALANDO' : 'PTT',
                style: const TextStyle(
                  color: Colors.white,
                  fontWeight: FontWeight.bold,
                  fontSize: 12,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  void _handlePress() {
    _controller.forward();
    widget.onPressed?.call();
  }

  void _handleRelease() {
    _controller.reverse();
    widget.onReleased?.call();
  }
}
