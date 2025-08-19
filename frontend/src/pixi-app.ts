import * as PIXI from 'pixi.js';
import type { Player, Wall, Projectile, ClientGameState, StaticGameData } from './state';
import { gameState } from './state';

// Camera viewport in backend units
const CAMERA_WIDTH_UNITS = 20;
const CAMERA_HEIGHT_UNITS = 20;
// Pixel scale: how many pixels per backend unit
const PIXELS_PER_UNIT = 40;
// Canvas size in pixels
const CANVAS_WIDTH = CAMERA_WIDTH_UNITS * PIXELS_PER_UNIT;
const CANVAS_HEIGHT = CAMERA_HEIGHT_UNITS * PIXELS_PER_UNIT;

const WALL_COLOR = 0x8B4513; // Brown color
const PROJECTILE_COLOR = 0xff0000;

class GameRenderer {
  private app: PIXI.Application;
  private worldContainer: PIXI.Container; // Contains all world objects
  private gameContainer: PIXI.Container;
  private staticContainer: PIXI.Container;
  private backgroundGrid: PIXI.Graphics;
  private gridLabels: PIXI.Container;
  private debugContainer: PIXI.Container;
  private debugText: PIXI.Text;
  private playerSprites: Map<string, PIXI.Graphics> = new Map();
  private wallSprites: Map<string, PIXI.Graphics> = new Map();
  private projectileSprites: Map<string, PIXI.Graphics> = new Map();

  constructor() {
    this.app = new PIXI.Application();

    // Create world container that will handle camera transforms
    this.worldContainer = new PIXI.Container();
    this.worldContainer.sortableChildren = true;

    this.gameContainer = new PIXI.Container();
    this.gameContainer.sortableChildren = true; // Enable z-index sorting for dynamic objects
    this.staticContainer = new PIXI.Container();
    this.staticContainer.sortableChildren = true; // Enable z-index sorting for static objects

    // Add game containers to world container
    this.worldContainer.addChild(this.staticContainer);
    this.worldContainer.addChild(this.gameContainer);

    this.backgroundGrid = new PIXI.Graphics();
    this.gridLabels = new PIXI.Container();
    this.gridLabels.sortableChildren = true;
    this.debugContainer = new PIXI.Container();
    this.debugText = new PIXI.Text();
    this.setupApplication();
    this.setupStaticDataCallback();
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

    // Initialize static objects BEFORE adding containers to stage
    this.initializeStaticObjectsIfReady();

    this.app.stage.sortableChildren = true;

    // Grid and debug should be in screen space, not world space
    this.backgroundGrid.zIndex = 1;
    this.worldContainer.zIndex = 100;
    this.gridLabels.zIndex = 1000;
    this.debugContainer.zIndex = 2000;

    this.app.stage.addChild(this.backgroundGrid);
    this.app.stage.addChild(this.gridLabels);
    this.app.stage.addChild(this.worldContainer); // Add world container instead of individual containers
    this.app.stage.addChild(this.debugContainer);

    // Position world container at screen center for camera system
    this.worldContainer.position.set(CANVAS_WIDTH / 2, CANVAS_HEIGHT / 2);

    // Show grid to help judge wall size and positioning
    this.backgroundGrid.visible = true;
    this.gridLabels.visible = true;

    const appDiv = document.getElementById('app');
    if (appDiv) {
      appDiv.appendChild(this.app.canvas);
      this.createStageObjectsList(appDiv);
    } else {
      document.body.appendChild(this.app.canvas);
      this.createStageObjectsList(document.body);
    }

    this.app.ticker.add(() => this.render());
  }

  private initializeStaticObjectsIfReady(): void {
    // Check if static data already exists and initialize immediately
    const existingStaticData = gameState.getStaticData();
    if (existingStaticData) {
      console.log("Static data available during app init, initializing walls immediately");
      this.initializeStaticObjects(existingStaticData);
    }
  }

