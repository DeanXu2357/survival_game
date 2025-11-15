import type { 
  BaseRenderer, 
  RendererConfig
} from '../types/renderer-types';
import { RendererType } from '../types/renderer-types';
import type { Player, Wall, Projectile, ClientGameState, StaticGameData } from '../state';

export abstract class AbstractRenderer implements BaseRenderer {
  protected config: RendererConfig | null = null;
  protected container: HTMLElement | null = null;
  protected isInitialized: boolean = false;
  protected isPaused: boolean = false;
  protected isDestroyed: boolean = false;

  // Event callbacks
  public onPlayerClick?: (playerId: string) => void;
  public onWallClick?: (wallId: string) => void;

  // Abstract methods that must be implemented by concrete renderers
  abstract init(container: HTMLElement, config: RendererConfig): Promise<void>;
  abstract render(gameState: ClientGameState, staticData: StaticGameData): void;
  abstract renderPlayers(players: { [key: string]: Player }): void;
  abstract renderWalls(walls: Wall[]): void;
  abstract renderProjectiles(projectiles: Projectile[]): void;
  abstract updateCamera(targetPlayer: Player): void;
  abstract resize(width: number, height: number): void;
  abstract destroy(): void;

  // Common implementation for pause/resume
  pause(): void {
    this.isPaused = true;
  }

  resume(): void {
    this.isPaused = false;
  }

  // Utility methods
  protected validateInitialization(): void {
    if (!this.isInitialized) {
      throw new Error('Renderer not initialized. Call init() first.');
    }
    
    if (this.isDestroyed) {
      throw new Error('Renderer has been destroyed.');
    }
  }

  protected validateContainer(container: HTMLElement): void {
    if (!container) {
      throw new Error('Container element is required');
    }
    
    if (!document.body.contains(container)) {
      throw new Error('Container must be attached to the DOM');
    }
  }

  protected validateConfig(config: RendererConfig): void {
    if (!config) {
      throw new Error('Renderer config is required');
    }
    
    if (config.width <= 0 || config.height <= 0) {
      throw new Error('Renderer width and height must be positive');
    }
  }

  // Getters
  getType(): RendererType {
    return this.config?.type || RendererType.PIXI_2D;
  }

  isReady(): boolean {
    return this.isInitialized && !this.isDestroyed && !this.isPaused;
  }

  getContainer(): HTMLElement | null {
    return this.container;
  }

  getConfig(): RendererConfig | null {
    return this.config;
  }
}