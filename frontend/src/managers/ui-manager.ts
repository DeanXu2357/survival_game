import { AppState, type AppStateData } from '../types/app-types';
import type { UIManagerConfig } from '../types/ui-types';
import { LobbyScreen } from '../ui/lobby-screen';
import { GameHUD } from '../ui/game-hud';
import { LoadingScreen } from '../ui/loading-screen';
import type { AppStateManager } from './app-state';

export class UIManager {
  private config: UIManagerConfig;
  private appState: AppStateManager;
  
  // Screen components
  private lobbyScreen: LobbyScreen | null = null;
  private gameHUD: GameHUD | null = null;
  private loadingScreen: LoadingScreen | null = null;
  
  private currentScreen: AppState | null = null;

  constructor(appState: AppStateManager, config: UIManagerConfig) {
    this.appState = appState;
    this.config = config;
    
    this.initializeScreens();
    this.setupEventListeners();
  }

  private initializeScreens(): void {
    // Initialize lobby screen
    this.lobbyScreen = new LobbyScreen({
      container: this.config.screens.lobby,
      onJoinRoom: (roomId: string) => this.handleJoinRoom(roomId),
      onRefreshRooms: () => this.handleRefreshRooms(),
      onCreateRoom: (roomName: string) => this.handleCreateRoom(roomName)
    });

    // Initialize game HUD
    this.gameHUD = new GameHUD({
      container: this.config.screens.game,
      onLeaveRoom: () => this.handleLeaveRoom(),
      onToggleRenderer: () => this.handleToggleRenderer(),
      onToggleSettings: () => this.handleToggleSettings()
    });

    // Initialize loading screen
    this.loadingScreen = new LoadingScreen({
      container: this.config.screens.loading,
      message: 'Connecting...'
    });

    console.log('UI screens initialized');
  }

  private setupEventListeners(): void {
    // Listen to app state changes
    this.appState.onStateChange((newState: AppState, data: AppStateData) => {
      this.handleStateChange(newState, data);
    });

    // Handle window resize
    window.addEventListener('resize', () => {
      this.handleResize();
    });

    // Handle escape key
    document.addEventListener('keydown', (event) => {
      if (event.code === 'Escape') {
        this.handleEscapeKey();
      }
    });
  }

  handleStateChange(newState: AppState, data: AppStateData): void {
    console.log(`UI: State change ${this.currentScreen} → ${newState}`, data);
    
    // Hide current screen
    this.hideAllScreens();
    
    // Show appropriate screen based on state
    switch (newState) {
      case AppState.CONNECTING:
        this.showLoadingScreen('Connecting to server...');
        break;
        
      case AppState.LOBBY:
        this.showLobbyScreen(data);
        break;
        
      case AppState.JOINING:
        this.showLoadingScreen('Joining room...');
        break;
        
      case AppState.IN_GAME:
        this.showGameScreen(data);
        break;
        
      case AppState.DISCONNECTED:
        this.showLoadingScreen('Disconnected. Reconnecting...');
        break;
        
      default:
        console.warn(`Unknown app state: ${newState}`);
    }
    
    this.currentScreen = newState;
  }

  private hideAllScreens(): void {
    this.config.screens.lobby.style.display = 'none';
    this.config.screens.game.style.display = 'none';
    this.config.screens.loading.style.display = 'none';
    
    // Hide individual screen components
    if (this.lobbyScreen) this.lobbyScreen.hide();
    if (this.gameHUD) this.gameHUD.hide();
    if (this.loadingScreen) this.loadingScreen.hide();
  }

  private showLobbyScreen(data: AppStateData): void {
    this.config.screens.lobby.style.display = 'block';
    
    if (this.lobbyScreen) {
      this.lobbyScreen.show();
      this.lobbyScreen.updateRoomList(data.rooms);
      
      if (data.errorMessage) {
        this.lobbyScreen.showError(data.errorMessage);
      }
      
      this.lobbyScreen.setLoading(data.isLoading);
    }
  }

