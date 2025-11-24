console.log('====================================');
console.log('main.ts is loading...');
console.log('====================================');

import './style.css';
import { AppStateManager } from './managers/app-state';
import { UIManager } from './managers/ui-manager';
import { RenderManager } from './managers/render-manager';
import { NetworkClient } from './network/client';
import { InputManager } from './input-manager';
import { RendererType } from './types/renderer-types';
import { AppState } from './types/app-types';

console.log('====================================');
console.log('All imports loaded successfully');
console.log('====================================');

class Application {
  private appState: AppStateManager;
  private uiManager: UIManager;
  private renderManager: RenderManager;
  private networkClient: NetworkClient;
  private inputManager: InputManager | null = null;

  constructor() {
    console.log('Application starting...');
    
    // Initialize core managers
    this.appState = new AppStateManager();
    this.networkClient = new NetworkClient(this.appState);
    
    // Initialize UI manager
    this.uiManager = new UIManager(this.appState, {
      appContainer: document.getElementById('app')!,
      screens: {
        lobby: document.getElementById('lobby-screen')!,
        game: document.getElementById('game-screen')!,
        loading: document.getElementById('loading-screen')!
      }
    });

    // Initialize render manager
    const gameScreen = document.getElementById('game-screen')!;
    const canvasContainer = gameScreen.querySelector('#game-canvas-container') as HTMLElement;
    
    this.renderManager = new RenderManager({
      defaultRenderer: RendererType.PIXI_2D,
      fallbackRenderer: RendererType.PIXI_2D,
      container: canvasContainer,
      width: 800,
      height: 600
    });

    this.setupApplicationFlow();
  }

  private setupApplicationFlow(): void {
    // Listen for state changes to manage render lifecycle
    this.appState.onStateChange((newState: AppState) => {
      this.handleApplicationStateChange(newState);
    });

    // Setup window event handlers
    window.addEventListener('resize', () => this.handleWindowResize());
    window.addEventListener('beforeunload', () => this.destroy());

    console.log('Application initialized');
  }

  private async handleApplicationStateChange(newState: AppState): Promise<void> {
    console.log('Application: State changed to', newState);

    switch (newState) {
      case AppState.CONNECTING:
        // Stop any existing game rendering
        if (this.renderManager.isRendererActive()) {
          this.renderManager.stopRenderLoop();
        }
        
        // Disable input
        if (this.inputManager) {
          this.inputManager.setEnabled(false);
        }
        break;

      case AppState.LOBBY:
        // Stop rendering and input when in lobby
        if (this.renderManager.isRendererActive()) {
          this.renderManager.stopRenderLoop();
        }
        
        if (this.inputManager) {
          this.inputManager.setEnabled(false);
        }
        break;

      case AppState.JOINING:
        // Keep everything stopped while joining
        break;

      case AppState.IN_GAME:
        // Initialize game rendering and input
        await this.startGameMode();
        break;

      case AppState.DISCONNECTED:
        // Stop everything
        if (this.renderManager.isRendererActive()) {
          this.renderManager.stopRenderLoop();
        }
        
        if (this.inputManager) {
          this.inputManager.setEnabled(false);
        }
        break;
    }
  }

  private async startGameMode(): Promise<void> {
    try {
      console.log('Starting game mode...');

      // Initialize renderer if not already done
      if (!this.renderManager.getCurrentRenderer()) {
        await this.renderManager.initializeRenderer();
        console.log('Game renderer initialized');
      }

      // Start render loop
      this.renderManager.startRenderLoop();
      console.log('Render loop started');

      // Initialize input manager if not already done
      if (!this.inputManager) {
        this.inputManager = new InputManager(this.networkClient);
        console.log('Input manager initialized');
      }

      // Enable input
      this.inputManager.setEnabled(true);
      console.log('Game input enabled');

      console.log('Game mode started successfully');
      console.log('Controls: WASD to move, QE to rotate, Space to fire, ESC to leave room');

    } catch (error) {
      console.error('Failed to start game mode:', error);
      this.appState.handleError('Failed to initialize game: ' + (error as Error).message);
    }
  }

  private handleWindowResize(): void {
    const width = window.innerWidth;
    const height = window.innerHeight;
    
    // Update render manager size
    this.renderManager.resizeRenderer(width, height);
  }

  async start(): Promise<void> {
    try {
      console.log('Application starting connection...');
      
      // Start the connection process
      this.appState.requestConnect();
      
      console.log('Application startup complete');
    } catch (error) {
      console.error('Failed to start application:', error);
      this.appState.handleError('Application startup failed: ' + (error as Error).message);
    }
  }

  destroy(): void {
    console.log('Application shutting down...');

    // Destroy managers in reverse order
    if (this.inputManager) {
      this.inputManager.destroy();
    }

    if (this.renderManager) {
      this.renderManager.destroyRenderer();
    }

    if (this.uiManager) {
      this.uiManager.destroy();
    }

    if (this.networkClient) {
      this.networkClient.destroy();
    }

    // Remove window event listeners
    window.removeEventListener('resize', this.handleWindowResize);
    window.removeEventListener('beforeunload', this.destroy);

    console.log('Application shutdown complete');
  }
}

// Application entry point
async function initializeApplication(): Promise<void> {
  try {
    const app = new Application();
    await app.start();
    
    // Store app instance globally for debugging
    (window as any).app = app;
    
  } catch (error) {
    console.error('Failed to initialize application:', error);
    
    // Show error to user
    const appDiv = document.getElementById('app');
    if (appDiv) {
      appDiv.innerHTML = `
        <div class="startup-error">
          <h1>Failed to Start Game</h1>
          <p>Error: ${error instanceof Error ? error.message : 'Unknown error'}</p>
          <button onclick="window.location.reload()">Reload Page</button>
        </div>
      `;
    }
  }
}

// Start the application
initializeApplication();
