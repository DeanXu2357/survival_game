import type { UIScreen, GameHUDConfig } from '../types/ui-types';
import type { RoomInfo } from '../types/app-types';

interface NotificationOptions {
  type: 'info' | 'warning' | 'error' | 'success';
  duration?: number;
}

interface Notification {
  id: string;
  message: string;
  type: 'info' | 'warning' | 'error' | 'success';
  element: HTMLElement;
  timeoutId: number;
}

export class GameHUD implements UIScreen {
  private config: GameHUDConfig;
  private container: HTMLElement;
  private visible: boolean = false;
  
  // UI elements
  private hudContainer: HTMLElement | null = null;
  private roomInfoPanel: HTMLElement | null = null;
  private playerInfoPanel: HTMLElement | null = null;
  private controlsPanel: HTMLElement | null = null;
  private notificationsContainer: HTMLElement | null = null;
  
  // Notifications
  private notifications: Map<string, Notification> = new Map();
  private notificationIdCounter: number = 0;

  constructor(config: GameHUDConfig) {
    this.config = config;
    this.container = config.container;
    
    this.createUI();
    this.setupEventListeners();
  }

  private createUI(): void {
    // Clear container
    this.container.innerHTML = '';
    this.container.className = 'game-screen';

    // Create game canvas container (where the renderer will attach)
    const canvasContainer = document.createElement('div');
    canvasContainer.id = 'game-canvas-container';
    canvasContainer.className = 'game-canvas-container';

    // Create HUD overlay
    this.hudContainer = document.createElement('div');
    this.hudContainer.className = 'game-hud';

    // Top panel - Room info
    this.roomInfoPanel = document.createElement('div');
    this.roomInfoPanel.className = 'hud-panel room-info-panel';
    
    // Top-right panel - Player info
    this.playerInfoPanel = document.createElement('div');
    this.playerInfoPanel.className = 'hud-panel player-info-panel';

    // Bottom panel - Controls
    this.controlsPanel = document.createElement('div');
    this.controlsPanel.className = 'hud-panel controls-panel';
    this.createControlsPanel();

    // Notifications container
    this.notificationsContainer = document.createElement('div');
    this.notificationsContainer.className = 'notifications-container';

    // Assemble HUD
    this.hudContainer.appendChild(this.roomInfoPanel);
    this.hudContainer.appendChild(this.playerInfoPanel);
    this.hudContainer.appendChild(this.controlsPanel);
    this.hudContainer.appendChild(this.notificationsContainer);

    // Assemble main container
    this.container.appendChild(canvasContainer);
    this.container.appendChild(this.hudContainer);

    console.log('Game HUD created');
  }

  private createControlsPanel(): void {
    if (!this.controlsPanel) return;

    this.controlsPanel.innerHTML = `
      <div class="controls-section">
        <button id="leave-room-btn" class="btn btn-danger">Leave Room</button>
        <button id="toggle-renderer-btn" class="btn btn-secondary">Switch Renderer</button>
        <button id="settings-btn" class="btn btn-secondary">Settings</button>
      </div>
      <div class="controls-help">
        <div class="help-item">
          <span class="key">WASD</span> Move
        </div>
        <div class="help-item">
          <span class="key">Q/E</span> Rotate
        </div>
        <div class="help-item">
          <span class="key">Space</span> Fire
        </div>
        <div class="help-item">
          <span class="key">ESC</span> Leave Room
        </div>
      </div>
    `;
  }

  private setupEventListeners(): void {
    // Leave room button
    const leaveButton = this.container.querySelector('#leave-room-btn') as HTMLButtonElement;
    if (leaveButton) {
      leaveButton.addEventListener('click', () => {
        this.config.onLeaveRoom();
      });
    }

    // Toggle renderer button
    const rendererButton = this.container.querySelector('#toggle-renderer-btn') as HTMLButtonElement;
    if (rendererButton) {
      rendererButton.addEventListener('click', () => {
        this.config.onToggleRenderer();
      });
    }

    // Settings button
    const settingsButton = this.container.querySelector('#settings-btn') as HTMLButtonElement;
    if (settingsButton) {
      settingsButton.addEventListener('click', () => {
        this.config.onToggleSettings();
      });
    }
  }

  show(): void {
    this.visible = true;
    this.container.classList.remove('hidden');
    console.log('Game HUD shown');
  }

