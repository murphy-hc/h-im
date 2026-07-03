import 'package:flutter/material.dart';
import 'package:him_flutter/him_flutter.dart';

class ChatBubble extends StatelessWidget {
  final Message message; final bool isMe;
  final void Function()? onTap; final void Function()? onLongPress;
  const ChatBubble({super.key, required this.message, required this.isMe, this.onTap, this.onLongPress});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final color = isMe ? theme.colorScheme.primaryContainer : theme.colorScheme.surfaceContainerHighest;
    final align = isMe ? CrossAxisAlignment.end : CrossAxisAlignment.start;
    return GestureDetector(
      onTap: onTap, onLongPress: onLongPress,
      child: Container(
        margin: const EdgeInsets.symmetric(vertical: 4, horizontal: 12),
        child: Column(crossAxisAlignment: align, children: [
          if (!isMe && message.senderId.isNotEmpty)
            Padding(padding: const EdgeInsets.only(bottom: 2),
              child: Text(message.senderId, style: theme.textTheme.labelSmall?.copyWith(color: theme.colorScheme.onSurfaceVariant))),
          Container(
            constraints: BoxConstraints(maxWidth: MediaQuery.of(context).size.width * 0.75),
            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
            decoration: BoxDecoration(color: color, borderRadius: BorderRadius.only(
              topLeft: const Radius.circular(16), topRight: const Radius.circular(16),
              bottomLeft: Radius.circular(isMe ? 16 : 4), bottomRight: Radius.circular(isMe ? 4 : 16))),
            child: Column(crossAxisAlignment: align, children: [
              if (message.attachment?.image != null)
                ClipRRect(borderRadius: BorderRadius.circular(8),
                  child: Image.network(message.attachment!.image!.url, width: 200, height: 200, fit: BoxFit.cover)),
              if (message.text.isNotEmpty) SelectableText(message.text, style: theme.textTheme.bodyMedium),
              const SizedBox(height: 4),
              Row(mainAxisSize: MainAxisSize.min, children: [
                Text(_fmt(message.serverTime > 0 ? message.serverTime : message.createTime),
                  style: theme.textTheme.labelSmall?.copyWith(color: theme.colorScheme.onSurfaceVariant)),
                if (isMe) ...[const SizedBox(width: 4), _icon(message.status)],
              ]),
            ]),
          ),
        ]),
      ),
    );
  }

  Widget _icon(MsgStatus s) {
    final icons = {MsgStatus.sending: Icons.access_time, MsgStatus.sent: Icons.check,
      MsgStatus.delivered: Icons.done_all, MsgStatus.read: Icons.done_all,
      MsgStatus.recalled: Icons.undo, MsgStatus.failed: Icons.error_outline};
    final colors = {MsgStatus.read: Colors.blue, MsgStatus.failed: Colors.red};
    return Icon(icons[s] ?? Icons.check, size: 14, color: colors[s] ?? Colors.grey);
  }

  String _fmt(int ms) {
    final d = DateTime.fromMillisecondsSinceEpoch(ms);
    return '${d.hour.toString().padLeft(2,'0')}:${d.minute.toString().padLeft(2,'0')}';
  }
}
