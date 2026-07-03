import 'dart:async';
import 'dart:math';
import 'dart:typed_data';
import 'package:web_socket_channel/web_socket_channel.dart';
import '../ws/frame.dart';

enum ConnectionState { disconnected, connecting, connected }

typedef OnFrame = void Function(Frame);
typedef OnState = void Function(ConnectionState);
typedef OnErr = void Function(Object);
typedef OnTokenExpired = Future<String> Function();

class ConnectionManager {
  final String _url;
  final Duration _maxReconnect;
  final OnFrame onFrame;
  final OnState? onStateChange;
  final OnErr? onError;
  final OnTokenExpired? onTokenExpired;
  WebSocketChannel? _ch;
  StreamSubscription? _sub;
  ConnectionState _state = ConnectionState.disconnected;
  int _attempts = 0;
  Timer? _timer;
  bool _done = false;
  final _random = Random();

  ConnectionManager({
    required String host,
    required int port,
    required String appId,
    required String userId,
    required String token,
    String deviceId = 'default',
    this.onStateChange,
    this.onError,
    required this.onFrame,
    this.onTokenExpired,
    Duration maxReconnect = const Duration(seconds: 30),
  }) : _maxReconnect = maxReconnect,
       _url = '${port == 443 ? "wss" : "ws"}://$host:$port/ws'
           '?app_id=$appId&user_id=$userId&token=$token&device_id=$deviceId';

  ConnectionState get state => _state;

  Future<void> connect() async {
    if (_done) return;
    _set(ConnectionState.connecting);
    try {
      _ch = WebSocketChannel.connect(Uri.parse(_url));
      await _ch!.ready;
      _attempts = 0;
      _set(ConnectionState.connected);
      _sub = _ch!.stream.listen(
        (d) {
          if (d is List<int>) {
            final bytes = d is Uint8List ? d : Uint8List.fromList(d);
            final f = Frame.decode(bytes);
            if (f != null) onFrame(f);
          }
        },
        onError: (e) {
          _set(ConnectionState.disconnected);
          _reconnect();
          onError?.call(e);
        },
        onDone: () {
          if (_state != ConnectionState.disconnected) {
            _set(ConnectionState.disconnected);
            _reconnect();
          }
        },
      );
    } catch (e) {
      _set(ConnectionState.disconnected);
      if (_attempts == 0) rethrow;
      _reconnect();
    }
  }

  void send(Frame f) => _ch?.sink.add(f.encode());

  Future<void> disconnect() async {
    _done = true;
    _timer?.cancel();
    await _sub?.cancel();
    await _ch?.sink.close();
    _set(ConnectionState.disconnected);
  }

  void _reconnect() {
    if (_done) return;
    _timer?.cancel();
    final base = 1000 * pow(2, _attempts).toInt();
    final jitter = _random.nextInt(base ~/ 2);
    final ms = min(base + jitter, _maxReconnect.inMilliseconds);
    _attempts++;
    _timer = Timer(Duration(milliseconds: ms), () => connect());
  }

  void _set(ConnectionState s) {
    if (_state != s) {
      _state = s;
      onStateChange?.call(s);
    }
  }
}
