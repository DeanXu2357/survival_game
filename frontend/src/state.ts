export interface Vector2D {
  x: number;
  y: number;
}

export interface Player {
  ID: number;
  Position: Vector2D;
  Direction: number;
}

export interface Wall {
  id: string;
  center: Vector2D;
  half_size: Vector2D;
  rotation: number;
}

export interface Projectile {
  ID: string;
  Position: Vector2D;
  Direction: Vector2D;  // Direction unit vector (matches backend)
  Speed: number;        // Pixels per second
  Range: number;        // Maximum travel distance
  Damage: number;       // Damage on hit
  OwnerID: string;
}

export interface ClientGameState {
  players: { [key: number]: Player };
  projectiles: Projectile[];
  timestamp: number;
}

export interface StaticGameData {
  walls: Wall[];
  mapWidth: number;
  mapHeight: number;
}

export interface PlayerInput {
  MoveUp: boolean;
  MoveDown: boolean;
  MoveLeft: boolean;
  MoveRight: boolean;
  RotateLeft: boolean;
  RotateRight: boolean;
  SwitchWeapon: boolean;
  Reload: boolean;
  FastReload: boolean;
  Fire: boolean;
  Timestamp: number;
}

export interface DebugInfo {
  keysPressed: string;
  inputState: PlayerInput | null;
  lastInputSent: string;
  connectionStatus: boolean;
}

export type StaticDataCallback = (staticData: StaticGameData) => void;

class GameState {
  private gameState: ClientGameState | null = null;
  private staticData: StaticGameData | null = null;
  public currentPlayerID: number | null = null;
  private currentSessionId: string | null = null;
  private staticDataCallbacks: StaticDataCallback[] = [];
  private debugInfo: DebugInfo = {
    keysPressed: '',
    inputState: null,
    lastInputSent: 'None',
    connectionStatus: false
  };

  updateState(newState: ClientGameState): void {
    this.gameState = newState;
  }

  getState(): ClientGameState | null {
    return this.gameState;
  }

  getCurrentPlayer(): Player | null {
    if (!this.gameState || this.currentPlayerID === null) {
      return null;
    }
    return this.gameState.players[this.currentPlayerID] || null;
  }

  setCurrentPlayerID(playerID: number): void {
    this.currentPlayerID = playerID;
  }

  updateDebugInfo(debugInfo: Partial<DebugInfo>): void {
    this.debugInfo = { ...this.debugInfo, ...debugInfo };
  }

  getDebugInfo(): DebugInfo {
    return this.debugInfo;
  }

  isConnected(): boolean {
    return this.debugInfo.connectionStatus;
  }

  setSessionId(sessionId: string): void {
    this.currentSessionId = sessionId;
  }

  getSessionId(): string | null {
    return this.currentSessionId;
  }

  clearSession(): void {
    this.currentSessionId = null;
  }

  updateStaticData(staticData: StaticGameData): void {
    this.staticData = staticData;
    // Trigger all callbacks when static data is updated
    this.staticDataCallbacks.forEach(callback => callback(staticData));
  }

  onStaticDataUpdate(callback: StaticDataCallback): void {
    this.staticDataCallbacks.push(callback);
    // If static data already exists, call the callback immediately
    if (this.staticData) {
      callback(this.staticData);
    }
  }

  removeStaticDataCallback(callback: StaticDataCallback): void {
    const index = this.staticDataCallbacks.indexOf(callback);
    if (index > -1) {
      this.staticDataCallbacks.splice(index, 1);
    }
  }

  getStaticData(): StaticGameData | null {
    return this.staticData;
  }

  getWalls(): Wall[] {
    return this.staticData?.walls || [];
  }

  getMapDimensions(): { width: number; height: number } {
    return {
      width: this.staticData?.mapWidth || 800,
      height: this.staticData?.mapHeight || 600
    };
  }
}

export const gameState = new GameState();