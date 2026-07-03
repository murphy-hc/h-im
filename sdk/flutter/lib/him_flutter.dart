/// H-IM Flutter SDK — WebSocket real-time messaging + chat widgets.
library him_flutter;

// Protocol
export 'src/ws/frame.dart' show Frame, FrameType, currentVersion;
export 'src/ws/heartbeat.dart' show HeartbeatManager;
export 'src/core/him_client.dart' show HimClient, HimConfig, HimCallbacks;
export 'src/core/connection.dart' show ConnectionState;
export 'src/core/session.dart' show Session;

// Models
export 'src/models/message.dart' show Message, MsgType, MsgStatus, ConvType, Attachment;
export 'src/models/conversation.dart' show Conversation;

// State
export 'src/state/message_stream.dart' show MessageStream, MessageEvent;

// Widgets
export 'src/widgets/chat_bubble.dart' show ChatBubble;
export 'src/widgets/message_list.dart' show MessageList;
export 'src/widgets/conversation_list.dart' show ConversationList;

// Platform
export 'src/platform/push_setup.dart' show PushSetup;
