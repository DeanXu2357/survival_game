
import type { PlayerInput, ClientGameState } from './state';
import { gameState } from './state';

// const SERVER_URL = 'ws://localhost:3033/ws';
const CLIENT_ID = 'player-001';
const GAME_NAME = 'default_room';
const PLAYER_NAME = 'TestPlayer';

let ws: WebSocket | null = null;

// interface ConnectionRequest {
//   game_name: string;
//   client_id: string;
//   name: string;
//   session_id?: string;
// }

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
    
    const wsUrl = `ws://localhost:3033/ws?client_id=${CLIENT_ID}&game_name=${GAME_NAME}&name=${PLAYER_NAME}`;
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
