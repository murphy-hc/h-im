import 'dart:typed_data';
import 'package:test/test.dart';
import 'package:him_flutter/src/ws/frame.dart';

void main() {
  group('Frame', () {
    test('heartbeat round-trip', () {
      final f = Frame.heartbeat(); final d = Frame.decode(f.encode());
      expect(d, isNotNull); expect(d!.version, 1); expect(d.frameType, FrameType.heartbeat);
    });
    test('raw string', () {
      final d = Frame.decode(Frame.raw(FrameType.chatroomJoin, 'room-1').encode());
      expect(d!.payloadAsString, 'room-1');
    });
    test('short frame null', () => expect(Frame.decode(Uint8List(5)), isNull));
    test('length mismatch null', () {
      final h = ByteData(9); h.setUint8(0, 1); h.setUint32(5, 100, Endian.big);
      expect(Frame.decode(h.buffer.asUint8List()), isNull);
    });
  });
  group('FrameType', () {
    test('names', () { expect(FrameType.name(7), 'HEARTBEAT'); expect(FrameType.name(99), contains('UNKNOWN')); });
  });
}
