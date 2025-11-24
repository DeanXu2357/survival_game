import type { UIScreen, LobbyScreenConfig, RoomListItem } from '../types/ui-types';
import type { RoomInfo } from '../types/app-types';

export class LobbyScreen implements UIScreen {
  private config: LobbyScreenConfig;
  private container: HTMLElement;
  private visible: boolean = false;

  // UI elements
  private titleElement: HTMLElement | null = null;
  private roomListContainer: HTMLElement | null = null;
  private refreshButton: HTMLButtonElement | null = null;
  private createRoomButton: HTMLButtonElement | null = null;
  private errorContainer: HTMLElement | null = null;
  private loadingIndicator: HTMLElement | null = null;

  // Room management
  private roomListItems: Map<string, RoomListItem> = new Map();

  constructor(config: LobbyScreenConfig) {
    this.config = config;
    this.container = config.container;

    this.createUI();
    this.setupEventListeners();
  }

  private createUI(): void {
    // Clear container
    this.container.innerHTML = '';
    this.container.className = 'lobby-screen';

    // Create main structure
    const mainContainer = document.createElement('div');
    mainContainer.className = 'lobby-main';

    // Title
    this.titleElement = document.createElement('h1');
    this.titleElement.className = 'lobby-title';
    this.titleElement.textContent = 'Game Lobby';

    // Error container (hidden by default)
    this.errorContainer = document.createElement('div');
    this.errorContainer.className = 'error-message hidden';

    // Controls section
    const controlsContainer = document.createElement('div');
    controlsContainer.className = 'lobby-controls';

    this.refreshButton = document.createElement('button');
    this.refreshButton.className = 'btn btn-primary';
    this.refreshButton.textContent = 'Refresh Rooms';

    this.createRoomButton = document.createElement('button');
    this.createRoomButton.className = 'btn btn-secondary';
    this.createRoomButton.textContent = 'Create Room';
    this.createRoomButton.disabled = true; // TODO: Enable when backend supports it

    controlsContainer.appendChild(this.refreshButton);
    controlsContainer.appendChild(this.createRoomButton);

    // Room list section
    const roomListSection = document.createElement('div');
    roomListSection.className = 'room-list-section';

    const roomListTitle = document.createElement('h2');
    roomListTitle.className = 'room-list-title';
    roomListTitle.textContent = 'Available Rooms';

    this.roomListContainer = document.createElement('div');
    this.roomListContainer.className = 'room-list';

    // Loading indicator
    this.loadingIndicator = document.createElement('div');
    this.loadingIndicator.className = 'loading-indicator hidden';
    this.loadingIndicator.innerHTML = `
      <div class="spinner"></div>
      <span>Loading rooms...</span>
    `;

    roomListSection.appendChild(roomListTitle);
    roomListSection.appendChild(this.loadingIndicator);
    roomListSection.appendChild(this.roomListContainer);

    // Assemble main container
    mainContainer.appendChild(this.titleElement);
    mainContainer.appendChild(this.errorContainer);
    mainContainer.appendChild(controlsContainer);
    mainContainer.appendChild(roomListSection);

    this.container.appendChild(mainContainer);
  }

  private setupEventListeners(): void {
    if (this.refreshButton) {
      this.refreshButton.addEventListener('click', () => {
        this.config.onRefreshRooms();
      });
    }

    if (this.createRoomButton) {
      this.createRoomButton.addEventListener('click', () => {
        this.handleCreateRoom();
      });
    }
  }

  private handleCreateRoom(): void {
    const roomName = prompt('Enter room name:');
    if (roomName && roomName.trim()) {
      this.config.onCreateRoom(roomName.trim());
    }
  }

  show(): void {
    this.visible = true;
    this.container.classList.remove('hidden');
    console.log('Lobby screen shown');
  }

  hide(): void {
    this.visible = false;
    this.container.classList.add('hidden');
    console.log('Lobby screen hidden');
  }

  isVisible(): boolean {
    return this.visible;
  }

  updateRoomList(rooms: RoomInfo[]): void {
    if (!this.roomListContainer) return;

    console.log('Updating room list with', rooms.length, 'rooms');

    // Clear existing rooms
    this.roomListContainer.innerHTML = '';
    this.roomListItems.clear();

    if (rooms.length === 0) {
      const emptyMessage = document.createElement('div');
      emptyMessage.className = 'empty-room-list';
      emptyMessage.textContent = 'No rooms available. Click "Refresh Rooms" to check again.';
      this.roomListContainer.appendChild(emptyMessage);
      return;
    }

    // Create room items
    rooms.forEach(room => {
      const roomItem = this.createRoomListItem(room);
      this.roomListItems.set(room.room_id, roomItem);
      this.roomListContainer!.appendChild(roomItem.element);
    });
  }

  private createRoomListItem(room: RoomInfo): RoomListItem {
    const element = document.createElement('div');
    element.className = 'room-item';

    // Room info section
    const infoSection = document.createElement('div');
    infoSection.className = 'room-info';

    const nameElement = document.createElement('div');
    nameElement.className = 'room-name';
    nameElement.textContent = room.name || room.room_id;

    const detailsElement = document.createElement('div');
    detailsElement.className = 'room-details';

    const playerCountText = room.max_players > 0
      ? `${room.player_count}/${room.max_players} players`
      : `${room.player_count} players`;

    detailsElement.innerHTML = `
      <span class="player-count">${playerCountText}</span>
    `;

    infoSection.appendChild(nameElement);
    infoSection.appendChild(detailsElement);

    // Actions section
    const actionsSection = document.createElement('div');
    actionsSection.className = 'room-actions';

    const joinButton = document.createElement('button');
    joinButton.className = 'btn btn-join';
    joinButton.textContent = 'Join';

    joinButton.addEventListener('click', () => {
        this.config.onJoinRoom(room.room_id);
    });

    actionsSection.appendChild(joinButton);

    // Assemble room item
    element.appendChild(infoSection);
    element.appendChild(actionsSection);

    return {
      room,
      element,
      joinButton
    };
  }

  showError(message: string): void {
    if (this.errorContainer) {
      this.errorContainer.textContent = message;
      this.errorContainer.classList.remove('hidden');
      console.log('Lobby error shown:', message);

      // Auto-hide error after 5 seconds
      setTimeout(() => {
        this.clearError();
      }, 5000);
    }
  }

  clearError(): void {
    if (this.errorContainer) {
      this.errorContainer.classList.add('hidden');
      this.errorContainer.textContent = '';
    }
  }

  setLoading(isLoading: boolean): void {
    if (this.loadingIndicator) {
      if (isLoading) {
        this.loadingIndicator.classList.remove('hidden');
      } else {
        this.loadingIndicator.classList.add('hidden');
      }
    }

    // Disable buttons while loading
    if (this.refreshButton) {
      this.refreshButton.disabled = isLoading;
    }

    if (this.createRoomButton) {
      this.createRoomButton.disabled = isLoading;
    }

    // Disable all join buttons while loading
    this.roomListItems.forEach(item => {
      // item.joinButton.disabled = isLoading || item.room.status === 'full';
    });
  }

  destroy(): void {
    // Remove event listeners
    if (this.refreshButton) {
      this.refreshButton.removeEventListener('click', () => {});
    }

    if (this.createRoomButton) {
      this.createRoomButton.removeEventListener('click', () => {});
    }

    // Clear room items
    this.roomListItems.forEach(item => {
      item.joinButton.removeEventListener('click', () => {});
    });
    this.roomListItems.clear();

    // Clear container
    this.container.innerHTML = '';

    console.log('Lobby screen destroyed');
  }
}
