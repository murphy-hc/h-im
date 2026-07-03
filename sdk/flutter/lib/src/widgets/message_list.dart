import 'package:flutter/material.dart';
import 'package:him_flutter/him_flutter.dart';
import '../state/message_stream.dart';
import 'chat_bubble.dart';

class MessageList extends StatefulWidget {
  final MessageStream stream; final String currentUserId;
  final ScrollController? scrollController;
  final void Function(Message m)? onTapMessage, onLongPressMessage;
  const MessageList({super.key, required this.stream, required this.currentUserId,
    this.scrollController, this.onTapMessage, this.onLongPressMessage});

  @override
  State<MessageList> createState() => _MessageListState();
}

class _MessageListState extends State<MessageList> {
  final List<Message> _msgs = [];
  ScrollController? _sc;

  @override
  void initState() {
    super.initState(); _sc = widget.scrollController ?? ScrollController();
    widget.stream.events.listen((e) => setState(() {
      if (e.event == 'new') { _msgs.add(e.message); }
      else if (e.event.startsWith('recall:')) {
        final sid = int.tryParse(e.event.split(':').last);
        if (sid != null) {
          final i = _msgs.indexWhere((m) => m.serverId == sid);
          if (i >= 0) _msgs[i] = _msgs[i].copyWith(isDeleted: true, status: MsgStatus.recalled);
        }
      } else if (e.event.startsWith('status:')) {
        final cid = e.event.substring(7);
        final i = _msgs.indexWhere((m) => m.clientId == cid);
        if (i >= 0) _msgs[i] = _msgs[i].copyWith(status: e.message.status);
      }
      _scroll();
    }));
  }

  void _scroll() {
    if (_sc!.hasClients) WidgetsBinding.instance.addPostFrameCallback((_) {
      _sc!.animateTo(_sc!.position.maxScrollExtent, duration: const Duration(milliseconds: 300), curve: Curves.easeOut);
    });
  }

  @override
  void dispose() { if (widget.scrollController == null) _sc?.dispose(); super.dispose(); }

  @override
  Widget build(BuildContext ctx) => ListView.builder(controller: _sc, itemCount: _msgs.length,
    padding: const EdgeInsets.only(bottom: 8), itemBuilder: (ctx, i) {
      final m = _msgs[i];
      if (m.isDeleted) return Center(child: Text('Message recalled',
        style: Theme.of(ctx).textTheme.labelSmall?.copyWith(color: Colors.grey)));
      return ChatBubble(message: m, isMe: m.senderId == widget.currentUserId,
        onTap: () => widget.onTapMessage?.call(m), onLongPress: () => widget.onLongPressMessage?.call(m));
    });
}
