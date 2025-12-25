import * as PIXI from 'pixi.js';
import { AbstractRenderer } from './base-renderer';
import type { RendererConfig } from '../types/renderer-types';
import type { Player, Wall, Projectile, ClientGameState, StaticGameData } from '../state';
import { gameState } from '../state';

// Camera viewport in backend units
const CAMERA_WIDTH_UNITS = 20;
const CAMERA_HEIGHT_UNITS = 20;
const WALL_COLOR = 0x8B4513;
const PROJECTILE_COLOR = 0xff0000;

export class PixiRenderer extends AbstractRenderer {
  private app: PIXI.Application | null = null;
  private worldContainer: PIXI.Container | null = null;
  private gameContainer: PIXI.Container | null = null;
  private staticContainer: PIXI.Container | null = null;
  private backgroundGrid: PIXI.Graphics | null = null;
  private gridLabels: PIXI.Container | null = null;
  private debugContainer: PIXI.Container | null = null;
  private debugText: PIXI.Text | null = null;
  
  private playerSprites: Map<number, PIXI.Graphics> = new Map();
  private wallSprites: Map<string, PIXI.Graphics> = new Map();
  private projectileSprites: Map<string, PIXI.Graphics> = new Map();
  
  private pixelsPerUnit: number = 40;

  async init(container: HTMLElement, config: RendererConfig): Promise<void> {
    this.validateContainer(container);
    this.validateConfig(config);
    
    this.container = container;
    this.config = config;
    this.pixelsPerUnit = config.pixelsPerUnit || 40;

    // Create PixiJS application
    this.app = new PIXI.Application();
    
    await this.app.init({
      width: config.width,
      height: config.height,
      backgroundColor: config.backgroundColor || 0x1a1a1a,
      antialias: config.antialias !== false
    });

    // Create container hierarchy
    this.createContainerHierarchy();
    this.setupBackground();
    this.setupDebugDisplay();
    
    // Add canvas to DOM
    container.appendChild(this.app.canvas);
    
    this.isInitialized = true;
    console.log('PixiJS renderer initialized');
  }

  render(gameState: ClientGameState, staticData: StaticGameData): void {
    this.validateInitialization();
    
    if (this.isPaused || !gameState) {
      return;
    }

    // Update camera to follow current player
    this.updateCamera(gameState);
    
    // Update grid based on camera position
    this.updateGrid();
    
    // Render game objects
    this.renderPlayers(gameState.players);
    this.renderProjectiles(gameState.projectiles);
    this.renderWalls(staticData.walls);
    
    // Update debug display
    this.updateDebugDisplay(gameState, staticData);
  }

  renderPlayers(players: { [key: number]: Player }): void {
    const currentPlayers = new Set(Object.keys(players).map(Number));

    // Remove sprites for players that no longer exist
    for (const [playerId, sprite] of this.playerSprites) {
      if (!currentPlayers.has(playerId)) {
        this.gameContainer!.removeChild(sprite);
        this.playerSprites.delete(playerId);
      }
    }

    // Update existing players and create new ones
    for (const [playerIdStr, player] of Object.entries(players)) {
      const playerId = Number(playerIdStr);
      let sprite = this.playerSprites.get(playerId);

      if (!sprite) {
        sprite = this.createPlayerSprite(player);
        sprite.zIndex = 100;
        this.playerSprites.set(playerId, sprite);
        this.gameContainer!.addChild(sprite);
      }

      this.updatePlayerSprite(sprite, player);
    }
  }

  renderWalls(walls: Wall[]): void {
    // Clear existing wall sprites
    this.wallSprites.forEach((sprite) => {
      this.staticContainer!.removeChild(sprite);
    });
    this.wallSprites.clear();

    // Create new wall sprites
    for (const wall of walls) {
      const sprite = this.createWallSprite(wall);
      this.wallSprites.set(wall.id, sprite);
      this.staticContainer!.addChild(sprite);
      sprite.zIndex = 10;
    }
  }

  renderProjectiles(projectiles: Projectile[]): void {
    const currentProjectiles = new Set(projectiles.map(p => p.ID));

    // Remove sprites for projectiles that no longer exist
    for (const [projectileId, sprite] of this.projectileSprites) {
      if (!currentProjectiles.has(projectileId)) {
        this.gameContainer!.removeChild(sprite);
        this.projectileSprites.delete(projectileId);
      }
    }

    // Update existing projectiles and create new ones
    for (const projectile of projectiles) {
      let sprite = this.projectileSprites.get(projectile.ID);

      if (!sprite) {
        sprite = this.createProjectileSprite(projectile);
        this.projectileSprites.set(projectile.ID, sprite);
        this.gameContainer!.addChild(sprite);
      }

      this.updateProjectileSprite(sprite, projectile);
    }
  }

