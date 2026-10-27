import 'package:flutter/material.dart';

import '../../../core/api_client.dart';
import '../../../core/config.dart';

/// Bottom sheet para exibir, testar e salvar a URL do servidor (RNF06).
class ServerUrlSheet {
  ServerUrlSheet._();

  static Future<void> show(BuildContext context) {
    return showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      builder: (context) => const _ServerUrlSheetBody(),
    );
  }
}

class _ServerUrlSheetBody extends StatefulWidget {
  const _ServerUrlSheetBody();

  @override
  State<_ServerUrlSheetBody> createState() => _ServerUrlSheetBodyState();
}

class _ServerUrlSheetBodyState extends State<_ServerUrlSheetBody> {
  late final TextEditingController _controller;
  String? _message;
  bool _messageIsError = false;
  bool _busy = false;

  @override
  void initState() {
    super.initState();
    _controller = TextEditingController(text: AppConfig.apiBaseUrl);
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  void _showMessage(String text, {required bool error}) {
    if (!mounted) return;
    setState(() {
      _message = text;
      _messageIsError = error;
    });
  }

  String? _normalized() {
    final value = AppConfig.normalizeUrl(_controller.text);
    if (value == null) {
      _showMessage(
        'URL inválida. Use o formato http://host ou https://host.',
        error: true,
      );
    }
    return value;
  }

  Future<void> _test() async {
    final url = _normalized();
    if (url == null) return;
    setState(() {
      _busy = true;
      _message = null;
    });
    try {
      final resp = await apiGet('$url/health');
      if (!mounted) return;
      if (resp.statusCode == 200) {
        _showMessage('Servidor respondeu OK', error: false);
      } else {
        _showMessage('Servidor respondeu HTTP ${resp.statusCode}', error: true);
      }
    } catch (e) {
      _showMessage(friendlyError(e), error: true);
    } finally {
      if (mounted) {
        setState(() => _busy = false);
      }
    }
  }

  Future<void> _save() async {
    final url = _normalized();
    if (url == null) return;
    try {
      await AppConfig.save(url);
      if (!mounted) return;
      final messenger = ScaffoldMessenger.of(context);
      Navigator.pop(context);
      messenger.showSnackBar(
        SnackBar(content: Text('Servidor: ${AppConfig.displayHost}')),
      );
    } catch (e) {
      _showMessage('URL inválida. Use o formato http://host ou https://host.',
          error: true);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.only(
        left: 16,
        right: 16,
        top: 16,
        bottom: MediaQuery.of(context).viewInsets.bottom + 16,
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          const Text(
            'Servidor',
            style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 4),
          Text(
            'Atual: ${AppConfig.displayHost}',
            style: const TextStyle(fontSize: 13, color: Colors.grey),
          ),
          const SizedBox(height: 16),
          TextField(
            controller: _controller,
            keyboardType: TextInputType.url,
            decoration: const InputDecoration(
              labelText: 'URL do servidor',
              hintText: 'https://seu-tunnel.ngrok-free.dev',
              border: OutlineInputBorder(),
            ),
          ),
          if (_message != null) ...[
            const SizedBox(height: 12),
            Text(
              _message!,
              style: TextStyle(
                color: _messageIsError ? Colors.red : Colors.green,
                fontSize: 13,
              ),
            ),
          ],
          const SizedBox(height: 16),
          Row(
            children: [
              Expanded(
                child: OutlinedButton(
                  onPressed: _busy ? null : _test,
                  child: _busy
                      ? const SizedBox(
                          width: 18,
                          height: 18,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : const Text('Testar'),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: ElevatedButton(
                  onPressed: _busy ? null : _save,
                  child: const Text('Salvar'),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}
