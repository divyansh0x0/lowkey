import React, { useState } from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  SafeAreaView,
  Clipboard,
  Alert,
  Platform,
  KeyboardAvoidingView,
  ScrollView,
  StatusBar
} from 'react-native';

export const HomeScreen = () => {
  // Dummy local UUID
  const myUuid = '123e4567-e89b-12d3-a456-426614174000';
  const [targetUuid, setTargetUuid] = useState('');

  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    Clipboard.setString(myUuid);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000); // Revert back after 2 seconds
  };

  const handleConnect = () => {
    if (!targetUuid.trim()) {
      Alert.alert('Hold up', "You need your partner's ID to connect.");
      return;
    }
    console.log('Initiating secure P2P connection to:', targetUuid);
    // TODO: Wire WebRTCManager.createOffer(targetUuid)
  };

  return (
    <SafeAreaView style={styles.safeArea}>
      <StatusBar barStyle="dark-content" backgroundColor="#FFFFFF" />
      <KeyboardAvoidingView 
        style={styles.keyboardAvoid} 
        behavior={Platform.OS === 'ios' ? 'padding' : undefined}
      >
        <ScrollView contentContainerStyle={styles.container} bounces={false}>
          
          {/* Brand Header */}
          <View style={styles.header}>
            <Text style={styles.brandTitle}>Lowkey</Text>
            <Text style={styles.subtitle}>Secure, serverless peer-to-peer connection.</Text>
          </View>

          {/* Your ID Card */}
          <View style={styles.card}>
            <View style={styles.cardHeader}>
              <Text style={styles.label}>YOUR ID</Text>
            </View>
            
            <View style={styles.uuidDisplayContainer}>
              <Text style={styles.uuidText} numberOfLines={1} ellipsizeMode="middle">
                {myUuid}
              </Text>
              <TouchableOpacity style={styles.copyButton} onPress={handleCopy} activeOpacity={0.6}>
                <Text style={styles.copyButtonText}>{copied ? 'Copied!' : 'Copy'}</Text>
              </TouchableOpacity>
            </View>
          </View>

          {/* Target ID Card */}
          <View style={styles.card}>
            <View style={styles.cardHeader}>
              <Text style={styles.label}>PARTNER'S ID</Text>
            </View>

            <View style={styles.inputContainer}>
              <TextInput
                style={styles.input}
                placeholder="Paste their ID here..."
                placeholderTextColor="#A39E98"
                value={targetUuid}
                onChangeText={setTargetUuid}
                autoCapitalize="none"
                autoCorrect={false}
                selectionColor="#000000"
              />
            </View>

            <TouchableOpacity style={styles.connectButton} onPress={handleConnect} activeOpacity={0.85}>
              <Text style={styles.connectButtonText}>Initiate Secure Connection</Text>
            </TouchableOpacity>
          </View>
          
          {/* Status Footer */}
          <View style={styles.footer}>
            <View style={styles.statusDot} />
            <Text style={styles.statusText}>Awaiting Handshake</Text>
          </View>

        </ScrollView>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  safeArea: {
    flex: 1,
    backgroundColor: '#FFFFFF', // Pure white theme
  },
  keyboardAvoid: {
    flex: 1,
  },
  container: {
    flexGrow: 1,
    paddingHorizontal: 32,
    justifyContent: 'center',
    paddingBottom: 40,
  },
  header: {
    marginBottom: 56,
    alignItems: 'center',
  },
  brandTitle: {
    fontFamily: 'StoryScript-Regular',
    fontSize: 64,
    color: '#000000', // Crisp black
    marginBottom: 8,
  },
  subtitle: {
    fontSize: 14,
    color: '#666666',
    fontWeight: '400',
    letterSpacing: 0.3,
    textAlign: 'center',
  },
  card: {
    marginBottom: 32,
  },
  cardHeader: {
    marginBottom: 12,
  },
  label: {
    fontSize: 12,
    fontWeight: '700',
    color: '#888888',
    letterSpacing: 1.5,
  },
  uuidDisplayContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#F7F7F7', // Very light grey for contrast against pure white
    borderRadius: 8,
    paddingLeft: 16,
    paddingRight: 8,
    paddingVertical: 8,
  },
  uuidText: {
    flex: 1,
    fontSize: 15,
    color: '#000000',
    fontWeight: '500',
    marginRight: 12,
    letterSpacing: 0.5,
  },
  copyButton: {
    backgroundColor: '#000000',
    paddingVertical: 10,
    paddingHorizontal: 16,
    borderRadius: 6,
  },
  copyButtonText: {
    color: '#FFFFFF',
    fontWeight: '600',
    fontSize: 13,
  },
  inputContainer: {
    marginBottom: 20,
  },
  input: {
    backgroundColor: '#F7F7F7',
    borderRadius: 8,
    paddingHorizontal: 16,
    paddingVertical: 16,
    fontSize: 16,
    color: '#000000',
    fontWeight: '400',
  },
  connectButton: {
    backgroundColor: '#000000', // Solid black button
    paddingVertical: 18,
    borderRadius: 8,
    alignItems: 'center',
    justifyContent: 'center',
  },
  connectButtonText: {
    color: '#FFFFFF',
    fontSize: 15,
    fontWeight: '700',
    letterSpacing: 0.5,
  },
  footer: {
    marginTop: 24,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
  },
  statusDot: {
    width: 6,
    height: 6,
    borderRadius: 3,
    backgroundColor: '#000000',
    marginRight: 8,
  },
  statusText: {
    fontSize: 12,
    color: '#888888',
    fontWeight: '500',
    letterSpacing: 0.5,
  }
});