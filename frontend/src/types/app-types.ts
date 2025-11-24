export enum AppState {
  CONNECTING = 'connecting',
  LOBBY = 'lobby',
  JOINING = 'joining',
  IN_GAME = 'in-game',
  DISCONNECTED = 'disconnected'
}

export interface RoomInfo {
  room_id: string;
  name: string;
  player_count: number;
  max_players: number;
}

export interface RoomListResponse {
  rooms: RoomInfo[];
}

export interface JoinRoomRequest {
  room_id: string;
}

export interface JoinRoomResponse {
  success: boolean;
  message?: string;
  room_info?: RoomInfo;
}

export interface AppStateData {
  currentState: AppState;
  rooms: RoomInfo[];
  currentRoom: RoomInfo | null;
  connectionStatus: boolean;
  errorMessage: string | null;
  isLoading: boolean;
}

export type AppStateChangeCallback = (state: AppState, data: AppStateData) => void;

export interface NetworkRequest {
  type: 'room_list' | 'join_room' | 'leave_room' | 'connect';
  payload?: any;
}