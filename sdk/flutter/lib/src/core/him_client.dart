import 'dart:typed_data';
import '../ws/frame.dart';import '../ws/heartbeat.dart';import 'connection.dart';import 'session.dart';

class HimConfig{final String host,appId,deviceId;final int port;
  const HimConfig({required this.host,this.port=8080,required this.appId,this.deviceId='default'});}

class HimCallbacks{void Function(Frame)?onFrame;final void Function(ConnectionState)?onConnectionChange;
  final void Function(Object)?onError;HimCallbacks({this.onFrame,this.onConnectionChange,this.onError});}

class HimClient{
  final HimConfig config;HimCallbacks callbacks;final Session session=Session();
  late ConnectionManager _conn;late HeartbeatManager _hb;static int _idSeq=0;

  HimClient({required this.config,required this.callbacks}){
    _conn=ConnectionManager(host:config.host,port:config.port,appId:config.appId,userId:'',token:'',
      deviceId:config.deviceId,onFrame:_onFrame,onStateChange:callbacks.onConnectionChange,onError:callbacks.onError);
    _hb=HeartbeatManager(send:_conn.send,onTimeout:(){_conn.disconnect();_conn.connect();});}
  ConnectionState get connectionState=>_conn.state;

  Future<void> connect({required String userId,required String token})async{
    session.onLogin(accessToken:token,refreshToken:token,expiresAt:DateTime.now().add(const Duration(days:7)),userId:userId);
    await _conn.disconnect();
    _conn=ConnectionManager(host:config.host,port:config.port,appId:config.appId,userId:userId,token:token,
      deviceId:config.deviceId,onFrame:_onFrame,onStateChange:callbacks.onConnectionChange,onError:callbacks.onError);
    _hb=HeartbeatManager(send:_conn.send,onTimeout:(){_conn.disconnect();_conn.connect();});
    await _conn.connect();_hb.start();}

  void sendFrame(Frame f)=>_conn.send(f);
  String generateClientId()=>'${DateTime.now().millisecondsSinceEpoch}_${_idSeq++}';
  void sendPrivateMessage(String rid,String text,{String?cid,List<int>?att})=>
    _conn.send(Frame(version:currentVersion,frameType:FrameType.privateChat,payload:_buildMsg(rid,text,0,cid,att)));
  void sendGroupMessage(String gid,String text,{String?cid})=>
    _conn.send(Frame(version:currentVersion,frameType:FrameType.groupChat,payload:_buildMsg(gid,text,1,cid,null)));
  void sendChatroomMessage(String rid,String text,{String?cid})=>
    _conn.send(Frame(version:currentVersion,frameType:FrameType.chatroomMsg,payload:_buildMsg(rid,text,2,cid,null)));
  void recallMessage(int sid)=>
    _conn.send(Frame(version:currentVersion,frameType:FrameType.privateRecall,payload:_encodeRecall(sid)));
  void joinChatroom(String id)=>_conn.send(Frame.raw(FrameType.chatroomJoin,id));
  void leaveChatroom(String id)=>_conn.send(Frame.raw(FrameType.chatroomLeave,id));
  void joinGroup(String id)=>_conn.send(Frame.raw(FrameType.groupJoin,id));
  void leaveGroup(String id)=>_conn.send(Frame.raw(FrameType.groupLeave,id));
  Future<void> dispose()async{_hb.stop();await _conn.disconnect();session.clear();}

  void _onFrame(Frame f){if(f.frameType==FrameType.heartbeat){_hb.onEcho();}
    else if(f.frameType==FrameType.error){callbacks.onError?.call('Server error: ${f.payloadAsString}');}
    else{callbacks.onFrame?.call(f);}}

  static void _wt(BytesBuilder b,int fn,int wt)=>b.add(_ev((fn<<3)|wt));
  static Uint8List _ev(int v){final o=<int>[];var x=v;while(x>0x7F){o.add((x&0x7F)|0x80);x>>=7;}o.add(x&0x7F);return Uint8List.fromList(o);}
  static Uint8List _es(String s){final b=Uint8List.fromList(s.codeUnits);return Uint8List.fromList([..._ev(b.length),...b]);}
  Uint8List _buildMsg(String rid,String text,int ct,String?cid,List<int>?att){final b=BytesBuilder();
    final c=cid??generateClientId();final now=DateTime.now().millisecondsSinceEpoch;
    _wt(b,1,2);b.add(_es(c));_wt(b,3,2);b.add(_es(session.userId??''));
    _wt(b,4,2);b.add(_es(rid));_wt(b,5,0);b.add(_ev(ct));_wt(b,6,0);b.add(_ev(0));_wt(b,8,2);b.add(_es(text));
    if(att!=null){_wt(b,9,2);b.add(Uint8List.fromList([..._ev(att.length),...att]));}
    _wt(b,11,0);b.add(_ev(now));_wt(b,12,0);b.add(_ev(now));return b.takeBytes();}
  Uint8List _encodeRecall(int sid){final b=BytesBuilder();_wt(b,1,0);b.add(_ev(sid));_wt(b,2,2);b.add(_es(session.userId??''));return b.takeBytes();}
}
