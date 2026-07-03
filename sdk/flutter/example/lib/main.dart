import 'package:flutter/material.dart';
import 'package:him_flutter/him_flutter.dart';

void main() {
  runApp(const HimExampleApp());
}

class HimExampleApp extends StatelessWidget {
  const HimExampleApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'h-im Example',
      theme: ThemeData(
        colorSchemeSeed: Colors.indigo,
        useMaterial3: true,
      ),
      home: const ChatPage(),
    );
  }
}

class ChatPage extends StatefulWidget {
  const ChatPage({super.key});

  @override
  State<ChatPage> createState() => _ChatPageState();
}

class _ChatPageState extends State<ChatPage> {
  final _textController = TextEditingController();
  final _client = HimClient(
    config: const HimConfig(host: 'localhost', port: 8080, appId: 'example'),
    callbacks: HimCallbacks(
      onConnectionChange: (state) => debugPrint('Connection: $state'),
      onError: (e) => debugPrint('Error: $e'),
    ),
  );
  late final MessageStream _stream;

  @override
  void initState() {
    super.initState();
    _stream = MessageStream(_client);
    _connect();
  }

  Future<void> _connect() async {
    try {
      await _client.connect(userId: 'demo-user', token: 'demo-token');
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Connection failed: $e')));
      }
    }
  }

  void _sendMessage() {
    final text = _textController.text.trim();
    if (text.isEmpty) return;
    _client.sendPrivateMessage('peer-id', text);
    _textController.clear();
  }

  @override
  void dispose() {
    _textController.dispose();
    _stream.dispose();
    _client.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Chat'), actions: [
        IconButton(
          icon: const Icon(Icons.push_pin),
          tooltip: 'Register Push',
          onPressed: () => PushSetup.requestPermission(),
        ),
      ]),
      body: Column(
        children: [
          Expanded(
            child: MessageList(stream: _stream, currentUserId: 'demo-user'),
          ),
          _buildInputBar(),
        ],
      ),
    );
  }

  Widget _buildInputBar() {
    return Container(
      padding: const EdgeInsets.all(8),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surfaceContainerLow,
        border: Border(top: BorderSide(color: Theme.of(context).dividerColor)),
      ),
      child: SafeArea(
        child: Row(
          children: [
            IconButton(icon: const Icon(Icons.add_circle_outline), onPressed: () {}),
            Expanded(
              child: TextField(
                controller: _textController,
                decoration: const InputDecoration(
                  hintText: 'Type a message...',
                  border: OutlineInputBorder(borderRadius: BorderRadius.all(Radius.circular(24))),
                  contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                  isDense: true,
                ),
                textInputAction: TextInputAction.send,
                onSubmitted: (_) => _sendMessage(),
              ),
            ),
            const SizedBox(width: 4),
            IconButton.filled(icon: const Icon(Icons.send), onPressed: _sendMessage),
          ],
        ),
      ),
    );
  }
}
