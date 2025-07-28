import * as PIXI from 'pixi.js';
import type { Player, Wall, Projectile, ClientGameState } from './state';
import { gameState } from './state';

// Camera viewport in backend units
const CAMERA_WIDTH_UNITS = 15;
const CAMERA_HEIGHT_UNITS = 15;
// Pixel scale: how many pixels per backend unit
const PIXELS_PER_UNIT = 40;
// Canvas size in pixels
const CANVAS_WIDTH = CAMERA_WIDTH_UNITS * PIXELS_PER_UNIT;
const CANVAS_HEIGHT = CAMERA_HEIGHT_UNITS * PIXELS_PER_UNIT;

const WALL_COLOR = 0x8b4513;
const PROJECTILE_COLOR = 0xff0000;

class GameRenderer {
  private app: PIXI.Application;
  private gameContainer: PIXI.Container;
  private backgroundGrid: PIXI.Graphics;
  private gridLabels: PIXI.Container;
  private debugContainer: PIXI.Container;
  private debugText: PIXI.Text;
  private playerSprites: Map<string, PIXI.Graphics> = new Map();
  private wallSprites: PIXI.Graphics[] = [];
  private projectileSprites: Map<string, PIXI.Graphics> = new Map();
  
  // Camera system
  private cameraX: number = 0;
  private cameraY: number = 0;

  constructor() {
    this.app = new PIXI.Application();
    this.gameContainer = new PIXI.Container();
    this.backgroundGrid = new PIXI.Graphics();
    this.gridLabels = new PIXI.Container();
    this.debugContainer = new PIXI.Container();
    this.debugText = new PIXI.Text();
    this.setupApplication();
  }

  private async setupApplication(): Promise<void> {
    await this.app.init({
      width: CANVAS_WIDTH,
      height: CANVAS_HEIGHT,
      backgroundColor: 0x1a1a1a,
      antialias: true
    });

    this.createBackgroundGrid();
    this.setupDebugDisplay();
    
    // Grid should be in screen space, not world space
    this.app.stage.addChild(this.backgroundGrid);
    this.app.stage.addChild(this.gridLabels);
    this.app.stage.addChild(this.gameContainer);
    this.app.stage.addChild(this.debugContainer);
    document.body.appendChild(this.app.canvas);

    this.app.ticker.add(() => this.render());
  }

  private createBackgroundGrid(): void {
    // This will be updated dynamically in updateGrid()
  }

  private updateGrid(): void {
    const gridColor = 0x666666;
    this.backgroundGrid.clear();
    this.gridLabels.removeChildren();

    // Draw grid lines that align with world units
    this.backgroundGrid.setStrokeStyle({ width: 1, color: gridColor, alpha: 0.5 });
    
    // Calculate visible world bounds
    const topLeft = this.gameContainer.toLocal(new PIXI.Point(0, 0));
    const bottomRight = this.gameContainer.toLocal(new PIXI.Point(CANVAS_WIDTH, CANVAS_HEIGHT));
    
    // Convert to world units
    const leftWorldUnit = Math.floor(topLeft.x / PIXELS_PER_UNIT) - 1;
    const rightWorldUnit = Math.ceil(bottomRight.x / PIXELS_PER_UNIT) + 1;
    const topWorldUnit = Math.floor(topLeft.y / PIXELS_PER_UNIT) - 1;
    const bottomWorldUnit = Math.ceil(bottomRight.y / PIXELS_PER_UNIT) + 1;
    
    // Draw vertical lines
    for (let worldX = leftWorldUnit; worldX <= rightWorldUnit; worldX++) {
      const worldPos = new PIXI.Point(worldX * PIXELS_PER_UNIT, topLeft.y);
      const screenPos = this.gameContainer.toGlobal(worldPos);
      
      if (screenPos.x >= 0 && screenPos.x <= CANVAS_WIDTH) {
        this.backgroundGrid.moveTo(screenPos.x, 0);
        this.backgroundGrid.lineTo(screenPos.x, CANVAS_HEIGHT);
        
        // Add coordinate label
        const label = new PIXI.Text(`${worldX}`, {
          fontFamily: 'Arial',
          fontSize: 10,
          fill: 0xff0000,
          align: 'center'
        });
        label.x = screenPos.x - label.width / 2;
        label.y = 5;
        this.gridLabels.addChild(label);
      }
    }
    
    // Draw horizontal lines  
    for (let worldY = topWorldUnit; worldY <= bottomWorldUnit; worldY++) {
      const worldPos = new PIXI.Point(topLeft.x, worldY * PIXELS_PER_UNIT);
      const screenPos = this.gameContainer.toGlobal(worldPos);
      
      if (screenPos.y >= 0 && screenPos.y <= CANVAS_HEIGHT) {
        this.backgroundGrid.moveTo(0, screenPos.y);
        this.backgroundGrid.lineTo(CANVAS_WIDTH, screenPos.y);
        
        // Add coordinate label
        const label = new PIXI.Text(`${worldY}`, {
          fontFamily: 'Arial',
          fontSize: 10,
          fill: 0xff0000,
          align: 'center'
        });
        label.x = 5;
        label.y = screenPos.y - label.height / 2;
        this.gridLabels.addChild(label);
      }
    }
    
    // Apply stroke
    this.backgroundGrid.stroke();
  }