  updateCamera(gameStateOrPlayer: ClientGameState | Player): void {
    if (!this.worldContainer) return;
    
    let targetPlayer: Player | null = null;
    
    if ('players' in gameStateOrPlayer) {
      // It's a ClientGameState
      const currentPlayerId = gameState.currentPlayerID;
      targetPlayer = currentPlayerId !== null ? gameStateOrPlayer.players[currentPlayerId] : null;

      // TODO: Remove this fallback once server sends EntityID on join room success.
      // currentPlayerID should be set from JoinRoomSuccess response, not guessed here.
      if (!targetPlayer) {
        const playerIds = Object.keys(gameStateOrPlayer.players).map(Number);
        if (playerIds.length > 0) {
          const firstPlayerId = playerIds[0];
          gameState.setCurrentPlayerID(firstPlayerId);
          targetPlayer = gameStateOrPlayer.players[firstPlayerId];
        }
      }
    } else {
      // It's a Player object
      targetPlayer = gameStateOrPlayer;
    }

    if (targetPlayer) {
      this.worldContainer.pivot.set(
        targetPlayer.Position.x * this.pixelsPerUnit,
        targetPlayer.Position.y * this.pixelsPerUnit
      );
    }
  }

  resize(width: number, height: number): void {
    if (this.app && this.config) {
      this.app.renderer.resize(width, height);
      this.config.width = width;
      this.config.height = height;
      
      if (this.worldContainer) {
        this.worldContainer.position.set(width / 2, height / 2);
      }
    }
  }

  destroy(): void {
    if (this.app) {
      this.app.destroy(true);
      this.app = null;
    }
    
    // Clear all sprite maps
    this.playerSprites.clear();
    this.wallSprites.clear();
    this.projectileSprites.clear();
    
    this.isDestroyed = true;
    this.isInitialized = false;
    console.log('PixiJS renderer destroyed');
  }

  // Private helper methods
  private createContainerHierarchy(): void {
    if (!this.app) return;

    // Create world container that will handle camera transforms
    this.worldContainer = new PIXI.Container();
    this.worldContainer.sortableChildren = true;

    // Create game containers
    this.gameContainer = new PIXI.Container();
    this.gameContainer.sortableChildren = true;
    this.staticContainer = new PIXI.Container();
    this.staticContainer.sortableChildren = true;

    // Add containers to hierarchy
    this.worldContainer.addChild(this.staticContainer);
    this.worldContainer.addChild(this.gameContainer);

    // Create UI containers
    this.backgroundGrid = new PIXI.Graphics();
    this.gridLabels = new PIXI.Container();
    this.gridLabels.sortableChildren = true;
    this.debugContainer = new PIXI.Container();

    // Set z-indices
    this.backgroundGrid.zIndex = 1;
    this.worldContainer.zIndex = 100;
    this.gridLabels.zIndex = 1000;
    this.debugContainer.zIndex = 2000;

    // Add to stage
    this.app.stage.sortableChildren = true;
    this.app.stage.addChild(this.backgroundGrid);
    this.app.stage.addChild(this.gridLabels);
    this.app.stage.addChild(this.worldContainer);
    this.app.stage.addChild(this.debugContainer);

    // Position world container at screen center
    this.worldContainer.position.set(this.config!.width / 2, this.config!.height / 2);
  }

  private setupBackground(): void {
    // Background setup - grid will be updated dynamically
  }

  private setupDebugDisplay(): void {
    if (!this.debugContainer) return;

    this.debugText = new PIXI.Text('', {
      fontFamily: 'Arial',
      fontSize: 14,
      fill: 0xffffff,
      align: 'left'
    });
    
    this.debugText.x = 10;
    this.debugText.y = 10;
    this.debugContainer.addChild(this.debugText);
  }

  private updateGrid(): void {
    if (!this.backgroundGrid || !this.worldContainer || !this.gridLabels) return;

    const gridColor = 0x666666;
    this.backgroundGrid.clear();
    this.gridLabels.removeChildren();

    this.backgroundGrid.setStrokeStyle({ width: 1, color: gridColor, alpha: 0.5 });

    // Calculate visible world bounds
    const topLeft = this.worldContainer.toLocal(new PIXI.Point(0, 0));
    const bottomRight = this.worldContainer.toLocal(
      new PIXI.Point(this.config!.width, this.config!.height)
    );

    const leftWorldUnit = Math.floor(topLeft.x / this.pixelsPerUnit) - 1;
    const rightWorldUnit = Math.ceil(bottomRight.x / this.pixelsPerUnit) + 1;
    const topWorldUnit = Math.floor(topLeft.y / this.pixelsPerUnit) - 1;
    const bottomWorldUnit = Math.ceil(bottomRight.y / this.pixelsPerUnit) + 1;

    // Draw vertical lines
    for (let worldX = leftWorldUnit; worldX <= rightWorldUnit; worldX++) {
      const worldPos = new PIXI.Point(worldX * this.pixelsPerUnit, topLeft.y);
      const screenPos = this.worldContainer.toGlobal(worldPos);

      if (screenPos.x >= 0 && screenPos.x <= this.config!.width) {
        this.backgroundGrid.moveTo(screenPos.x, 0);
        this.backgroundGrid.lineTo(screenPos.x, this.config!.height);

        const label = new PIXI.Text(`${worldX}`, {
          fontFamily: 'Arial',
          fontSize: 10,
          fill: 0xff0000,
          align: 'center'
        });
        label.x = screenPos.x - label.width / 2;
        label.y = 5;
        label.zIndex = 1000;
        this.gridLabels.addChild(label);
      }
    }

    // Draw horizontal lines
    for (let worldY = topWorldUnit; worldY <= bottomWorldUnit; worldY++) {
      const worldPos = new PIXI.Point(topLeft.x, worldY * this.pixelsPerUnit);
      const screenPos = this.worldContainer.toGlobal(worldPos);

      if (screenPos.y >= 0 && screenPos.y <= this.config!.height) {
        this.backgroundGrid.moveTo(0, screenPos.y);
        this.backgroundGrid.lineTo(this.config!.width, screenPos.y);

        const label = new PIXI.Text(`${worldY}`, {
          fontFamily: 'Arial',
          fontSize: 10,
          fill: 0xff0000,
          align: 'center'
        });
        label.x = 5;
        label.y = screenPos.y - label.height / 2;
        label.zIndex = 1000;
        this.gridLabels.addChild(label);
      }
    }

    this.backgroundGrid.stroke();
  }

