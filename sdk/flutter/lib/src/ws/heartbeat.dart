import 'dart:async';import 'frame.dart';
class HeartbeatManager{final void Function(Frame)send;final Duration interval,timeout;final void Function()?onTimeout;
  Timer?_t;DateTime?_last;bool _r=false;
  HeartbeatManager({required this.send,this.interval=const Duration(seconds:10),
    this.timeout=const Duration(seconds:180),this.onTimeout});
  bool get isRunning=>_r;
  void start(){if(_r)return;_r=true;_last=DateTime.now();_t=Timer.periodic(interval,(_){if(!_r)return;
    send(Frame.heartbeat());if(_last!=null&&DateTime.now().difference(_last!)>timeout){_r=false;_t?.cancel();onTimeout?.call();}});}
  void onEcho()=>_last=DateTime.now();
  void stop(){_r=false;_t?.cancel();_t=null;}
}