  private setupDebugDisplay(): void {
    this.debugText.style = {
      fontFamily: 'Arial',
      fontSize: 14,
      fill: 0xffffff,
      align: 'left'
    };
    this.debugText.x = 10;
    this.debugText.y = 10;
    this.debugContainer.addChild(this.debugText);
  }

  private render(): void {
    const state = gameState.getState();
    if (!state) {
        // No game state available
      return;
    }

    // Update camera to follow current player
    this.updateCamera(state);
    
    // Update grid based on camera position
    this.updateGrid();

    this.renderPlayers(state);
    this.renderWalls(state);
    this.renderProjectiles(state);
    this.updateDebugDisplay(state);
  }

  private renderPlayers(state: ClientGameState): void {
    const currentPlayers = new Set(Object.keys(state.players));
    
    // Players being rendered silently
    
    for (const [playerId, sprite] of this.playerSprites) {
      if (!currentPlayers.has(playerId)) {
        this.gameContainer.removeChild(sprite);
        this.playerSprites.delete(playerId);
        // Player sprite removed
      }
    }

    for (const [playerId, player] of Object.entries(state.players)) {
      let sprite = this.playerSprites.get(playerId);
      
      if (!sprite) {
        sprite = this.createPlayerSprite(player);
        this.playerSprites.set(playerId, sprite);
        this.gameContainer.addChild(sprite);
        // New player sprite created
      }
      
      this.updatePlayerSprite(sprite, player);
    }
  }

  private createPlayerSprite(player: Player): PIXI.Graphics {
    const sprite = new PIXI.Graphics();
    
    const isCurrentPlayer = player.ID === gameState.currentPlayerID;
    const color = isCurrentPlayer ? 0x00ff00 : 0x0080ff;
    
    // Player collision: diameter 1.0 unit (radius 0.5 unit)
    // Triangle should circumscribe this circle
    const collisionRadius = 0.5; // Backend units
    const radiusPixels = collisionRadius * PIXELS_PER_UNIT; // Convert to pixels
    
    // For equilateral triangle circumscribing a circle:
    // Triangle height = radius * 2 * sqrt(3) / 3 * 3/2 = radius * sqrt(3)
    // But for better visibility, we use a slightly larger triangle
    const triangleSize = radiusPixels * 1.2; // Slightly larger than circumscribed
    
    // Draw main triangle shape (pointing forward, direction 0)
    sprite.setStrokeStyle({ width: 1, color: color });
    
    sprite.moveTo(triangleSize, 0);                         // Front point (direction 0)
    sprite.lineTo(-triangleSize/2, -triangleSize * 0.866);  // Back left
    sprite.lineTo(-triangleSize/2, triangleSize * 0.866);   // Back right
    sprite.lineTo(triangleSize, 0);                         // Close triangle
    sprite.fill(color);
    
    // Draw direction indicator (small red triangle in front)
    const indicatorSize = radiusPixels * 0.3; // Small triangle
    const indicatorDistance = triangleSize + indicatorSize * 0.5; // Distance from center
    
    sprite.setStrokeStyle({ width: 1, color: 0xff0000 });
    
    sprite.moveTo(indicatorDistance + indicatorSize, 0);           // Front point of indicator
    sprite.lineTo(indicatorDistance - indicatorSize/2, -indicatorSize * 0.5); // Back left
    sprite.lineTo(indicatorDistance - indicatorSize/2, indicatorSize * 0.5);  // Back right
    sprite.lineTo(indicatorDistance + indicatorSize, 0);           // Close indicator triangle
    sprite.fill(0xff0000);
    
    // Optional: Draw collision circle for debugging (comment out for production)
    // sprite.circle(0, 0, radiusPixels);
    // sprite.stroke({ width: 1, color: 0xff0000, alpha: 0.3 });
    
    return sprite;
  }