  private createPlayerSprite(player: Player): PIXI.Graphics {
    const sprite = new PIXI.Graphics();
    const isCurrentPlayer = player.ID === gameState.currentPlayerID;
    const color = isCurrentPlayer ? 0x00ff00 : 0x0080ff;

    const collisionRadius = 0.5;
    const radiusPixels = collisionRadius * this.pixelsPerUnit;
    const triangleSize = radiusPixels * 1.2;

    // Draw triangle
    sprite.setStrokeStyle({ width: 1, color: color });
    sprite.moveTo(triangleSize, 0);
    sprite.lineTo(-triangleSize/2, -triangleSize * 0.866);
    sprite.lineTo(-triangleSize/2, triangleSize * 0.866);
    sprite.lineTo(triangleSize, 0);
    sprite.fill(color);

    // Direction indicator
    const indicatorSize = radiusPixels * 0.3;
    const indicatorDistance = triangleSize + indicatorSize * 0.5;
    
    sprite.setStrokeStyle({ width: 1, color: 0xff0000 });
    sprite.moveTo(indicatorDistance + indicatorSize, 0);
    sprite.lineTo(indicatorDistance - indicatorSize/2, -indicatorSize * 0.5);
    sprite.lineTo(indicatorDistance - indicatorSize/2, indicatorSize * 0.5);
    sprite.lineTo(indicatorDistance + indicatorSize, 0);
    sprite.fill(0xff0000);

    return sprite;
  }

  private updatePlayerSprite(sprite: PIXI.Graphics, player: Player): void {
    sprite.x = player.Position.x * this.pixelsPerUnit;
    sprite.y = player.Position.y * this.pixelsPerUnit;
    sprite.rotation = player.Direction;
    sprite.visible = true;
  }

  private createWallSprite(wall: Wall): PIXI.Graphics {
    const sprite = new PIXI.Graphics();
    const width = wall.half_size.x * 2 * this.pixelsPerUnit;
    const height = wall.half_size.y * 2 * this.pixelsPerUnit;

    sprite.rect(-width / 2, -height / 2, width, height);
    sprite.fill(WALL_COLOR);

    sprite.x = wall.center.x * this.pixelsPerUnit;
    sprite.y = wall.center.y * this.pixelsPerUnit;
    sprite.rotation = wall.rotation;
    sprite.visible = true;
    sprite.alpha = 1.0;

    return sprite;
  }

  private createProjectileSprite(_projectile: Projectile): PIXI.Graphics {
    const sprite = new PIXI.Graphics();
    const radius = 0.1 * this.pixelsPerUnit;
    
    sprite.circle(0, 0, radius);
    sprite.fill(PROJECTILE_COLOR);
    
    return sprite;
  }

  private updateProjectileSprite(sprite: PIXI.Graphics, projectile: Projectile): void {
    sprite.x = projectile.Position.x * this.pixelsPerUnit;
    sprite.y = projectile.Position.y * this.pixelsPerUnit;
  }

  private updateDebugDisplay(gameStateParam: ClientGameState, staticData: StaticGameData): void {
    if (!this.debugText) return;

    const debugInfo = [
      `Players: ${Object.keys(gameStateParam.players).length}`,
      `Walls: ${staticData.walls.length}`,
      `Projectiles: ${gameStateParam.projectiles?.length || 0}`,
      `Current Player: ${gameState.currentPlayerID || 'N/A'}`,
      `Renderer: PixiJS`,
      `FPS: ${Math.round(this.app?.ticker.FPS || 0)}`
    ];

    this.debugText.text = debugInfo.join('\n');
  }
}