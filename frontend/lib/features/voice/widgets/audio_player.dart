import 'package:flutter/material.dart';

class AudioPlayerWidget extends StatefulWidget {
  final bool isPlaying;
  final String? currentSpeaker;
  final VoidCallback? onStop;

  const AudioPlayerWidget({
    super.key,
    required this.isPlaying,
    this.currentSpeaker,
    this.onStop,
  });

  @override
  State<AudioPlayerWidget> createState() => _AudioPlayerWidgetState();
}

class _AudioPlayerWidgetState extends State<AudioPlayerWidget>
    with SingleTickerProviderStateMixin {
  late AnimationController _controller;
  late Animation<double> _pulseAnimation;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      duration: const Duration(milliseconds: 1000),
      vsync: this,
    );
    _pulseAnimation = Tween<double>(begin: 1.0, end: 1.2).animate(
      CurvedAnimation(parent: _controller, curve: Curves.easeInOut),
    );

    if (widget.isPlaying) {
      _controller.repeat(reverse: true);
    }
  }

  @override
  void didUpdateWidget(AudioPlayerWidget oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.isPlaying && !oldWidget.isPlaying) {
      _controller.repeat(reverse: true);
    } else if (!widget.isPlaying && oldWidget.isPlaying) {
      _controller.stop();
      _controller.reset();
    }
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: widget.isPlaying ? Colors.blue.withOpacity(0.1) : Colors.grey.withOpacity(0.1),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: widget.isPlaying ? Colors.blue : Colors.grey,
          width: 1,
        ),
      ),
      child: Row(
        children: [
          AnimatedBuilder(
            animation: _pulseAnimation,
            builder: (context, child) {
              return Transform.scale(
                scale: widget.isPlaying ? _pulseAnimation.value : 1.0,
                child: Icon(
                  Icons.volume_up,
                  color: widget.isPlaying ? Colors.blue : Colors.grey,
                  size: 32,
                ),
              );
            },
          ),
          const SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  widget.isPlaying ? 'Recebendo áudio' : 'Aguardando áudio',
                  style: TextStyle(
                    fontWeight: FontWeight.bold,
                    color: widget.isPlaying ? Colors.blue : Colors.grey,
                  ),
                ),
                if (widget.currentSpeaker != null)
                  Text(
                    'Falante: ${widget.currentSpeaker}',
                    style: TextStyle(
                      color: Colors.grey[600],
                      fontSize: 12,
                    ),
                  ),
              ],
            ),
          ),
          if (widget.isPlaying && widget.onStop != null)
            IconButton(
              icon: const Icon(Icons.stop),
              color: Colors.red,
              onPressed: widget.onStop,
            ),
        ],
      ),
    );
  }
}