  private updatePlayerSprite(sprite: PIXI.Graphics, player: Player): void {
    // Keep sprites in world coordinates
    sprite.x = player.Position.X * PIXELS_PER_UNIT;
    sprite.y = player.Position.Y * PIXELS_PER_UNIT;
    sprite.rotation = player.Direction;
    sprite.visible = true; // Let container transformation handle visibility
  }

  private renderWalls(state: ClientGameState): void {
    for (const sprite of this.wallSprites) {
      this.gameContainer.removeChild(sprite);
    }
    this.wallSprites = [];

    for (const wall of state.walls) {
      const sprite = this.createWallSprite(wall);
      this.wallSprites.push(sprite);
      this.gameContainer.addChild(sprite);
    }
  }

  private createWallSprite(wall: Wall): PIXI.Graphics {
    const sprite = new PIXI.Graphics();
    
    // Keep in world coordinates
    const worldX = wall.Position.X * PIXELS_PER_UNIT;
    const worldY = wall.Position.Y * PIXELS_PER_UNIT;
    const worldWidth = wall.Width * PIXELS_PER_UNIT;
    const worldHeight = wall.Height * PIXELS_PER_UNIT;
    
    sprite.rect(worldX, worldY, worldWidth, worldHeight);
    sprite.fill(WALL_COLOR);
    return sprite;
  }

  private renderProjectiles(state: ClientGameState): void {
    const currentProjectiles = new Set(state.projectiles.map(p => p.ID));
    
    for (const [projectileId, sprite] of this.projectileSprites) {
      if (!currentProjectiles.has(projectileId)) {
        this.gameContainer.removeChild(sprite);
        this.projectileSprites.delete(projectileId);
      }
    }

    for (const projectile of state.projectiles) {
      let sprite = this.projectileSprites.get(projectile.ID);
      
      if (!sprite) {
        sprite = this.createProjectileSprite(projectile);
        this.projectileSprites.set(projectile.ID, sprite);
        this.gameContainer.addChild(sprite);
      }
      
      this.updateProjectileSprite(sprite, projectile);
    }
  }

  private createProjectileSprite(_projectile: Projectile): PIXI.Graphics {
    const sprite = new PIXI.Graphics();
    // Scale projectile size by pixels per unit
    const radius = 0.1 * PIXELS_PER_UNIT; // 0.1 backend units radius
    sprite.circle(0, 0, radius);
    sprite.fill(PROJECTILE_COLOR);
    return sprite;
  }

  private updateProjectileSprite(sprite: PIXI.Graphics, projectile: Projectile): void {
    // Keep in world coordinates
    sprite.x = projectile.Position.X * PIXELS_PER_UNIT;
    sprite.y = projectile.Position.Y * PIXELS_PER_UNIT;
  }

  private updateCamera(state: ClientGameState): void {
    // Center camera on current player
    const currentPlayer = state.players[gameState.currentPlayerID];
    
    if (currentPlayer) {
      this.cameraX = currentPlayer.Position.X;
      this.cameraY = currentPlayer.Position.Y;
      
      // Transform the game container to center the camera
      const newContainerX = CANVAS_WIDTH / 2 - this.cameraX * PIXELS_PER_UNIT;
      const newContainerY = CANVAS_HEIGHT / 2 - this.cameraY * PIXELS_PER_UNIT;
      
      this.gameContainer.x = newContainerX;
      this.gameContainer.y = newContainerY;
    } else {
      // Fallback: use first available player
      const playerEntries = Object.entries(state.players);
      if (playerEntries.length > 0) {
        const [firstPlayerId, firstPlayer] = playerEntries[0];
        gameState.setCurrentPlayerID(firstPlayerId);
        
        this.cameraX = firstPlayer.Position.X;
        this.cameraY = firstPlayer.Position.Y;
        
        const newContainerX = CANVAS_WIDTH / 2 - this.cameraX * PIXELS_PER_UNIT;
        const newContainerY = CANVAS_HEIGHT / 2 - this.cameraY * PIXELS_PER_UNIT;
        
        this.gameContainer.x = newContainerX;
        this.gameContainer.y = newContainerY;
      }
    }
  }