  private setupStaticDataCallback(): void {
    console.log("Setting up static data callback...");
    gameState.onStaticDataUpdate((staticData: StaticGameData) => {
      console.log("Static data callback triggered!", staticData);
      this.initializeStaticObjects(staticData);
    });

    // Check if static data already exists and initialize immediately
    const existingStaticData = gameState.getStaticData();
    if (existingStaticData) {
      console.log("Static data already exists, initializing immediately:", existingStaticData);
      this.initializeStaticObjects(existingStaticData);
    }
  }

  private initializeStaticObjects(staticData: StaticGameData): void {
    console.log("initializeStaticObjects called with:", staticData);
    console.log("Static data walls:", staticData.walls);

    // Clear existing wall sprites from staticContainer
    this.wallSprites.forEach((sprite, wallId) => {
      this.staticContainer.removeChild(sprite);
    });
    this.wallSprites.clear();

    // Clear static container (remove test rect)
    this.staticContainer.removeChildren();

    // Create walls and add to staticContainer
    if (staticData.walls && staticData.walls.length > 0) {
      for (const wall of staticData.walls) {
        const sprite = this.createWallSprite(wall);
        this.wallSprites.set(wall.id, sprite);
        this.staticContainer.addChild(sprite);
        sprite.zIndex = 10; // Ensure walls are behind players

        console.log(`Wall added to staticContainer: ${wall.id}, zIndex: ${sprite.zIndex}`);
        console.log(`StaticContainer children count: ${this.staticContainer.children.length}`);
      }
      console.log(`Initialized ${staticData.walls.length} walls in staticContainer`);
    }


    console.log('Static objects initialization complete');
  }

  private createFixedTestWall(): void {
    // Create the simplest possible wall at screen center
    const testWall = new PIXI.Graphics();
    testWall.name = 'fixedTestWall';

    // Small 50x50 pixel square at center of screen (camera at 400,300)
    testWall.rect(-25, -25, 50, 50);
    testWall.fill(0xff0000); // Bright red

    // Position at world coordinates where camera should see it
    testWall.x = 400; // Same as typical player spawn
    testWall.y = 300; // Same as typical player spawn
    testWall.zIndex = 50;

    this.gameContainer.addChild(testWall);

    console.log("Fixed test wall created at (400, 300) - should be visible at screen center");
  }

  private createDebugTestWall(): void {
    // Create a small, bright cyan wall that should be clearly visible
    const debugWall = new PIXI.Graphics();
    debugWall.name = 'debugWall';

    // Small 2x2 unit wall = 80x80 pixels - definitely fits in 15x15 unit camera view
    const width = 2 * PIXELS_PER_UNIT;  // 80 pixels
    const height = 2 * PIXELS_PER_UNIT; // 80 pixels

    debugWall.rect(-width / 2, -height / 2, width, height);
    debugWall.fill(0x00ffff); // Bright cyan

    // Position it near player spawn at world coordinates (not pixel coordinates)
    debugWall.x = 400; // World coordinate - near typical player spawn
    debugWall.y = 300; // World coordinate - near typical player spawn
    debugWall.zIndex = 5; // Between background and player

    this.gameContainer.addChild(debugWall);

    console.log(`DEBUG: Created test wall at world coordinates (${debugWall.x}, ${debugWall.y})`);
    console.log(`DEBUG: Test wall size: ${width}x${height} pixels`);
  }

