import React, { createContext, useContext, useEffect, useRef, useState } from 'react';
import { SignalingService } from './SignalingService';
import { WebRTCManager } from './WebRTCManager';
import { SdpExchange, IceCandidate } from '../proto/signaling_pb';
import { database } from '../database';

// TODO: Replace with your deployed backend URL
const BACKEND_URL = 'http://10.0.2.2:8080'; // Android emulator -> host machine

export type ConnectionState = 'disconnected' | 'connecting' | 'registered' | 'negotiating' | 'connected';

interface ServiceContextValue {
  signalingService: SignalingService;
  webRTCManager: WebRTCManager;
  myUuid: string;
  connectionState: ConnectionState;
  initiateConnection: (targetUuid: string) => void;
}

const ServiceContext = createContext<ServiceContextValue | null>(null);

export const useServices = (): ServiceContextValue => {
  const ctx = useContext(ServiceContext);
  if (!ctx) {
    throw new Error('useServices must be used within a ServiceProvider');
  }
  return ctx;
};

interface ServiceProviderProps {
  children: React.ReactNode;
}

export const ServiceProvider: React.FC<ServiceProviderProps> = ({ children }) => {
  const signalingRef = useRef<SignalingService | null>(null);
  const webRTCRef = useRef<WebRTCManager | null>(null);
  const [connectionState, setConnectionState] = useState<ConnectionState>('disconnected');

  // Initialize services once
  if (!signalingRef.current) {
    signalingRef.current = new SignalingService(BACKEND_URL);
  }
  if (!webRTCRef.current) {
    webRTCRef.current = new WebRTCManager(signalingRef.current, database);
  }

  const signaling = signalingRef.current;
  const webRTC = webRTCRef.current;

  useEffect(() => {
    // Wire signaling callbacks to WebRTC manager
    signaling.setCallbacks({
      onSdp: (sdp: SdpExchange) => {
        const sdpType = sdp.getType();
        const sdpString = sdp.getSdp();
        const fromUuid = sdp.getTargetUuid();

        console.log(`[ServiceContext] Received SDP type=${sdpType} from=${fromUuid}`);

        if (sdpType === SdpExchange.Type.TYPE_OFFER) {
          // We are the callee — auto-answer the offer
          setConnectionState('negotiating');
          webRTC.createAnswer(sdpString, fromUuid);
        } else if (sdpType === SdpExchange.Type.TYPE_ANSWER) {
          // We are the caller — finalize with the answer
          webRTC.handleAnswer(sdpString);
        }
      },
      onIce: (ice: IceCandidate) => {
        console.log('[ServiceContext] Received ICE candidate');
        webRTC.handleIceCandidate(ice);
      },
      onIdentity: (_identity) => {
        console.log('[ServiceContext] Identity registered');
        setConnectionState('registered');
      },
      onError: (error) => {
        console.error('[ServiceContext] Signaling error:', error);
      },
    });

    // Connect to the signaling server
    setConnectionState('connecting');
    signaling.connect();

    // Monitor WebRTC connection state
    const handleConnectionStateChange = () => {
      const state = webRTC.peerConnection.connectionState;
      console.log('[ServiceContext] WebRTC state:', state);
      if (state === 'connected') {
        setConnectionState('connected');
      } else if (state === 'failed' || state === 'disconnected' || state === 'closed') {
        setConnectionState('registered'); // Fall back to registered but not connected
      }
    };

    webRTC.peerConnection.addEventListener(
      'connectionstatechange',
      handleConnectionStateChange,
    );

    return () => {
      webRTC.peerConnection.removeEventListener(
        'connectionstatechange',
        handleConnectionStateChange,
      );
      signaling.disconnect();
      webRTC.close();
    };
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const initiateConnection = (targetUuid: string) => {
    setConnectionState('negotiating');
    webRTC.createOffer(targetUuid);
  };

  const value: ServiceContextValue = {
    signalingService: signaling,
    webRTCManager: webRTC,
    myUuid: signaling.myUuid,
    connectionState,
    initiateConnection,
  };

  return (
    <ServiceContext.Provider value={value}>
      {children}
    </ServiceContext.Provider>
  );
};
