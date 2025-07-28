export interface Vector2D {
  X: number;
  Y: number;
}

export interface Player {
  ID: string;
  Position: Vector2D;
  Direction: number;
  Radius: number;
  RotationSpeed: number;
  MovementSpeed: number;
  Health: number;
  IsAlive: boolean;
}

export interface Wall {
  ID: string;
  Position: Vector2D;
  Width: number;
  Height: number;
}

export interface Projectile {
  ID: string;
  Position: Vector2D;
  Velocity: Vector2D;
  OwnerID: string;
}

export interface ClientGameState {
  players: { [key: string]: Player };
  walls: Wall[];
  projectiles: Projectile[];
  timestamp: number;
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
}

export interface DebugInfo {
  keysPressed: string;
  inputState: PlayerInput | null;
  lastInputSent: string;
  connectionStatus: boolean;
}

class GameState {
  private gameState: ClientGameState | null = null;
  public currentPlayerID: string | null = null;
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
    if (!this.gameState || !this.currentPlayerID) {
      return null;
    }
    return this.gameState.players[this.currentPlayerID] || null;
  }

  setCurrentPlayerID(playerID: string): void {
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
}

export const gameState = new GameState();