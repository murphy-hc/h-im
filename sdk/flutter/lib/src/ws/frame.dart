import 'dart:typed_data';
const int currentVersion = 1;
class FrameType {
  static const int unspecified=0, privateChat=1, privateAck=2, groupChat=3, groupAck=4, chatroomMsg=5;
  static const int chatroomAck=6, heartbeat=7, error=8, sync=9, privateRecall=10;
  static const int chatroomJoin=11, chatroomLeave=12, groupJoin=13, groupLeave=14;
  static String name(int t) => switch(t){1=>'PRIVATE_CHAT',2=>'PRIVATE_ACK',3=>'GROUP_CHAT',4=>'GROUP_ACK',
    5=>'CHATROOM_MSG',6=>'CHATROOM_ACK',7=>'HEARTBEAT',8=>'ERROR',9=>'SYNC',10=>'PRIVATE_RECALL',
    11=>'CHATROOM_JOIN',12=>'CHATROOM_LEAVE',13=>'GROUP_JOIN',14=>'GROUP_LEAVE',_=>'UNKNOWN($t)'};
}
class Frame {
  final int version, frameType; final Uint8List payload;
  Frame({required this.version, required this.frameType, Uint8List? payload}):payload=payload??Uint8List(0);
  Uint8List encode(){final h=ByteData(9);h.setUint8(0,version);h.setUint32(1,frameType,Endian.big);
    h.setUint32(5,payload.length,Endian.big);final r=Uint8List(9+payload.length);
    r.setRange(0,9,h.buffer.asUint8List());if(payload.isNotEmpty)r.setRange(9,9+payload.length,payload);return r;}
  static Frame? decode(Uint8List d){if(d.length<9)return null;final h=ByteData.sublistView(d,0,9);
    final v=h.getUint8(0),ft=h.getUint32(1,Endian.big),pl=h.getUint32(5,Endian.big);
    if(d.length<9+pl)return null;return Frame(version:v,frameType:ft,payload:Uint8List.sublistView(d,9,9+pl));}
  factory Frame.heartbeat()=>Frame(version:currentVersion,frameType:FrameType.heartbeat);
  factory Frame.raw(int ft,String s)=>Frame(version:currentVersion,frameType:ft,payload:Uint8List.fromList(s.codeUnits));
  factory Frame.proto(int ft,List<int> b)=>Frame(version:currentVersion,frameType:ft,payload:Uint8List.fromList(b));
  String get payloadAsString=>String.fromCharCodes(payload);
  String toString()=>'Frame(v$version,${FrameType.name(frameType)},${payload.length}B)';
}