  private createTestRectangle(): void {
    const testRect = new PIXI.Graphics();
    testRect.name = 'testRect';

    testRect.rect(-50, -50, 100, 100);
    testRect.fill(0xff00ff);

    testRect.x = 600 * PIXELS_PER_UNIT;
    testRect.y = 750 * PIXELS_PER_UNIT;

    this.staticContainer.addChild(testRect);
    // No caching for now - keep it simple
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

    // Calculate visible world bounds using worldContainer
    const topLeft = this.worldContainer.toLocal(new PIXI.Point(0, 0));
    const bottomRight = this.worldContainer.toLocal(new PIXI.Point(CANVAS_WIDTH, CANVAS_HEIGHT));

    // Convert to world units
    const leftWorldUnit = Math.floor(topLeft.x / PIXELS_PER_UNIT) - 1;
    const rightWorldUnit = Math.ceil(bottomRight.x / PIXELS_PER_UNIT) + 1;
    const topWorldUnit = Math.floor(topLeft.y / PIXELS_PER_UNIT) - 1;
    const bottomWorldUnit = Math.ceil(bottomRight.y / PIXELS_PER_UNIT) + 1;

    // Draw vertical lines
    for (let worldX = leftWorldUnit; worldX <= rightWorldUnit; worldX++) {
      const worldPos = new PIXI.Point(worldX * PIXELS_PER_UNIT, topLeft.y);
      const screenPos = this.worldContainer.toGlobal(worldPos);

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
        label.zIndex = 1000;
        this.gridLabels.addChild(label);
      }
    }

