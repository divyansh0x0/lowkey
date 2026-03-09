/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import { NewAppScreen } from '@react-native/new-app-screen';
import { StatusBar, StyleSheet, useColorScheme, View } from 'react-native';
import {
  SafeAreaProvider,
  useSafeAreaInsets,
} from 'react-native-safe-area-context';

function App() {
  const isDarkMode = useColorScheme() === 'dark';

  return (
    <SafeAreaProvider>
      <StatusBar barStyle={isDarkMode ? 'light-content' : 'dark-content'} />
      <AppContent />
    </SafeAreaProvider>
  );
}

import { ChatScreenRaw } from './src/screens/ChatScreen';
import { HomeScreen } from './src/screens/HomeScreen';

function AppContent() {
  const safeAreaInsets = useSafeAreaInsets();

  // Mocking an initial prop payload
  const mockMessages: any[] = [
    { id: '1', ciphertext: 'Hey! Secure connection established.', sender_id: 'other', created_at: Date.now() },
    { id: '2', ciphertext: 'Awesome, I see it! This flat UI looks great.', sender_id: '123e4567-e89b-12d3-a456-426614174000', created_at: Date.now() + 1000 },
  ];

  return (
    <View style={styles.container}>
      <ChatScreenRaw messages={mockMessages as any} />
      <HomeScreen />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
});

export default App;