  private worldToScreen(worldX: number, worldY: number): { x: number, y: number } {
    return {
      x: (worldX - this.cameraX) * PIXELS_PER_UNIT + CANVAS_WIDTH / 2,
      y: (worldY - this.cameraY) * PIXELS_PER_UNIT + CANVAS_HEIGHT / 2
    };
  }

  private updateDebugDisplay(state: ClientGameState): void {
    const currentPlayer = state.players[gameState.currentPlayerID];
    const playerEntries = Object.entries(state.players);
    
    const debugInfo = [
      `Players: ${Object.keys(state.players).length}`,
      `Walls: ${state.walls?.length || 0}`,
      `Projectiles: ${state.projectiles?.length || 0}`,
      `Current Player: ${gameState.currentPlayerID}`,
      `Connection: ${gameState.isConnected() ? 'Connected' : 'Disconnected'}`,
      '',
      'Camera Info:',
      `Camera: (${this.cameraX.toFixed(1)}, ${this.cameraY.toFixed(1)})`,
      `View: ${CAMERA_WIDTH_UNITS}x${CAMERA_HEIGHT_UNITS} units`,
      `Canvas: ${CANVAS_WIDTH}x${CANVAS_HEIGHT}px`,
      `Scale: ${PIXELS_PER_UNIT}px/unit`,
      '',
      'Player Details:',
    ];

    // Add detailed info for each player
    if (playerEntries.length === 0) {
      debugInfo.push('No players found!');
    } else {
      playerEntries.forEach(([id, player], index) => {
        const isCurrent = id === gameState.currentPlayerID;
        const screenPos = this.worldToScreen(player.Position.X, player.Position.Y);
        const scaledRadius = player.Radius * PIXELS_PER_UNIT;
        
        debugInfo.push(`${index + 1}. ${id}${isCurrent ? ' (YOU)' : ''}:`);
        debugInfo.push(`   World: (${player.Position.X.toFixed(2)}, ${player.Position.Y.toFixed(2)})`);
        debugInfo.push(`   World Pixels: (${(player.Position.X * PIXELS_PER_UNIT).toFixed(1)}, ${(player.Position.Y * PIXELS_PER_UNIT).toFixed(1)})`);
        debugInfo.push(`   Radius: ${player.Radius.toFixed(2)} units (${scaledRadius.toFixed(1)}px)`);
        debugInfo.push(`   Direction: ${player.Direction.toFixed(2)} rad`);
        const sprite = this.playerSprites.get(id);
        const spriteInfo = sprite 
          ? `EXISTS (${sprite.x.toFixed(1)}, ${sprite.y.toFixed(1)}) visible:${sprite.visible}`
          : 'MISSING';
        debugInfo.push(`   Sprite: ${spriteInfo}`);
        debugInfo.push(`   Container: (${this.gameContainer.x.toFixed(1)}, ${this.gameContainer.y.toFixed(1)})`);
        
        // Calculate where sprite should appear on screen
        const finalScreenX = sprite ? sprite.x + this.gameContainer.x : 'N/A';
        const finalScreenY = sprite ? sprite.y + this.gameContainer.y : 'N/A';
        debugInfo.push(`   Final Screen: (${finalScreenX}, ${finalScreenY})`);
        
        if (index < playerEntries.length - 1) debugInfo.push('');
      });
    }

    debugInfo.push('');
    debugInfo.push('Input Debug:');
    debugInfo.push(`Keys: ${gameState.getDebugInfo().keysPressed}`);
    debugInfo.push(`State: ${JSON.stringify(gameState.getDebugInfo().inputState)}`);
    debugInfo.push(`Last: ${gameState.getDebugInfo().lastInputSent}`);

    this.debugText.text = debugInfo.join('\n');
  }

  public destroy(): void {
    this.app.destroy(true);
  }
}

let renderer: GameRenderer | null = null;

export function createApp(): void {
  if (!renderer) {
    renderer = new GameRenderer();
  }
}

export function destroyApp(): void {
  if (renderer) {
    renderer.destroy();
    renderer = null;
  }
}