    // Draw horizontal lines
    for (let worldY = topWorldUnit; worldY <= bottomWorldUnit; worldY++) {
      const worldPos = new PIXI.Point(topLeft.x, worldY * PIXELS_PER_UNIT);
      const screenPos = this.worldContainer.toGlobal(worldPos);

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
        label.zIndex = 1000;
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

  private createStageObjectsList(parent: HTMLElement): void {
    // Check if elements already exist to avoid duplicates
    if (document.getElementById('game-objects-list') || document.getElementById('stage-objects-list')) {
      return; // Already created
    }

    // Create a div to show game objects list (positioned relative to right panel)
    const gameObjectsDiv = document.createElement('div');
    gameObjectsDiv.id = 'game-objects-list';
    gameObjectsDiv.style.position = 'fixed';
    gameObjectsDiv.style.top = '10px';
    gameObjectsDiv.style.right = '330px'; // 300px (right panel width) + 30px (gap)
    gameObjectsDiv.style.width = '350px';
    gameObjectsDiv.style.maxHeight = '400px';
    gameObjectsDiv.style.overflow = 'auto';
    gameObjectsDiv.style.backgroundColor = 'rgba(0, 0, 0, 0.8)';
    gameObjectsDiv.style.color = 'white';
    gameObjectsDiv.style.padding = '10px';
    gameObjectsDiv.style.fontFamily = 'monospace';
    gameObjectsDiv.style.fontSize = '11px';
    gameObjectsDiv.style.border = '1px solid #333';
    gameObjectsDiv.style.zIndex = '9999';

    // Create a div to show stage objects list (right side)
    const listDiv = document.createElement('div');
    listDiv.id = 'stage-objects-list';
    listDiv.style.position = 'fixed';
    listDiv.style.top = '10px';
    listDiv.style.right = '10px';
    listDiv.style.width = '300px';
    listDiv.style.maxHeight = '400px';
    listDiv.style.overflow = 'auto';
    listDiv.style.backgroundColor = 'rgba(0, 0, 0, 0.8)';
    listDiv.style.color = 'white';
    listDiv.style.padding = '10px';
    listDiv.style.fontFamily = 'monospace';
    listDiv.style.fontSize = '12px';
    listDiv.style.border = '1px solid #333';
    listDiv.style.zIndex = '9999';

    parent.appendChild(gameObjectsDiv);
    parent.appendChild(listDiv);

    // Update both lists periodically (only create one interval)
    setInterval(() => {
      this.updateGameObjectsList();
      this.updateStageObjectsList();
    }, 1000);
  }

  private updateGameObjectsList(): void {
    const gameObjectsDiv = document.getElementById('game-objects-list');
    if (!gameObjectsDiv) return;

    const state = gameState.getState();
    const staticData = gameState.getStaticData();
    const currentPlayerId = gameState.currentPlayerID;

    const objects: string[] = ['=== GAME OBJECTS ==='];
    objects.push('');

    // Static Objects (Walls)
    objects.push('📦 STATIC OBJECTS:');
    if (staticData && staticData.walls) {
      staticData.walls.forEach((wall, index) => {
        const distance = currentPlayerId && state?.players[currentPlayerId]
          ? this.calculateDistance(state.players[currentPlayerId].Position, wall.center)
          : 'N/A';

        objects.push(`  ${index + 1}. Wall [${wall.id}]`);
        objects.push(`     Pos: (${wall.center.x.toFixed(1)}, ${wall.center.y.toFixed(1)})`);
        objects.push(`     Size: ${(wall.half_size.x * 2).toFixed(1)}x${(wall.half_size.y * 2).toFixed(1)}`);
        objects.push(`     Rot: ${wall.rotation.toFixed(2)} rad`);
        objects.push(`     Distance: ${typeof distance === 'number' ? distance.toFixed(1) : distance} units`);

        const sprite = this.wallSprites.get(wall.id);
        objects.push(`     Sprite: ${sprite ? '✓' : '✗'} ${sprite ? `visible:${sprite.visible}` : ''}`);
        objects.push('');
      });
    } else {
      objects.push('     No walls');
      objects.push('');
    }

    // Dynamic Objects (Players - excluding current player)
    objects.push('🏃 OTHER PLAYERS:');
    if (state?.players) {
      const otherPlayers = Object.entries(state.players).filter(([id]) => id !== currentPlayerId);
      if (otherPlayers.length > 0) {
        otherPlayers.forEach(([id, player], index) => {
          const distance = currentPlayerId && state.players[currentPlayerId]
            ? this.calculateDistance(state.players[currentPlayerId].Position, player.Position)
            : 'N/A';

          objects.push(`  ${index + 1}. Player [${id}]`);
          objects.push(`     Pos: (${player.Position.x.toFixed(1)}, ${player.Position.y.toFixed(1)})`);
          objects.push(`     Dir: ${player.Direction.toFixed(2)} rad`);
          objects.push(`     Radius: ${player.Radius.toFixed(2)}`);
          objects.push(`     Distance: ${typeof distance === 'number' ? distance.toFixed(1) : distance} units`);

          const sprite = this.playerSprites.get(id);
          objects.push(`     Sprite: ${sprite ? '✓' : '✗'} ${sprite ? `visible:${sprite.visible}` : ''}`);
          objects.push('');
        });
      } else {
        objects.push('     No other players');
        objects.push('');
      }
    }

    // Dynamic Objects (Projectiles)
    objects.push('💥 PROJECTILES:');
    if (state?.projectiles && state.projectiles.length > 0) {
      state.projectiles.forEach((projectile, index) => {
        const distance = currentPlayerId && state.players[currentPlayerId]
          ? this.calculateDistance(state.players[currentPlayerId].Position, projectile.Position)
          : 'N/A';

        objects.push(`  ${index + 1}. Projectile [${projectile.ID}]`);
        objects.push(`     Pos: (${projectile.Position.x.toFixed(1)}, ${projectile.Position.y.toFixed(1)})`);
        objects.push(`     Vel: (${projectile.Velocity.x.toFixed(1)}, ${projectile.Velocity.y.toFixed(1)})`);
        objects.push(`     Distance: ${typeof distance === 'number' ? distance.toFixed(1) : distance} units`);

        const sprite = this.projectileSprites.get(projectile.ID);
        objects.push(`     Sprite: ${sprite ? '✓' : '✗'} ${sprite ? `visible:${sprite.visible}` : ''}`);
        objects.push('');
      });
    } else {
      objects.push('     No projectiles');
      objects.push('');
    }

    // Camera Info
    objects.push('📷 CAMERA INFO:');
    if (currentPlayerId && state?.players[currentPlayerId]) {
      const player = state.players[currentPlayerId];
      objects.push(`     Following: ${currentPlayerId}`);
      objects.push(`     Center: (${player.Position.x.toFixed(1)}, ${player.Position.y.toFixed(1)})`);
      objects.push(`     View: ${CAMERA_WIDTH_UNITS}x${CAMERA_HEIGHT_UNITS} units`);
      objects.push(`     Visible range: ±${(CAMERA_WIDTH_UNITS/2).toFixed(1)} units`);
    }

    gameObjectsDiv.innerHTML = objects.join('<br>');
  }

  private calculateDistance(pos1: {x: number, y: number}, pos2: {x: number, y: number}): number {
    const dx = pos1.x - pos2.x;
    const dy = pos1.y - pos2.y;
    return Math.sqrt(dx * dx + dy * dy);
  }

  private updateStageObjectsList(): void {
    const listDiv = document.getElementById('stage-objects-list');
    if (!listDiv) return;

    const objects: string[] = ['=== STAGE OBJECTS ==='];
    objects.push(`staticContainer pos: (${this.staticContainer.x}, ${this.staticContainer.y})`);
    objects.push(`gameContainer pos: (${this.gameContainer.x}, ${this.gameContainer.y})`);
    objects.push('');

    // Traverse stage children
    this.traverseContainer(this.app.stage, '', objects);

    objects.push('');
    objects.push(`=== WALL SPRITES MAP ===`);
    objects.push(`Size: ${this.wallSprites.size}`);
    this.wallSprites.forEach((sprite, id) => {
      objects.push(`${id}: pos(${sprite.x}, ${sprite.y}) visible:${sprite.visible} alpha:${sprite.alpha} zIndex:${sprite.zIndex}`);
    });

    listDiv.innerHTML = objects.join('<br>');
  }

  private traverseContainer(container: PIXI.Container, indent: string, objects: string[]): void {
    const name = (container as any).name || container.constructor.name;
    const visible = container.visible ? '✓' : '✗';
    const childCount = container.children.length;

    objects.push(`${indent}${visible} ${name} (${childCount} children) zIndex:${container.zIndex}`);

    container.children.forEach(child => {
      if (child instanceof PIXI.Container) {
        this.traverseContainer(child, indent + '  ', objects);
      } else {
        const childName = (child as any).name || (child as any).constructor.name;
        const childVisible = (child as any).visible ? '✓' : '✗';
        objects.push(`${indent}  ${childVisible} ${childName} zIndex:${(child as any).zIndex} pos:(${(child as any).x}, ${(child as any).y})`);
      }
    });
  }

  private render(): void {
    const state = gameState.getState();
    if (!state) {
      // No dynamic game state available
      return;
    }

    // Update camera to follow current player
    this.updateCamera(state);

    // Update grid based on camera position
    this.updateGrid();

    this.renderPlayers(state);
    this.renderProjectiles(state);
    this.renderDebugVisuals(state);
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
        sprite.zIndex = 100; // Ensure players are above walls
        this.playerSprites.set(playerId, sprite);
        this.gameContainer.addChild(sprite);
        console.log(`New player sprite created with zIndex: ${sprite.zIndex}`);
        console.log(`Player sprite added to gameContainer, total children: ${this.gameContainer.children.length}`);
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
    sprite.x = player.Position.x * PIXELS_PER_UNIT;
    sprite.y = player.Position.y * PIXELS_PER_UNIT;
    sprite.rotation = player.Direction;
    sprite.visible = true; // Let container transformation handle visibility
  }


  private createWallSprite(wall: Wall): PIXI.Graphics {
    const sprite = new PIXI.Graphics();

    // Convert backend units to pixels correctly
    const width = wall.half_size.x * 2 * PIXELS_PER_UNIT; // Convert backend units to pixels
    const height = wall.half_size.y * 2 * PIXELS_PER_UNIT; // Convert backend units to pixels

    console.log(`Creating REAL wall sprite: ${wall.id}`);
    console.log(`  Position: (${wall.center.x}, ${wall.center.y})`);
    console.log(`  Half-size: (${wall.half_size.x}, ${wall.half_size.y}) backend units`);
    console.log(`  Expected full size: ${wall.half_size.x * 2}x${wall.half_size.y * 2} backend units`);
    console.log(`  PIXELS_PER_UNIT: ${PIXELS_PER_UNIT}`);
    console.log(`  Calculated pixel size: ${width}x${height} pixels`);
    console.log(`  For comparison: player radius ${0.5} units = ${0.5 * PIXELS_PER_UNIT} pixels`);

    // Using modern PixiJS v8 API
    sprite.rect(-width / 2, -height / 2, width, height);
    sprite.fill(WALL_COLOR);

    // Position sprite at the wall's center in world coordinates (convert backend units to pixels)
    sprite.x = wall.center.x * PIXELS_PER_UNIT;
    sprite.y = wall.center.y * PIXELS_PER_UNIT;
    sprite.rotation = wall.rotation;

    // Ensure visibility
    sprite.visible = true;
    sprite.alpha = 1.0;
    sprite.zIndex = 110; // Set z-index here as well

    console.log(`Wall sprite created at world position: (${sprite.x}, ${sprite.y})`);
    console.log(`Wall sprite size: ${width}x${height} pixels`);
    console.log(`Wall sprite bounds after creation:`, sprite.getBounds());
    console.log(`Wall sprite visible: ${sprite.visible}, alpha: ${sprite.alpha}, zIndex: ${sprite.zIndex}`);

    // Force different properties to ensure visibility
    sprite.alpha = 1.0;
    sprite.visible = true;
    sprite.zIndex = 1; // Lower zIndex to be above background

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
    sprite.x = projectile.Position.x * PIXELS_PER_UNIT;
    sprite.y = projectile.Position.y * PIXELS_PER_UNIT;
  }

  private renderDebugVisuals(state: ClientGameState): void {
    // Remove existing debug visuals
    const existingCameraDebug = this.gameContainer.getChildByName('cameraDebug');
    if (existingCameraDebug) {
      this.gameContainer.removeChild(existingCameraDebug);
    }

    // Remove existing wall boundary boxes
    const childrenToRemove = this.gameContainer.children.filter(child =>
      (child as any).name && (child as any).name.startsWith('wallBoundary_')
    );
    childrenToRemove.forEach(child => this.gameContainer.removeChild(child));

    const cameraDebug = new PIXI.Graphics();
    cameraDebug.name = 'cameraDebug';

    const viewWidth = CAMERA_WIDTH_UNITS * PIXELS_PER_UNIT;
    const viewHeight = CAMERA_HEIGHT_UNITS * PIXELS_PER_UNIT;

    cameraDebug.setStrokeStyle({ width: 2, color: 0xffff00, alpha: 0.8 });
    cameraDebug.rect(-viewWidth / 2, -viewHeight / 2, viewWidth, viewHeight);
    cameraDebug.stroke();

    // Position at camera center (pivot point in world coordinates)
    cameraDebug.x = this.worldContainer.pivot.x;
    cameraDebug.y = this.worldContainer.pivot.y;
    cameraDebug.zIndex = 200; // Above everything

    this.gameContainer.addChild(cameraDebug);

    // Wall boundary boxes removed - they were being created every frame
  }

  private updateCamera(state: ClientGameState): void {
    // Center camera on current player
    const currentPlayerId = gameState.currentPlayerID;
    const currentPlayer = currentPlayerId ? state.players[currentPlayerId] : null;

    if (currentPlayer) {
      // Use PixiJS pivot system to center camera on player
      // Convert backend units to pixels for pivot
      this.worldContainer.pivot.set(
        currentPlayer.Position.x * PIXELS_PER_UNIT,
        currentPlayer.Position.y * PIXELS_PER_UNIT
      );
    } else {
      // Fallback: use first available player
      const playerEntries = Object.entries(state.players);
      if (playerEntries.length > 0) {
        const [firstPlayerId, firstPlayer] = playerEntries[0];
        gameState.setCurrentPlayerID(firstPlayerId);

        this.worldContainer.pivot.set(
          firstPlayer.Position.x * PIXELS_PER_UNIT,
          firstPlayer.Position.y * PIXELS_PER_UNIT
        );
      }
    }
  }


  private updateDebugDisplay(state: ClientGameState): void {
    const playerEntries = Object.entries(state.players);
    const walls = gameState.getWalls();
    const staticData = gameState.getStaticData();

    const staticObjectsCount = this.staticContainer.children.length;
    const wallSpritesCount = this.wallSprites.size;

    const debugInfo = [
      `Players: ${Object.keys(state.players).length}`,
      `Walls: ${walls.length} (sprites: ${wallSpritesCount}, static objects: ${staticObjectsCount})`,
      `Projectiles: ${state.projectiles?.length || 0}`,
      `Current Player: ${gameState.currentPlayerID}`,
      `Connection: ${gameState.isConnected() ? 'Connected' : 'Disconnected'}`,
      `Static Data: ${staticData ? 'Received' : 'Not received'}`,
      `Static Container Cache: ${this.staticContainer.cacheAsBitmap ? 'Enabled' : 'Disabled'}`,
      '',
      'Wall Details:',
    ];

    // Add wall information
    if (walls.length > 0) {
      walls.forEach((wall, index) => {
        const wallSprite = this.wallSprites.get(wall.id);
        debugInfo.push(`${index + 1}. ${wall.id}:`);
        debugInfo.push(`   Center: (${wall.center.x}, ${wall.center.y})`);
        debugInfo.push(`   Half-size: (${wall.half_size.x}, ${wall.half_size.y})`);
        debugInfo.push(`   Rotation: ${wall.rotation}`);
        debugInfo.push(`   Sprite: ${wallSprite ? 'EXISTS' : 'MISSING'}`);
        if (wallSprite) {
          debugInfo.push(`   Sprite zIndex: ${wallSprite.zIndex}`);
          debugInfo.push(`   Sprite visible: ${wallSprite.visible}`);
          debugInfo.push(`   Sprite alpha: ${wallSprite.alpha}`);
        }
        debugInfo.push(`   Bounds: x[${(wall.center.x - wall.half_size.x).toFixed(1)}, ${(wall.center.x + wall.half_size.x).toFixed(1)}] y[${(wall.center.y - wall.half_size.y).toFixed(1)}, ${(wall.center.y + wall.half_size.y).toFixed(1)}]`);
        debugInfo.push('');
      });
    } else {
      debugInfo.push('No walls found');
      debugInfo.push('');
    }

    // Add container children z-index info
    debugInfo.push('GameContainer Children:');
    debugInfo.push(`Total children: ${this.gameContainer.children.length}`);
    this.gameContainer.children.forEach((child, index) => {
      const name = (child as any).name || 'unnamed';
      debugInfo.push(`${index + 1}. ${name} (zIndex: ${child.zIndex})`);
    });
    debugInfo.push('');

    debugInfo.push('Camera Info:');
    debugInfo.push(`Camera pivot: (${(this.worldContainer.pivot.x / PIXELS_PER_UNIT).toFixed(1)}, ${(this.worldContainer.pivot.y / PIXELS_PER_UNIT).toFixed(1)}) units`);
    debugInfo.push(`World position: (${this.worldContainer.position.x}, ${this.worldContainer.position.y})`);
    debugInfo.push(`View: ${CAMERA_WIDTH_UNITS}x${CAMERA_HEIGHT_UNITS} units`);
    debugInfo.push(`Canvas: ${CANVAS_WIDTH}x${CANVAS_HEIGHT}px`);
    debugInfo.push(`Scale: ${PIXELS_PER_UNIT}px/unit`);
    debugInfo.push('');
    debugInfo.push('Player Details:');

    // Add detailed info for each player
    if (playerEntries.length === 0) {
      debugInfo.push('No players found!');
    } else {
      playerEntries.forEach(([id, player], index) => {
        const isCurrent = id === gameState.currentPlayerID;
        const scaledRadius = player.Radius * PIXELS_PER_UNIT;

        debugInfo.push(`${index + 1}. ${id}${isCurrent ? ' (YOU)' : ''}:`);
        debugInfo.push(`   World: (${player.Position.x.toFixed(2)}, ${player.Position.y.toFixed(2)})`);
        debugInfo.push(`   World Pixels: (${(player.Position.x * PIXELS_PER_UNIT).toFixed(1)}, ${(player.Position.y * PIXELS_PER_UNIT).toFixed(1)})`);
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