  private showGameScreen(data: AppStateData): void {
    this.config.screens.game.style.display = 'block';
    
    if (this.gameHUD) {
      this.gameHUD.show();
      
      if (data.currentRoom) {
        this.gameHUD.updateRoomInfo(data.currentRoom);
      }
    }
  }

  private showLoadingScreen(message: string): void {
    this.config.screens.loading.style.display = 'block';
    
    if (this.loadingScreen) {
      this.loadingScreen.show();
      this.loadingScreen.updateMessage(message);
    }
  }

  // Event handlers
  private handleJoinRoom(roomId: string): void {
    console.log(`UI: Join room requested: ${roomId}`);

    this.appState.requestJoinRoom(roomId);
  }

  private handleRefreshRooms(): void {
    console.log('UI: Refresh rooms requested');
    this.appState.requestRoomList();
  }

  private handleCreateRoom(roomName: string): void {
    console.log(`UI: Create room requested: ${roomName}`);
    // TODO: Implement room creation when backend supports it
    console.warn('Room creation not yet implemented');
  }

  private handleLeaveRoom(): void {
    console.log('UI: Leave room requested');
    this.appState.requestLeaveRoom();
  }

  private handleToggleRenderer(): void {
    console.log('UI: Toggle renderer requested');
    // TODO: Implement renderer switching when render manager is connected
    console.warn('Renderer switching not yet implemented');
  }

  private handleToggleSettings(): void {
    console.log('UI: Toggle settings requested');
    // TODO: Implement settings panel
    console.warn('Settings panel not yet implemented');
  }

  private handleEscapeKey(): void {
    switch (this.currentScreen) {
      case AppState.IN_GAME:
        this.handleLeaveRoom();
        break;
      case AppState.JOINING:
        // Cancel join attempt (if possible)
        // For now, just clear any error
        this.appState.clearError();
        break;
      default:
        // Clear any errors
        this.appState.clearError();
    }
  }

  private handleResize(): void {
    // Notify all screens about resize
    const width = window.innerWidth;
    const height = window.innerHeight;
    
    console.log(`UI: Window resized to ${width}x${height}`);
    
    // Game screen resize is handled by the render manager
    // Lobby and loading screens are CSS-based and should auto-resize
  }

  // Utility methods
  private generateClientId(): string {
    // Generate a unique client ID
    // In a real app, this might come from authentication or localStorage
    return `client-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  }

  private getPlayerName(): string {
    // Get player name from localStorage or default
    const savedName = localStorage.getItem('player_name');
    return savedName || `Player${Math.floor(Math.random() * 1000)}`;
  }

  // Public methods for external control
  showError(message: string): void {
    // Show error on current screen
    switch (this.currentScreen) {
      case AppState.LOBBY:
        if (this.lobbyScreen) {
          this.lobbyScreen.showError(message);
        }
        break;
      case AppState.IN_GAME:
        if (this.gameHUD) {
          this.gameHUD.showNotification(message, 'error');
        }
        break;
      default:
        console.error('UI Error:', message);
    }
  }

  clearErrors(): void {
    if (this.lobbyScreen) this.lobbyScreen.clearError();
    if (this.gameHUD) this.gameHUD.clearNotifications();
  }

  destroy(): void {
    // Clean up event listeners
    window.removeEventListener('resize', this.handleResize);
    document.removeEventListener('keydown', this.handleEscapeKey);
    
    // Destroy screen components
    if (this.lobbyScreen) {
      this.lobbyScreen.destroy();
      this.lobbyScreen = null;
    }
    
    if (this.gameHUD) {
      this.gameHUD.destroy();
      this.gameHUD = null;
    }
    
    if (this.loadingScreen) {
      this.loadingScreen.destroy();
      this.loadingScreen = null;
    }
    
    console.log('UIManager destroyed');
  }
}