// Import existing protocol types
import type { PlayerInput } from '../state';

// Existing protocol types
export interface SystemSetSessionPayload {
  client_id: string;
  session_id: string;
}

export interface RequestEnvelope {
  envelope_type: string;
  payload: any;
}

export interface ResponseEnvelope {
  envelope_type: string;
  payload: any;
}

// New room-related protocol types
export interface RoomListRequestPayload {
  // Empty for now, might include filters in the future
}

export interface RoomListResponsePayload {
  rooms: Array<{
    id: string;
    name: string;
    player_count: number;
    max_players: number;
    status: 'waiting' | 'playing' | 'full';
    game_mode?: string;
  }>;
}

export interface JoinRoomRequestPayload {
  room_id: string;
  client_id: string;
  name: string;
}

export interface JoinRoomResponsePayload {
  success: boolean;
  message?: string;
  room_info?: {
    id: string;
    name: string;
    player_count: number;
    max_players: number;
    status: 'waiting' | 'playing' | 'full';
    game_mode?: string;
  };
}

export interface LeaveRoomRequestPayload {
  room_id?: string; // Optional, server can infer from client
}

export interface LeaveRoomResponsePayload {
  success: boolean;
  message?: string;
}

// Envelope type constants
export const REQUEST_TYPES = {
  PLAYER_INPUT: 'player_input',
  ROOM_LIST: 'room_list',
  JOIN_ROOM: 'join_room',
  LEAVE_ROOM: 'leave_room'
} as const;

export const RESPONSE_TYPES = {
  GAME_UPDATE: 'game_update',
  STATIC_DATA: 'static_data',
  SYSTEM_NOTIFY: 'system_notify',
  SYSTEM_SET_SESSION: 'system_set_session',
  ERROR_INVALID_SESSION: 'error_invalid_session',
  ROOM_LIST_RESPONSE: 'room_list_response',
  JOIN_ROOM_RESPONSE: 'join_room_response',
  LEAVE_ROOM_RESPONSE: 'leave_room_response'
} as const;

// Helper functions for creating protocol messages
export function createRoomListRequest(): RequestEnvelope {
  return {
    envelope_type: REQUEST_TYPES.ROOM_LIST,
    payload: {}
  };
}

export function createJoinRoomRequest(roomId: string, clientId: string, name: string): RequestEnvelope {
  return {
    envelope_type: REQUEST_TYPES.JOIN_ROOM,
    payload: {
      room_id: roomId,
      client_id: clientId,
      name: name
    } as JoinRoomRequestPayload
  };
}

export function createLeaveRoomRequest(roomId?: string): RequestEnvelope {
  return {
    envelope_type: REQUEST_TYPES.LEAVE_ROOM,
    payload: {
      room_id: roomId
    } as LeaveRoomRequestPayload
  };
}

export function createPlayerInputMessage(input: PlayerInput): RequestEnvelope {
  return {
    envelope_type: REQUEST_TYPES.PLAYER_INPUT,
    payload: input
  };
}

// Type guards for response validation
export function isRoomListResponse(envelope: ResponseEnvelope): envelope is ResponseEnvelope & { payload: RoomListResponsePayload } {
  return envelope.envelope_type === RESPONSE_TYPES.ROOM_LIST_RESPONSE;
}

export function isJoinRoomResponse(envelope: ResponseEnvelope): envelope is ResponseEnvelope & { payload: JoinRoomResponsePayload } {
  return envelope.envelope_type === RESPONSE_TYPES.JOIN_ROOM_RESPONSE;
}

export function isLeaveRoomResponse(envelope: ResponseEnvelope): envelope is ResponseEnvelope & { payload: LeaveRoomResponsePayload } {
  return envelope.envelope_type === RESPONSE_TYPES.LEAVE_ROOM_RESPONSE;
}

export function isSystemSetSession(envelope: ResponseEnvelope): envelope is ResponseEnvelope & { payload: SystemSetSessionPayload } {
  return envelope.envelope_type === RESPONSE_TYPES.SYSTEM_SET_SESSION;
}