  hide(): void {
    this.visible = false;
    this.container.classList.add('hidden');
    console.log('Game HUD hidden');
  }

  isVisible(): boolean {
    return this.visible;
  }

  updateRoomInfo(roomInfo: RoomInfo): void {
    if (!this.roomInfoPanel) return;

    this.roomInfoPanel.innerHTML = `
      <div class="room-info-header">
        <h3 class="room-name">${roomInfo.name}</h3>
      </div>
      <div class="room-details">
        <div class="detail-item">
          <span class="label">Players:</span>
          <span class="value">${roomInfo.player_count}${roomInfo.max_players > 0 ? `/${roomInfo.max_players}` : ''}</span>
        </div>
      </div>
    `;
  }

  updatePlayerInfo(playerInfo: { name: string; id: string; health?: number; score?: number }): void {
    if (!this.playerInfoPanel) return;

    this.playerInfoPanel.innerHTML = `
      <div class="player-info-header">
        <h3 class="player-name">${playerInfo.name}</h3>
        <span class="player-id">${playerInfo.id}</span>
      </div>
      <div class="player-stats">
        ${playerInfo.health !== undefined ? `
          <div class="stat-item">
            <span class="label">Health:</span>
            <span class="value health">${playerInfo.health}</span>
          </div>
        ` : ''}
        ${playerInfo.score !== undefined ? `
          <div class="stat-item">
            <span class="label">Score:</span>
            <span class="value score">${playerInfo.score}</span>
          </div>
        ` : ''}
      </div>
    `;
  }

  showNotification(message: string, type: 'info' | 'warning' | 'error' | 'success' = 'info', duration: number = 5000): string {
    if (!this.notificationsContainer) return '';

    const id = `notification-${++this.notificationIdCounter}`;
    
    const element = document.createElement('div');
    element.className = `notification notification-${type}`;
    element.innerHTML = `
      <div class="notification-content">
        <span class="notification-message">${message}</span>
        <button class="notification-close">×</button>
      </div>
    `;

    // Close button handler
    const closeButton = element.querySelector('.notification-close') as HTMLButtonElement;
    closeButton.addEventListener('click', () => {
      this.removeNotification(id);
    });

    // Auto-remove timeout
    const timeoutId = window.setTimeout(() => {
      this.removeNotification(id);
    }, duration);

    // Store notification
    const notification: Notification = {
      id,
      message,
      type,
      element,
      timeoutId
    };
    
    this.notifications.set(id, notification);
    this.notificationsContainer.appendChild(element);

    console.log(`Game HUD notification: [${type}] ${message}`);
    return id;
  }

  removeNotification(id: string): void {
    const notification = this.notifications.get(id);
    if (!notification) return;

    // Clear timeout
    clearTimeout(notification.timeoutId);
    
    // Remove element with animation
    notification.element.classList.add('notification-fade-out');
    
    setTimeout(() => {
      if (this.notificationsContainer && notification.element.parentNode) {
        this.notificationsContainer.removeChild(notification.element);
      }
      this.notifications.delete(id);
    }, 300);
  }

  clearNotifications(): void {
    this.notifications.forEach((_, id) => {
      this.removeNotification(id);
    });
  }

  showConnectionStatus(connected: boolean): void {
    const type = connected ? 'success' : 'error';
    const message = connected ? 'Connected to server' : 'Connection lost';
    this.showNotification(message, type, connected ? 3000 : 10000);
  }

  getCanvasContainer(): HTMLElement | null {
    return this.container.querySelector('#game-canvas-container');
  }

  destroy(): void {
    // Clear all notifications
    this.clearNotifications();

    // Remove event listeners
    const leaveButton = this.container.querySelector('#leave-room-btn') as HTMLButtonElement;
    if (leaveButton) {
      leaveButton.removeEventListener('click', () => {});
    }

    const rendererButton = this.container.querySelector('#toggle-renderer-btn') as HTMLButtonElement;
    if (rendererButton) {
      rendererButton.removeEventListener('click', () => {});
    }

    const settingsButton = this.container.querySelector('#settings-btn') as HTMLButtonElement;
    if (settingsButton) {
      settingsButton.removeEventListener('click', () => {});
    }

    // Clear container
    this.container.innerHTML = '';
    
    console.log('Game HUD destroyed');
  }
}