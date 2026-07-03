import 'message.dart';
class Conversation{final String convId,peerId;final ConvType type;final String lastMsg;final int lastTime,unreadCount;
  const Conversation({required this.convId,required this.peerId,required this.type,this.lastMsg='',this.lastTime=0,this.unreadCount=0});
  factory Conversation.fromMessage(Message m)=>Conversation(
    convId:m.convType==ConvType.private?m.senderId:m.receiverId,
    peerId:m.convType==ConvType.private?m.senderId:m.receiverId,
    type:m.convType,lastMsg:m.text,lastTime:m.serverTime,unreadCount:1);
  Map<String,dynamic>toJson()=>{'conv_id':convId,'peer_id':peerId,'type':type.value,'last_msg':lastMsg,'last_time':lastTime,'unread_count':unreadCount};
  factory Conversation.fromJson(Map<String,dynamic>j)=>Conversation(convId:j['conv_id'],peerId:j['peer_id'],
    type:ConvType.fromValue(j['type']),lastMsg:j['last_msg']??'',lastTime:j['last_time']??0,unreadCount:j['unread_count']??0);}
