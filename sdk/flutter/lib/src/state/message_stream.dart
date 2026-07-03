import 'dart:async';
import 'package:him_flutter/him_flutter.dart';

/// Events emitted by the message stream.
class MessageEvent {
  final Message message;
  final String event;

  const MessageEvent({required this.message, this.event = 'new'});

  static MessageEvent newMessage(Message msg) => MessageEvent(message: msg);
  static MessageEvent recall(int serverId) => MessageEvent(
    message: Message(senderId: '', receiverId: '', convType: ConvType.private),
    event: 'recall:$serverId',
  );
  static MessageEvent statusUpdate(String clientId, MsgStatus status) => MessageEvent(
    message: Message(senderId: '', receiverId: '', convType: ConvType.private, clientId: clientId, status: status),
    event: 'status:$clientId',
  );
}

/// A broadcast stream of message events from the SDK.
///
/// Usage:
/// ```dart
/// final stream = MessageStream(client);
/// stream.events.listen((event) {
///   print('${event.event}: ${event.message.text}');
/// });
/// ```
class MessageStream {
  final HimClient _client;
  final StreamController<MessageEvent> _controller = StreamController<MessageEvent>.broadcast();

  MessageStream(this._client) {
    _client.callbacks.onFrame = (frame) {
      _handleFrame(frame);
      _client.callbacks.onFrame?.call(frame);
    };
  }

  /// Broadcast stream of all message events.
  Stream<MessageEvent> get events => _controller.stream;

  /// Stream of private messages only.
  Stream<Message> get privateMessages =>
    events.where((e) => e.event == 'new' && e.message.convType == ConvType.private).map((e) => e.message);

  /// Stream of group messages only.
  Stream<Message> get groupMessages =>
    events.where((e) => e.event == 'new' && e.message.convType == ConvType.group).map((e) => e.message);

  void _handleFrame(Frame frame) {
    switch (frame.frameType) {
      case FrameType.privateChat:
        _controller.add(MessageEvent.newMessage(_parseIncoming(frame, ConvType.private)));
      case FrameType.groupChat:
        _controller.add(MessageEvent.newMessage(_parseIncoming(frame, ConvType.group)));
      case FrameType.chatroomMsg:
        _controller.add(MessageEvent.newMessage(_parseIncoming(frame, ConvType.room)));
      case FrameType.privateRecall:
      case FrameType.error:
        _controller.add(MessageEvent(message: Message(
          senderId: '', receiverId: '', convType: ConvType.private,
          text: frame.payloadAsString,
        ), event: 'error'));
    }
  }

  Message _parseIncoming(Frame frame, ConvType type) {
    // Simple text extraction from proto payload
    final payload = frame.payload;
    String text = '';
    String senderId = '';
    String receiverId = '';

    var i = 0;
    while (i < payload.length - 1) {
      final tag = _readVarint(payload, i);
      i = tag.offset;
      final fieldNum = tag.value >> 3;
      final wireType = tag.value & 0x7;

      if (wireType == 0) {
        while (i < payload.length && (payload[i] & 0x80) != 0) i++;
        i++;
      } else if (wireType == 2) {
        final len = _readVarint(payload, i);
        i = len.offset;
        final str = String.fromCharCodes(payload.sublist(i, i + len.value));
        if (fieldNum == 3) senderId = str;
        if (fieldNum == 4) receiverId = str;
        if (fieldNum == 8) text = str;
        i += len.value;
      } else {
        break;
      }
    }

    return Message(
      senderId: senderId,
      receiverId: receiverId,
      convType: type,
      text: text,
    );
  }

  _VarintResult _readVarint(List<int> data, int offset) {
    var value = 0;
    var shift = 0;
    var i = offset;
    while (i < data.length) {
      final b = data[i++];
      value |= (b & 0x7F) << shift;
      if ((b & 0x80) == 0) break;
      shift += 7;
    }
    return _VarintResult(value, i);
  }

  void dispose() {
    _controller.close();
  }
}

class _VarintResult {
  final int value;
  final int offset;
  _VarintResult(this.value, this.offset);
}
