import 'dart:async';
import 'dart:convert';
import 'package:web_socket_channel/web_socket_channel.dart';

/// Handles WebSocket communication with the Go signaling server.
class SignalingService {
  WebSocketChannel? _channel;
  String? _username;
  bool _connected = false;

  final _onSessionCreated = StreamController<Map<String, dynamic>>.broadcast();
  final _onSessionJoined = StreamController<Map<String, dynamic>>.broadcast();
  final _onSessionRequest = StreamController<Map<String, dynamic>>.broadcast();
  final _onKeyExchange = StreamController<Map<String, dynamic>>.broadcast();
  final _onSignal = StreamController<Map<String, dynamic>>.broadcast();
  final _onError = StreamController<Map<String, dynamic>>.broadcast();
  final _onConnectionChange = StreamController<bool>.broadcast();

  Stream<Map<String, dynamic>> get onSessionCreated => _onSessionCreated.stream;
  Stream<Map<String, dynamic>> get onSessionJoined => _onSessionJoined.stream;
  Stream<Map<String, dynamic>> get onSessionRequest => _onSessionRequest.stream;
  Stream<Map<String, dynamic>> get onKeyExchange => _onKeyExchange.stream;
  Stream<Map<String, dynamic>> get onSignal => _onSignal.stream;
  Stream<Map<String, dynamic>> get onError => _onError.stream;
  Stream<bool> get onConnectionChange => _onConnectionChange.stream;

  bool get isConnected => _connected;
  String? get username => _username;

  /// Connect to the signaling server.
  void connect(String serverUrl, String username) {
    _username = username;
    final uri = Uri.parse('$serverUrl/ws?username=$username');

    try {
      _channel = WebSocketChannel.connect(uri);
      _connected = true;
      _onConnectionChange.add(true);

      _channel!.stream.listen(
        (data) => _handleMessage(data),
        onError: (error) {
          _connected = false;
          _onConnectionChange.add(false);
        },
        onDone: () {
          _connected = false;
          _onConnectionChange.add(false);
        },
      );
    } catch (e) {
      _connected = false;
      _onConnectionChange.add(false);
    }
  }

  void _handleMessage(dynamic data) {
    final msg = jsonDecode(data as String) as Map<String, dynamic>;
    final type = msg['type'] as String;

    switch (type) {
      case 'session:created':
        final payload = jsonDecode(jsonEncode(msg['payload'])) as Map<String, dynamic>;
        _onSessionCreated.add(payload);
        break;
      case 'session:joined':
        final payload = jsonDecode(jsonEncode(msg['payload'])) as Map<String, dynamic>;
        _onSessionJoined.add(payload);
        break;
      case 'session:request':
        final payload = jsonDecode(jsonEncode(msg['payload'])) as Map<String, dynamic>;
        _onSessionRequest.add(payload);
        break;
      case 'key:exchange':
        _onKeyExchange.add(msg);
        break;
      case 'signal:offer':
      case 'signal:answer':
      case 'signal:ice':
        _onSignal.add(msg);
        break;
      case 'error':
        final payload = jsonDecode(jsonEncode(msg['payload'])) as Map<String, dynamic>;
        _onError.add(payload);
        break;
    }
  }

  void _send(Map<String, dynamic> msg) {
    if (_channel != null && _connected) {
      _channel!.sink.add(jsonEncode(msg));
    }
  }

  /// Request the server to create a new session.
  void createSession() {
    _send({'type': 'session:create'});
  }

  /// Join an existing session by UUID.
  void joinSession(String sessionId) {
    _send({'type': 'session:join', 'sessionId': sessionId});
  }

  /// Connect to a user by their username (server sends request to target).
  void connectToUser(String targetUsername) {
    _send({'type': 'session:connect', 'target': targetUsername});
  }

  /// Accept an incoming connection request.
  void acceptSession(String initiator) {
    _send({'type': 'session:accept', 'target': initiator});
  }

  /// Decline an incoming connection request.
  void declineSession(String initiator) {
    _send({'type': 'session:decline', 'target': initiator});
  }

  /// Send our X25519 public key to the peer.
  void sendPublicKey(String target, String publicKeyBase64) {
    _send({
      'type': 'key:exchange',
      'target': target,
      'payload': {'publicKey': publicKeyBase64},
    });
  }

  /// Send an SDP offer to a target peer.
  void sendOffer(String target, Map<String, dynamic> sdp) {
    _send({
      'type': 'signal:offer',
      'target': target,
      'payload': sdp,
    });
  }

  /// Send an SDP answer to a target peer.
  void sendAnswer(String target, Map<String, dynamic> sdp) {
    _send({
      'type': 'signal:answer',
      'target': target,
      'payload': sdp,
    });
  }

  /// Send an ICE candidate to a target peer.
  void sendIceCandidate(String target, Map<String, dynamic> candidate) {
    _send({
      'type': 'signal:ice',
      'target': target,
      'payload': candidate,
    });
  }

  /// Disconnect from the signaling server.
  void disconnect() {
    _channel?.sink.close();
    _connected = false;
    _onConnectionChange.add(false);
  }

  void dispose() {
    disconnect();
    _onSessionCreated.close();
    _onSessionJoined.close();
    _onSessionRequest.close();
    _onKeyExchange.close();
    _onSignal.close();
    _onError.close();
    _onConnectionChange.close();
  }
}
