import { 
  AppState, 
  type AppStateData, 
  type AppStateChangeCallback, 
  type RoomInfo, 
  type NetworkRequest,
  type JoinRoomRequest 
} from '../types/app-types';

export class AppStateManager {
  private currentState: AppState = AppState.CONNECTING;
  private data: AppStateData = {
    currentState: AppState.CONNECTING,
    rooms: [],
    currentRoom: null,
    connectionStatus: false,
    errorMessage: null,
    isLoading: false
  };

  private stateChangeCallbacks: AppStateChangeCallback[] = [];
  private networkRequestCallbacks: ((request: NetworkRequest) => void)[] = [];

  // State management methods
  setState(newState: AppState, updateData?: Partial<AppStateData>): void {
    const previousState = this.currentState;
    this.currentState = newState;
    
    if (updateData) {
      this.data = { ...this.data, ...updateData };
    }
    
    this.data.currentState = newState;
    
    console.log(`AppState: ${previousState} → ${newState}`, this.data);
    
    // Notify all state change listeners
    this.stateChangeCallbacks.forEach(callback => {
      callback(newState, this.data);
    });
  }

  getState(): AppState {
    return this.currentState;
  }

  getData(): AppStateData {
    return { ...this.data };
  }

  // Event subscription methods
  onStateChange(callback: AppStateChangeCallback): void {
    this.stateChangeCallbacks.push(callback);
  }

  removeStateChangeCallback(callback: AppStateChangeCallback): void {
    const index = this.stateChangeCallbacks.indexOf(callback);
    if (index > -1) {
      this.stateChangeCallbacks.splice(index, 1);
    }
  }

  onNetworkRequest(callback: (request: NetworkRequest) => void): void {
    this.networkRequestCallbacks.push(callback);
  }

  removeNetworkRequestCallback(callback: (request: NetworkRequest) => void): void {
    const index = this.networkRequestCallbacks.indexOf(callback);
    if (index > -1) {
      this.networkRequestCallbacks.splice(index, 1);
    }
  }

  // Business logic methods
  requestConnect(): void {
    if (this.currentState !== AppState.CONNECTING) {
      this.setState(AppState.CONNECTING, { isLoading: true, errorMessage: null });
    }
    
    this.networkRequestCallbacks.forEach(callback => {
      callback({ type: 'connect' });
    });
  }

  handleConnectionSuccess(): void {
    this.setState(AppState.LOBBY, { 
      connectionStatus: true, 
      isLoading: false, 
      errorMessage: null 
    });
    
    // Auto-request room list when connected
    this.requestRoomList();
  }

  handleConnectionFailure(error: string): void {
    this.setState(AppState.DISCONNECTED, { 
      connectionStatus: false, 
      isLoading: false, 
      errorMessage: error 
    });
  }

  requestRoomList(): void {
    console.log('[AppState] requestRoomList called, currentState:', this.currentState);
    console.log('[AppState] Number of network callbacks:', this.networkRequestCallbacks.length);

    if (this.currentState === AppState.LOBBY) {
      this.setState(this.currentState, { isLoading: true });
    }

    this.networkRequestCallbacks.forEach((callback, index) => {
      console.log(`[AppState] Calling network callback ${index} with room_list request`);
      callback({ type: 'room_list' });
    });

    console.log('[AppState] requestRoomList completed');
  }

  updateRoomList(rooms: RoomInfo[]): void {
    console.log('[AppState] updateRoomList called with', rooms.length, 'rooms:', rooms);

    this.setState(this.currentState, {
      rooms: rooms,
      isLoading: false,
      errorMessage: null
    });

    console.log('[AppState] Room list updated in state');
  }

  requestJoinRoom(roomId: string): void {
    const room = this.data.rooms.find(r => r.room_id === roomId);
    if (!room) {
      this.handleError('Room not found');
      return;
    }

    const isFull = room.max_players > 0 && room.player_count >= room.max_players;
    if (isFull) {
      this.handleError('Room is full');
      return;
    }

    this.setState(AppState.JOINING, {
      isLoading: true,
      errorMessage: null
    });

    const joinRequest: JoinRoomRequest = {
      room_id: roomId
    };

    this.networkRequestCallbacks.forEach(callback => {
      callback({
        type: 'join_room',
        payload: joinRequest
      });
    });
  }

  handleJoinRoomSuccessByRoomId(roomId: string): void {
    const room = this.data.rooms.find(r => r.room_id === roomId);

    if (!room) {
      this.handleJoinRoomFailure(`Room ${roomId} not found in local cache`);
      return;
    }

    this.handleJoinRoomSuccess(room);
  }

  handleJoinRoomSuccess(roomInfo: RoomInfo): void {
    this.setState(AppState.IN_GAME, {
      currentRoom: roomInfo,
      isLoading: false,
      errorMessage: null
    });
  }

  handleJoinRoomFailure(error: string): void {
    this.setState(AppState.LOBBY, {
      isLoading: false,
      errorMessage: error
    });
  }

  requestLeaveRoom(): void {
    if (this.currentState === AppState.IN_GAME && this.data.currentRoom) {
      this.setState(AppState.LOBBY, { 
        currentRoom: null, 
        isLoading: false, 
        errorMessage: null 
      });

      this.networkRequestCallbacks.forEach(callback => {
        callback({ type: 'leave_room' });
      });

      // Refresh room list when returning to lobby
      setTimeout(() => this.requestRoomList(), 100);
    }
  }

  handleDisconnection(): void {
    this.setState(AppState.DISCONNECTED, { 
      connectionStatus: false, 
      currentRoom: null, 
      isLoading: false, 
      errorMessage: 'Connection lost' 
    });
  }

  handleError(error: string): void {
    this.setState(this.currentState, { 
      errorMessage: error, 
      isLoading: false 
    });
    
    console.error('AppStateManager error:', error);
  }

  clearError(): void {
    this.setState(this.currentState, { errorMessage: null });
  }

  // Utility methods
  isInGame(): boolean {
    return this.currentState === AppState.IN_GAME;
  }

  isConnected(): boolean {
    return this.data.connectionStatus;
  }

  getCurrentRoom(): RoomInfo | null {
    return this.data.currentRoom;
  }

  getRooms(): RoomInfo[] {
    return [...this.data.rooms];
  }
}