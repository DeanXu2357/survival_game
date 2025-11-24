import type { PlayerInput, ClientGameState, StaticGameData } from '../state';
import { gameState } from '../state';
import { SessionManager } from '../session';
import type { AppStateManager } from '../managers/app-state';
import type { NetworkRequest, RoomInfo } from '../types/app-types';
import {
  createRoomListRequest,
  createJoinRoomRequest,
  createLeaveRoomRequest,
  createPlayerInputMessage,
  isRoomListResponse,
  isJoinRoomSuccess,
  isErrorResponse,
  isLeaveRoomResponse,
  isSystemSetSession,
  RESPONSE_TYPES,
  type ResponseEnvelope,
  type SystemSetSessionPayload,
  type RoomListResponsePayload,
  type JoinRoomResponsePayload
} from './protocols';

export class NetworkClient {
  private ws: WebSocket | null = null;
  private appState: AppStateManager;
  private isReconnecting: boolean = false;
  private reconnectAttempts: number = 0;
  private maxReconnectAttempts: number = 5;
  private reconnectDelay: number = 2000;

  // Connection parameters
  private serverUrl: string = 'ws://localhost:3033/ws';
  private clientId: string = '';
  private playerName: string = '';
  private pendingJoinRoomId: string | null = null;

  constructor(appState: AppStateManager) {
    this.appState = appState;
    this.setupAppStateListeners();
    this.initializeClientInfo();
  }

  private setupAppStateListeners(): void {
    // Listen for network requests from app state
    this.appState.onNetworkRequest((request: NetworkRequest) => {
      this.handleNetworkRequest(request);
    });
  }

  private initializeClientInfo(): void {
    // Generate or retrieve client ID
    this.clientId = this.generateClientId();
    this.playerName = this.getPlayerName();
    
    console.log('NetworkClient initialized:', { clientId: this.clientId, playerName: this.playerName });
  }

