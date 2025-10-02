export interface UIScreen {
  show(): void;
  hide(): void;
  isVisible(): boolean;
  destroy(): void;
}

export interface LobbyScreenConfig {
  container: HTMLElement;
  onJoinRoom: (roomId: string) => void;
  onRefreshRooms: () => void;
  onCreateRoom: (roomName: string) => void;
}

export interface GameHUDConfig {
  container: HTMLElement;
  onLeaveRoom: () => void;
  onToggleRenderer: () => void;
  onToggleSettings: () => void;
}

export interface LoadingScreenConfig {
  container: HTMLElement;
  message?: string;
}

export interface UIManagerConfig {
  appContainer: HTMLElement;
  screens: {
    lobby: HTMLElement;
    game: HTMLElement;
    loading: HTMLElement;
  };
}

export interface RoomListItem {
  room: import('./app-types').RoomInfo;
  element: HTMLElement;
  joinButton: HTMLButtonElement;
}