/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React, { useState } from 'react';
import { StatusBar, StyleSheet, useColorScheme, View } from 'react-native';
import {
  SafeAreaProvider,
  useSafeAreaInsets,
} from 'react-native-safe-area-context';

import { ChatScreenRaw } from './src/screens/ChatScreen';
import { HomeScreen } from './src/screens/HomeScreen';

type Screen = 'home' | 'chat';

function App() {
  const isDarkMode = useColorScheme() === 'dark';

  return (
    <SafeAreaProvider>
      <StatusBar barStyle={isDarkMode ? 'light-content' : 'dark-content'} />
      <AppContent />
    </SafeAreaProvider>
  );
}

function AppContent() {
  const safeAreaInsets = useSafeAreaInsets();
  const [currentScreen, setCurrentScreen] = useState<Screen>('home');
  const [targetUuid, setTargetUuid] = useState<string>('');

  // Mocking an initial prop payload
  const mockMessages: any[] = [
    { id: '1', ciphertext: 'Hey! Secure connection established.', sender_id: 'other', created_at: Date.now() },
  ];

  const handleConnect = (uuid: string) => {
    setTargetUuid(uuid);
    setCurrentScreen('chat');
  };

  return (
    <View style={styles.container}>
      {currentScreen === 'home' ? (
        <HomeScreen onConnect={handleConnect} />
      ) : (
        <ChatScreenRaw messages={mockMessages as any} targetUuid={targetUuid} onGoBack={() => setCurrentScreen('home')} />
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
});

export default App;