  private generateClientId(): string {
    // Try to get from localStorage first
    let clientId = localStorage.getItem('client_id');
    if (!clientId) {
      clientId = `client-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
      localStorage.setItem('client_id', clientId);
    }
    return clientId;
  }

  private getPlayerName(): string {
    const savedName = localStorage.getItem('player_name');
    return savedName || `Player${Math.floor(Math.random() * 1000)}`;
  }

  private handleNetworkRequest(request: NetworkRequest): void {
    console.log('NetworkClient: Handling request:', request.type);

    switch (request.type) {
      case 'connect':
        this.connect();
        break;
      case 'room_list':
        this.requestRoomList();
        break;
      case 'join_room':
        if (request.payload) {
          this.joinRoom(request.payload.room_id);
        }
        break;
      case 'leave_room':
        this.leaveRoom();
        break;
      default:
        console.warn('Unknown network request type:', request.type);
    }
  }

  async connect(): Promise<void> {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      console.log('Already connected to server');
      return;
    }

    console.log('Connecting to WebSocket server...');

    try {
      // Clean up any existing connection
      if (this.ws) {
        this.ws.close();
        this.ws = null;
      }

      // Clean up expired sessions
      SessionManager.cleanupExpiredSessions();

      // Try to get stored session ID
      const storedSessionId = SessionManager.getStoredSession(this.clientId);
      const sessionParam = storedSessionId ? `&session_id=${storedSessionId}` : '';

      // For now, we connect without specifying a game room
      // The room will be selected later through the lobby
      const wsUrl = `${this.serverUrl}?client_id=${this.clientId}&name=${this.playerName}${sessionParam}`;

      if (storedSessionId) {
        console.log(`Attempting to reconnect with session: ${storedSessionId}`);
      }

      this.ws = new WebSocket(wsUrl);

      // Wait for connection to be established first
      await new Promise<void>((resolve, reject) => {
        const timeout = setTimeout(() => {
          reject(new Error('Connection timeout'));
        }, 10000);

        this.ws!.onopen = () => {
          clearTimeout(timeout);
          console.log('[NetworkClient] WebSocket connection opened (Promise resolved)');
          resolve();
        };

        this.ws!.onerror = (error) => {
          clearTimeout(timeout);
          reject(error);
        };
      });

      // Now setup permanent handlers (connection is already open)
      this.setupWebSocketHandlers();

      // Manually trigger the onopen logic since connection is already established
      console.log('WebSocket connected successfully');
      this.reconnectAttempts = 0;
      this.isReconnecting = false;

      gameState.setCurrentPlayerID(this.clientId);
      gameState.updateDebugInfo({ connectionStatus: true });

      this.appState.handleConnectionSuccess();

    } catch (error) {
      console.error('Failed to connect to server:', error);
      this.appState.handleConnectionFailure(error instanceof Error ? error.message : 'Connection failed');
    }
  }

  private setupWebSocketHandlers(): void {
    if (!this.ws) return;

    this.ws.onopen = () => {
      console.log('WebSocket connected successfully');
      this.reconnectAttempts = 0;
      this.isReconnecting = false;
      
      gameState.setCurrentPlayerID(this.clientId);
      gameState.updateDebugInfo({ connectionStatus: true });
      
      this.appState.handleConnectionSuccess();
    };

    this.ws.onmessage = (event) => {
      try {
        const envelope: ResponseEnvelope = JSON.parse(event.data);
        this.handleServerMessage(envelope);
      } catch (error) {
        console.error('Error parsing server message:', error);
      }
    };

    this.ws.onclose = () => {
      console.log('WebSocket disconnected');
      gameState.updateDebugInfo({ connectionStatus: false });
      this.ws = null;

      if (!this.isReconnecting && this.reconnectAttempts < this.maxReconnectAttempts) {
        this.attemptReconnect();
      } else {
        this.appState.handleDisconnection();
      }
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };
  }

  private attemptReconnect(): void {
    if (this.isReconnecting) return;

    this.isReconnecting = true;
    this.reconnectAttempts++;

    console.log(`Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts}) in ${this.reconnectDelay}ms...`);

    setTimeout(() => {
      this.connect().catch(error => {
        console.error('Reconnection failed:', error);
        if (this.reconnectAttempts >= this.maxReconnectAttempts) {
          this.appState.handleConnectionFailure('Failed to reconnect to server');
        } else {
          this.attemptReconnect();
        }
      }).finally(() => {
        this.isReconnecting = false;
      });
    }, this.reconnectDelay);
  }

  private handleServerMessage(envelope: ResponseEnvelope): void {
    console.log('Received server message:', envelope.envelope_type);

    switch (envelope.envelope_type) {
      case RESPONSE_TYPES.ROOM_LIST_RESPONSE:
        if (isRoomListResponse(envelope)) {
          this.handleRoomListResponse(envelope.payload);
        }
        break;

      case RESPONSE_TYPES.JOIN_ROOM_SUCCESS:
        if (isJoinRoomSuccess(envelope)) {
          this.handleJoinRoomSuccess();
        }
        break;

      case RESPONSE_TYPES.ERROR:
        if (isErrorResponse(envelope)) {
          this.handleErrorResponse(envelope.payload);
        }
        break;

      case RESPONSE_TYPES.LEAVE_ROOM_RESPONSE:
        if (isLeaveRoomResponse(envelope)) {
          this.handleLeaveRoomResponse(envelope.payload);
        }
        break;

      case RESPONSE_TYPES.GAME_UPDATE:
        this.handleGameUpdate(envelope.payload);
        break;

      case RESPONSE_TYPES.STATIC_DATA:
        this.handleStaticData(envelope.payload);
        break;

      case RESPONSE_TYPES.SYSTEM_SET_SESSION:
        if (isSystemSetSession(envelope)) {
          this.handleSystemSetSession(envelope.payload);
        }
        break;

      case RESPONSE_TYPES.ERROR_INVALID_SESSION:
        this.handleInvalidSession(envelope.payload);
        break;

      case RESPONSE_TYPES.SYSTEM_NOTIFY:
        this.handleSystemNotify(envelope.payload);
        break;

      default:
        console.log('Unknown message type:', envelope.envelope_type, envelope.payload);
    }
  }

  private handleRoomListResponse(payload: RoomListResponsePayload): void {
    console.log('[NetworkClient] handleRoomListResponse called');
    console.log('[NetworkClient] Payload:', payload);
    console.log('[NetworkClient] Received room list with', payload.rooms ? payload.rooms.length : 0, 'rooms:', payload.rooms);

    this.appState.updateRoomList(payload.rooms);
    console.log('[NetworkClient] Room list passed to AppState');
  }

  private handleJoinRoomSuccess(): void {
    console.log('Successfully joined room');

    if (!this.pendingJoinRoomId) {
      console.error('No pending room join found');
      this.appState.handleJoinRoomFailure('No pending room join');
      return;
    }

    this.appState.handleJoinRoomSuccessByRoomId(this.pendingJoinRoomId);
    this.pendingJoinRoomId = null;
  }

  private handleErrorResponse(payload: JoinRoomResponsePayload): void {
    console.log('Received error response:', payload);

    const errorMessage = payload.message || 'An error occurred';

    if (this.pendingJoinRoomId) {
      this.appState.handleJoinRoomFailure(errorMessage);
      this.pendingJoinRoomId = null;
    } else {
      console.warn('Error response received but no pending operation:', errorMessage);
    }
  }

  private handleLeaveRoomResponse(payload: any): void {
    console.log('Received leave room response:', payload);
    // Room leaving is handled by the app state manager
    // This is just confirmation from the server
  }

  private handleGameUpdate(payload: any): void {
    try {
      let gameUpdate: ClientGameState;
      if (typeof payload === 'string') {
        gameUpdate = JSON.parse(payload);
      } else {
        gameUpdate = payload;
      }

      gameState.updateState(gameUpdate);
    } catch (error) {
      console.error('Error processing game update:', error);
    }
  }

  private handleStaticData(payload: any): void {
    try {
      let staticData: StaticGameData;
      if (typeof payload === 'string') {
        staticData = JSON.parse(payload);
      } else {
        staticData = payload;
      }

      console.log('Received static data:', staticData);
      gameState.updateStaticData(staticData);
    } catch (error) {
      console.error('Error processing static data:', error);
    }
  }

  private handleSystemSetSession(payload: SystemSetSessionPayload): void {
    console.log('Session ID received:', payload.session_id);
    SessionManager.storeSession(payload.client_id, payload.session_id);
    gameState.setSessionId(payload.session_id);
  }

  private handleInvalidSession(payload: any): void {
    console.log('Session invalid, clearing local session:', payload.message);
    
    SessionManager.clearSession(this.clientId);
    gameState.clearSession();
    
    console.warn('Your session has expired. Reconnecting...');
    
    // Close current connection and trigger reconnection
    if (this.ws) {
      this.ws.close();
    }
  }

  private handleSystemNotify(payload: any): void {
    console.log('System notification:', payload.message);
  }

  // Public methods for sending messages
  requestRoomList(): void {
    console.log('[NetworkClient] requestRoomList called');
    console.log('[NetworkClient] Connection status:', this.isConnected());
    console.log('[NetworkClient] WebSocket state:', this.ws ? this.ws.readyState : 'null');

    if (!this.isConnected()) {
      console.warn('[NetworkClient] Cannot request room list: not connected');
      return;
    }

    const message = createRoomListRequest();
    console.log('[NetworkClient] Created room list request message:', message);
    this.sendMessage(message);
    console.log('[NetworkClient] Room list request sent');
  }

  joinRoom(roomId: string): void {
    console.log('[NetworkClient] joinRoom called with roomId:', roomId);

    if (!this.isConnected()) {
      console.warn('Cannot join room: not connected');
      this.appState.handleJoinRoomFailure('Not connected to server');
      return;
    }

    this.pendingJoinRoomId = roomId;
    const message = createJoinRoomRequest(roomId);
    console.log('[NetworkClient] Created join room request:', JSON.stringify(message, null, 2));
    this.sendMessage(message);
    console.log('[NetworkClient] Join room request sent');
  }

  leaveRoom(): void {
    if (!this.isConnected()) {
      console.warn('Cannot leave room: not connected');
      return;
    }

    const message = createLeaveRoomRequest();
    this.sendMessage(message);
  }

  sendPlayerInput(input: PlayerInput): void {
    if (!this.isConnected()) return;

    // Wrap player input in envelope as per protocol specification
    const message = createPlayerInputMessage(input);
    this.sendMessage(message);
  }

  private sendMessage(message: any): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
      console.log('Sent message:', message.envelope_type);
    }
  }

  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
  }

  getClientId(): string {
    return this.clientId;
  }


  setPlayerName(name: string): void {
    this.playerName = name;
    localStorage.setItem('player_name', name);
  }

  destroy(): void {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    
    this.isReconnecting = false;
    this.reconnectAttempts = 0;
    
    console.log('NetworkClient destroyed');
  }
}