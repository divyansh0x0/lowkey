import React, { useState } from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  SafeAreaView,
  FlatList,
  KeyboardAvoidingView,
  Platform,
  StatusBar
} from 'react-native';
import { Camera, Mic, Send, Lock, ArrowLeft } from 'lucide-react-native';
import withObservables from '@nozbe/with-observables';

// --- Type Definitions ---
interface MessageRecord {
  id: string;
  ciphertext: string;
  sender_id: string;
  created_at: number;
}

interface ChatScreenProps {
  messages: MessageRecord[];
  targetUuid?: string;
  onGoBack?: () => void;
}

// Temporary myUuid for mocking "self" vs "partner" bubbles
const MY_SUPER_DUMMY_UUID = '123e4567-e89b-12d3-a456-426614174000';

// --- Inner Component ---
const ChatScreenInner: React.FC<ChatScreenProps> = ({ messages: initialMessages, targetUuid, onGoBack }) => {
  const [inputText, setInputText] = useState('');
  const [localMessages, setLocalMessages] = useState<MessageRecord[]>(initialMessages);

  const displayTargetUuid = targetUuid || 'Unknown Partner';

  const handleSend = () => {
    if (!inputText.trim()) return;
    const newMessage: MessageRecord = {
      id: Date.now().toString(),
      ciphertext: inputText.trim(),
      sender_id: MY_SUPER_DUMMY_UUID,
      created_at: Date.now(),
    };
    setLocalMessages(prev => [...prev, newMessage]);
    // TODO: Wire WebRTCManager.dataChannel.send() and WatermelonDB save
    setInputText('');
  };

  const renderMessage = ({ item }: { item: MessageRecord }) => {
    const isMe = item.sender_id === MY_SUPER_DUMMY_UUID;

    return (
      <View style={[styles.messageWrapper, isMe ? styles.messageWrapperMe : styles.messageWrapperOther]}>
        <View style={[styles.bubble, isMe ? styles.bubbleMe : styles.bubbleOther]}>
          <Text style={[styles.messageText, isMe ? styles.messageTextMe : styles.messageTextOther]}>
            {item.ciphertext}
          </Text>
        </View>
      </View>
    );
  };

  return (
    <SafeAreaView style={styles.safeArea}>
      <StatusBar barStyle="dark-content" backgroundColor="#FFFFFF" />
      <KeyboardAvoidingView
        style={styles.keyboardAvoid}
        behavior={Platform.OS === 'ios' ? 'padding' : undefined}
        keyboardVerticalOffset={Platform.OS === 'ios' ? 0 : 20}
      >
        {/* Header */}
        <View style={styles.header}>
          <View style={styles.headerTopRow}>
            <TouchableOpacity style={styles.backBtn} onPress={onGoBack} activeOpacity={0.6}>
              <ArrowLeft color="#000000" size={22} strokeWidth={2} />
            </TouchableOpacity>
            <Text style={styles.brandTitle}>Lowkey</Text>
            {/* Spacer to keep title centered */}
            <View style={styles.backBtn} />
          </View>
          <View style={styles.partnerInfo}>
            <Text style={styles.partnerId} numberOfLines={1} ellipsizeMode="middle">
              {displayTargetUuid}
            </Text>
            <View style={styles.securityBadge}>
              <Lock color="#10B981" size={12} strokeWidth={2.5} style={styles.lockIcon} />
              <Text style={styles.securityText}>Secure P2P Encrypted</Text>
            </View>
          </View>
        </View>

        {/* Message List */}
        <FlatList
          data={localMessages}
          keyExtractor={(item) => item.id}
          renderItem={renderMessage}
          contentContainerStyle={styles.messageList}
          showsVerticalScrollIndicator={false}
          inverted={false} // Would normally be true for standard chat behavior dependent on query sorting
        />

        {/* Input Bar */}
        <View style={styles.inputContainer}>
          <TouchableOpacity style={styles.accessoryBtn}>
            <Camera color="#888888" size={24} strokeWidth={2} />
          </TouchableOpacity>
          <TouchableOpacity style={styles.accessoryBtn}>
            <Mic color="#888888" size={24} strokeWidth={2} />
          </TouchableOpacity>
          
          <TextInput
            style={styles.textInput}
            placeholder="Secure message..."
            placeholderTextColor="#A39E98"
            value={inputText}
            onChangeText={setInputText}
            multiline
            selectionColor="#000000"
          />

          <TouchableOpacity
            style={[styles.sendBtn, inputText.trim() ? styles.sendBtnActive : styles.sendBtnInactive]}
            onPress={handleSend}
            disabled={!inputText.trim()}
          >
            <Send color="#FFFFFF" size={20} strokeWidth={2.5} />
          </TouchableOpacity>
        </View>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
};

// --- HOC Wrapper ---
// In the future, the parent will pass the raw WatermelonDB Collection or Query for `messages`
// e.g., `<ChatScreen messages={database.get('messages').query()} />`
const enhanceWithMessages = withObservables(['messages'], ({ messages }: { messages: any }) => ({
  messages: messages || [], // The observable query
}));

// Export the decorated component
export const ChatScreen = enhanceWithMessages(ChatScreenInner);
export const ChatScreenRaw = ChatScreenInner;

// --- Styles ---
const styles = StyleSheet.create({
  safeArea: {
    flex: 1,
    backgroundColor: '#FFFFFF', // Pure white theme
  },
  keyboardAvoid: {
    flex: 1,
  },
  
  // Header
  header: {
    paddingHorizontal: 24,
    paddingTop: 16,
    paddingBottom: 20,
    borderBottomWidth: 1,
    borderBottomColor: '#F0F0F0',
    alignItems: 'center',
  },
  headerTopRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    width: '100%',
    marginBottom: 2,
  },
  backBtn: {
    width: 36,
    height: 36,
    borderRadius: 18,
    alignItems: 'center',
    justifyContent: 'center',
  },
  brandTitle: {
    fontFamily: 'StoryScript-Regular',
    fontSize: 42,
    color: '#000000',
    marginBottom: 4,
  },
  partnerInfo: {
    alignItems: 'center',
  },
  partnerId: {
    fontSize: 13,
    fontWeight: '700',
    color: '#333333',
    letterSpacing: 0.5,
    marginBottom: 4,
  },
  securityBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#ECFDF5', // Super light emerald
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderRadius: 12,
  },
  lockIcon: {
    marginRight: 4,
  },
  securityText: {
    fontSize: 10,
    fontWeight: '700',
    color: '#10B981', // Solid emerald
    letterSpacing: 0.5,
    textTransform: 'uppercase',
  },

  // Message List
  messageList: {
    paddingHorizontal: 20,
    paddingVertical: 24,
    flexGrow: 1,
  },
  messageWrapper: {
    flexDirection: 'row',
    marginBottom: 16,
  },
  messageWrapperMe: {
    justifyContent: 'flex-end',
  },
  messageWrapperOther: {
    justifyContent: 'flex-start',
  },
  bubble: {
    maxWidth: '75%',
    paddingHorizontal: 16,
    paddingVertical: 12,
    borderRadius: 20,
  },
  bubbleMe: {
    backgroundColor: '#E5F4F3', // Light teal
    borderBottomRightRadius: 4,
  },
  bubbleOther: {
    backgroundColor: '#FAF8F4', // Light cream/warm grey
    borderBottomLeftRadius: 4,
  },
  messageText: {
    fontSize: 16,
    lineHeight: 22,
    fontWeight: '400',
  },
  messageTextMe: {
    color: '#004D40', // Dark teal text for contrast
  },
  messageTextOther: {
    color: '#2C2C2C', // Dark text for contrast
  },

  // Bottom Input Bar
  inputContainer: {
    flexDirection: 'row',
    alignItems: 'flex-end',
    paddingHorizontal: 16,
    paddingVertical: 12,
    borderTopWidth: 1,
    borderTopColor: '#F0F0F0',
    backgroundColor: '#FFFFFF',
  },
  accessoryBtn: {
    padding: 10,
    marginRight: 4,
    marginBottom: 2,
  },
  textInput: {
    flex: 1,
    backgroundColor: '#F7F7F7',
    borderRadius: 20,
    paddingHorizontal: 16,
    paddingTop: 12, // Required for multiline text alignment
    paddingBottom: 12,
    fontSize: 16,
    color: '#000000',
    maxHeight: 120,
    marginRight: 10,
  },
  sendBtn: {
    borderRadius: 24,
    padding: 12,
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: 2,
  },
  sendBtnActive: {
    backgroundColor: '#009688', // Solid rich teal
  },
  sendBtnInactive: {
    backgroundColor: '#E0E0E0', // Greyed out
  },
});
