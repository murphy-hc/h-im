enum ConvType{private(0),group(1),room(2);final int value;const ConvType(this.value);
  static ConvType fromValue(int v)=>ConvType.values.firstWhere((e)=>e.value==v,orElse:()=>ConvType.private);}
enum MsgType{text(0),image(1),audio(2),video(3),file(4),location(5),notification(6),tips(7),custom(8);final int value;const MsgType(this.value);}
enum MsgStatus{sending(0),sent(1),delivered(2),read(3),recalled(4),failed(5);final int value;const MsgStatus(this.value);}

class Message{
  final String?clientId;final int?serverId;final String senderId,receiverId;final ConvType convType;
  final MsgType msgType;final String text;final Attachment?attachment;final MsgStatus status;
  final int createTime,serverTime;final ThreadReply?threadReply;final bool isDeleted,isRemoteRead;
  const Message({this.clientId,this.serverId,required this.senderId,required this.receiverId,
    required this.convType,this.msgType=MsgType.text,this.text='',this.attachment,this.status=MsgStatus.sending,
    this.createTime=0,this.serverTime=0,this.threadReply,this.isDeleted=false,this.isRemoteRead=false});
  Message copyWith({int?serverId,MsgStatus?status,bool?isDeleted,bool?isRemoteRead})=>Message(
    clientId:clientId,serverId:serverId??this.serverId,senderId:senderId,receiverId:receiverId,
    convType:convType,msgType:msgType,text:text,attachment:attachment,status:status??this.status,
    createTime:createTime,serverTime:serverTime,threadReply:threadReply,
    isDeleted:isDeleted??this.isDeleted,isRemoteRead:isRemoteRead??this.isRemoteRead);
  Map<String,dynamic>toJson()=>{'client_id':clientId,'server_id':serverId,'sender_id':senderId,
    'receiver_id':receiverId,'conv_type':convType.value,'msg_type':msgType.value,'text':text,
    'status':status.value,'create_time':createTime,'server_time':serverTime};
  factory Message.fromJson(Map<String,dynamic>j)=>Message(clientId:j['client_id'],serverId:j['server_id'],
    senderId:j['sender_id']??'',receiverId:j['receiver_id']??'',convType:ConvType.fromValue(j['conv_type']??0),
    msgType:MsgType.values[j['msg_type']??0],text:j['text']??'',status:MsgStatus.values[j['status']??0],
    createTime:j['create_time']??0,serverTime:j['server_time']??0);
  String toString()=>'Message(cid=$clientId,sid=$serverId,$text)';
}

class Attachment{final ImageAttachment?image;final AudioAttachment?audio;final VideoAttachment?video;
  final FileAttachment?file;final LocationAttachment?location;
  const Attachment({this.image,this.audio,this.video,this.file,this.location});}

class ImageAttachment{final String url;final int?width,height,size;final String?format,md5;
  const ImageAttachment({required this.url,this.width,this.height,this.format,this.size,this.md5});}
class AudioAttachment{final String url;final int?duration,size;final String?format;
  const AudioAttachment({required this.url,this.duration,this.size,this.format});}
class VideoAttachment{final String url,format,thumbUrl;final int?duration,width,height,size;
  const VideoAttachment({required this.url,this.duration,this.width,this.height,this.size,this.format='',this.thumbUrl=''});}
class FileAttachment{final String url,filename;final int?size;final String?format,md5;
  const FileAttachment({required this.url,required this.filename,this.size,this.format,this.md5});}
class LocationAttachment{final double latitude,longitude;final String?title,address;
  const LocationAttachment({required this.latitude,required this.longitude,this.title,this.address});}
class ThreadReply{final int replyToMsgId;final String replyToClientId,replyToSender,replyContent;
  const ThreadReply({required this.replyToMsgId,required this.replyToClientId,required this.replyToSender,required this.replyContent});}
