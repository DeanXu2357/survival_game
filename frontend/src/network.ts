
import type { PlayerInput, ClientGameState, StaticGameData } from './state';
import { gameState } from './state';
import { SessionManager } from './session';

// const SERVER_URL = 'ws://localhost:3033/ws';
const CLIENT_ID = 'player-001';
const GAME_NAME = 'default_room';
const PLAYER_NAME = 'TestPlayer';

let ws: WebSocket | null = null;
let isReconnecting = false;

interface SystemSetSessionPayload {
  client_id: string;
  session_id: string;
}

interface RequestEnvelope {
  type: string;
  payload: any;
}

interface ResponseEnvelope {
  type: string;
  payload: any;
}

export function connectToServer(): Promise<void> {
  return new Promise((resolve, reject) => {
    console.log('Connecting to WebSocket server...');

    // Clean up expired sessions before connecting
    SessionManager.cleanupExpiredSessions();

    // Try to get stored session ID
    const storedSessionId = SessionManager.getStoredSession(CLIENT_ID);
    const sessionParam = storedSessionId ? `&session_id=${storedSessionId}` : '';

    const wsUrl = `ws://localhost:3033/ws?client_id=${CLIENT_ID}&game_name=${GAME_NAME}&name=${PLAYER_NAME}${sessionParam}`;

    if (storedSessionId) {
      console.log(`Attempting to reconnect with session: ${storedSessionId}`);
    }

    ws = new WebSocket(wsUrl);

    ws.onopen = () => {
      console.log('WebSocket connected successfully');
      gameState.setCurrentPlayerID(CLIENT_ID);
      gameState.updateDebugInfo({ connectionStatus: true });
      console.log('Current player ID set to:', CLIENT_ID);
      resolve();
    };

    ws.onmessage = (event) => {
      try {
        const envelope: ResponseEnvelope = JSON.parse(event.data);
        handleServerMessage(envelope);
      } catch (error) {
        console.error('Error parsing server message:', error);
      }
    };

    ws.onclose = () => {
      console.log('WebSocket disconnected');
      gameState.updateDebugInfo({ connectionStatus: false });
      ws = null;

      // Auto-reconnect after a short delay (useful for session invalidation scenarios)
      if (!isReconnecting) {
        isReconnecting = true;
        console.log('Attempting to reconnect in 2 seconds...');
        setTimeout(() => {
          connectToServer().catch(err => {
            console.error('Auto-reconnection failed:', err);
          }).finally(() => {
            isReconnecting = false;
          });
        }, 2000);
      }
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      reject(error);
    };
  });
}

function handleServerMessage(envelope: ResponseEnvelope): void {
  switch (envelope.type) {
    case 'game_update':
      try {
        // The payload might be a string that needs to be parsed again
        let gameUpdate: ClientGameState;
        if (typeof envelope.payload === 'string') {
          gameUpdate = JSON.parse(envelope.payload);
        } else {
          gameUpdate = envelope.payload;
        }

        // Game state updated silently

        gameState.updateState(gameUpdate);
      } catch (error) {
        console.error('Error processing game update:', error);
        console.log('Raw payload type:', typeof envelope.payload);
        console.log('Raw payload:', envelope.payload);
      }
      break;
    case 'system_notify':
      console.log('System notification:', envelope.payload.message);
      break;
    case 'system_set_session':
      try {
        let sessionPayload: SystemSetSessionPayload;
        if (typeof envelope.payload === 'string') {
          sessionPayload = JSON.parse(envelope.payload);
        } else {
          sessionPayload = envelope.payload;
        }

        // Store the new session ID
        SessionManager.storeSession(sessionPayload.client_id, sessionPayload.session_id);
        gameState.setSessionId(sessionPayload.session_id);

        console.log('Session ID received and stored:', sessionPayload.session_id);
      } catch (error) {
        console.error('Error processing session ID:', error);
      }
      break;
    case 'error_invalid_session':
      try {
        let errorPayload: any;
        if (typeof envelope.payload === 'string') {
          errorPayload = JSON.parse(envelope.payload);
        } else {
          errorPayload = envelope.payload;
        }

        console.log('Session invalid, clearing local session:', errorPayload.message);

        // Clear local session storage
        SessionManager.clearSession(CLIENT_ID);
        gameState.clearSession();

        // Optionally show user-friendly message
        console.warn('Your session has expired. The page will automatically reconnect.');

        // Close current connection and trigger reconnection
        if (ws) {
          ws.close();
        }
      } catch (error) {
        console.error('Error processing session error:', error);
        // Still clear session even if parsing fails
        SessionManager.clearSession(CLIENT_ID);
        gameState.clearSession();

        // Close connection even if parsing failed
        if (ws) {
          ws.close();
        }
      }
      break;
    case 'static_data':
      try {
        let staticData: StaticGameData;
        if (typeof envelope.payload === 'string') {
          staticData = JSON.parse(envelope.payload);
        } else {
          staticData = envelope.payload;
        }

        console.log('Received static data:', staticData);
        // DEBUG: Log wall data specifically
        if (staticData.walls) {
          console.log(`Received ${staticData.walls.length} walls from backend.`);
          console.log(JSON.stringify(staticData.walls, null, 2));
        } else {
          console.log('Static data received, but it contains no walls array.');
        }
        
        gameState.updateStaticData(staticData);
        console.log('Static data stored in game state');
      } catch (error) {
        console.error('Error processing static data:', error);
      }
      break;
    default:
      console.log('Unknown message type:', envelope.type, envelope.payload);
  }
}

export function sendPlayerInput(input: PlayerInput): void {
  if (ws && ws.readyState === WebSocket.OPEN) {
    // Send PlayerInput directly, not wrapped in envelope
    ws.send(JSON.stringify(input));
  }
}

export function sendMessage(message: object) {
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify(message));
  }
}

export function isConnected(): boolean {
  return ws !== null && ws.readyState === WebSocket.OPEN;
}
