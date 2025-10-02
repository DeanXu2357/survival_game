import type { Player, Wall, Projectile, ClientGameState, StaticGameData } from '../state';

export enum RendererType {
  PIXI_2D = 'pixi-2d',
  THREE_3D = 'three-3d',
  CANVAS_2D = 'canvas-2d',
  WEBGL = 'webgl'
}

export interface RendererConfig {
  type: RendererType;
  width: number;
  height: number;
  antialias?: boolean;
  backgroundColor?: number;
  pixelsPerUnit?: number;
}

export interface BaseRenderer {
  init(container: HTMLElement, config: RendererConfig): Promise<void>;
  render(gameState: ClientGameState, staticData: StaticGameData): void;
  resize(width: number, height: number): void;
  destroy(): void;
  pause(): void;
  resume(): void;
  getType(): RendererType;
  
  // Renderer-specific methods
  renderPlayers(players: { [key: string]: Player }): void;
  renderWalls(walls: Wall[]): void;
  renderProjectiles(projectiles: Projectile[]): void;
  updateCamera(targetPlayer: Player): void;
  
  // Event callbacks
  onPlayerClick?: (playerId: string) => void;
  onWallClick?: (wallId: string) => void;
}

export interface RendererFactory {
  createRenderer(type: RendererType): BaseRenderer;
  getSupportedTypes(): RendererType[];
}

export interface RenderManagerConfig {
  defaultRenderer: RendererType;
  fallbackRenderer: RendererType;
  container: HTMLElement;
  width: number;
  height: number;
}