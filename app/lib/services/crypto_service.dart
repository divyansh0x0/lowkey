import 'dart:convert';
import 'package:pinenacl/x25519.dart';

/// Client-side X25519 DH Ratchet + XSalsa20-Poly1305 (NaCl Box).
///
/// Implements a Diffie-Hellman ratchet for forward secrecy:
/// - Each "turn change" generates a fresh X25519 key pair for sending
/// - Old private keys are discarded — compromising one key only exposes
///   that turn's messages, not past or future ones
/// - The server NEVER sees any private keys or shared secrets
///
/// Message format: JSON {"rk": "<ratchet_pubkey>", "ct": "<ciphertext>"}
class CryptoService {
  // Identity key pair — used for initial key exchange only
  late final PrivateKey _identityPrivateKey;
  late final PublicKey _identityPublicKey;

  // Current ratchet key pair (rotates on each turn change)
  PrivateKey? _ratchetPrivateKey;
  PublicKey? _ratchetPublicKey;

  // Peer's latest ratchet public key
  PublicKey? _peerRatchetPublicKey;

  // Whether to generate a new key pair on the next send
  bool _shouldRatchet = false;

  // Whether initial key exchange is done
  bool _initialized = false;

  CryptoService() {
    _identityPrivateKey = PrivateKey.generate();
    _identityPublicKey = _identityPrivateKey.publicKey;
  }

  /// Our identity public key as base64 (sent to peer via server relay).
  String get publicKeyBase64 => base64Encode(Uint8List.fromList(_identityPublicKey));

  /// Bootstrap the ratchet from the initial key exchange.
  /// Both sides start with identity keys as their first ratchet keys.
  void deriveSharedKey(String peerPublicKeyBase64) {
    final peerPk = PublicKey(Uint8List.fromList(base64Decode(peerPublicKeyBase64)));
    _peerRatchetPublicKey = peerPk;
    _ratchetPrivateKey = _identityPrivateKey;
    _ratchetPublicKey = _identityPublicKey;
    _shouldRatchet = false;
    _initialized = true;
  }

  bool get hasKey => _initialized;

  /// Encrypt plaintext with DH ratchet.
  ///
  /// On each turn change (after receiving a message from peer), a fresh
  /// X25519 key pair is generated before encrypting. The ratchet public
  /// key is included in the message so the peer can derive the same
  /// shared secret.
  String encryptMessage(String plaintext) {
    if (!_initialized) throw StateError('Shared key not derived');

    // DH Ratchet step: generate fresh key pair after a turn change
    if (_shouldRatchet) {
      final newKey = PrivateKey.generate();
      _ratchetPrivateKey = newKey;
      _ratchetPublicKey = newKey.publicKey;
      _shouldRatchet = false;
    }

    // Create Box with current ratchet keys: DH(myPriv, peerPub)
    final box = Box(
      myPrivateKey: _ratchetPrivateKey!,
      theirPublicKey: _peerRatchetPublicKey!,
    );

    final encrypted = box.encrypt(Uint8List.fromList(utf8.encode(plaintext)));

    // Bundle ratchet public key + ciphertext
    return jsonEncode({
      'rk': base64Encode(Uint8List.fromList(_ratchetPublicKey!)),
      'ct': base64Encode(Uint8List.fromList(encrypted)),
    });
  }

  /// Decrypt a ratcheted message.
  ///
  /// If the peer's ratchet public key has changed (turn change), the
  /// receiving context is updated and our own keys are marked for
  /// ratchet on the next send.
  String decryptMessage(String message) {
    if (!_initialized) throw StateError('Shared key not derived');

    final parsed = jsonDecode(message) as Map<String, dynamic>;
    final peerRkBytes = base64Decode(parsed['rk'] as String);
    final ctBytes = base64Decode(parsed['ct'] as String);

    final peerRk = PublicKey(Uint8List.fromList(peerRkBytes));

    // If peer's ratchet key changed → mark for ratchet on next send
    if (!_publicKeysEqual(peerRk, _peerRatchetPublicKey!)) {
      _peerRatchetPublicKey = peerRk;
      _shouldRatchet = true;
    }

    // Decrypt: DH(myPriv, peerRatchetPub) derives the same shared secret
    // that the peer used to encrypt with DH(peerPriv, myPub)
    final box = Box(
      myPrivateKey: _ratchetPrivateKey!,
      theirPublicKey: peerRk,
    );

    final nonce = Uint8List.fromList(ctBytes.sublist(0, 24));
    final cipher = Uint8List.fromList(ctBytes.sublist(24));
    final encrypted = EncryptedMessage(nonce: nonce, cipherText: cipher);
    final decrypted = box.decrypt(encrypted);
    return utf8.decode(decrypted);
  }

  /// Compare two public keys byte-by-byte.
  bool _publicKeysEqual(PublicKey a, PublicKey b) {
    if (a.length != b.length) return false;
    for (int i = 0; i < a.length; i++) {
      if (a[i] != b[i]) return false;
    }
    return true;
  }

  void clear() {
    _ratchetPrivateKey = null;
    _ratchetPublicKey = null;
    _peerRatchetPublicKey = null;
    _shouldRatchet = false;
    _initialized = false;
  }
}
