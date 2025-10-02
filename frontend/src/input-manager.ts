import type { PlayerInput } from './state';
import type { NetworkClient } from './network/client';
import { gameState } from './state';

export class InputManager {
  private networkClient: NetworkClient;
  private keys: Set<string> = new Set();
  private currentInput: PlayerInput = {
    MoveUp: false,
    MoveDown: false,
    MoveLeft: false,
    MoveRight: false,
    RotateLeft: false,
    RotateRight: false,
    SwitchWeapon: false,
    Reload: false,
    FastReload: false,
    Fire: false,
    Timestamp: 0
  };

  private inputUpdateInterval: number | null = null;
  private isEnabled: boolean = false;

  constructor(networkClient: NetworkClient) {
    this.networkClient = networkClient;
    this.setupEventListeners();
    console.log('InputManager initialized');
  }

  private setupEventListeners(): void {
    document.addEventListener('keydown', (event) => {
      if (!this.isEnabled) return;
      
      // Prevent default behavior for game keys
      if (this.isGameKey(event.code)) {
        event.preventDefault();
      }
      
      this.keys.add(event.code.toLowerCase());
      this.updateInputState();
    });

    document.addEventListener('keyup', (event) => {
      if (!this.isEnabled) return;
      
      // Prevent default behavior for game keys
      if (this.isGameKey(event.code)) {
        event.preventDefault();
      }
      
      this.keys.delete(event.code.toLowerCase());
      this.updateInputState();
    });

    // Clear input when window loses focus
    window.addEventListener('focus', () => {
      this.keys.clear();
      this.updateInputState();
    });

    window.addEventListener('blur', () => {
      this.keys.clear();
      this.updateInputState();
    });
  }

  private isGameKey(code: string): boolean {
    const gameKeys = [
      'KeyW', 'KeyA', 'KeyS', 'KeyD', 
      'KeyQ', 'KeyE', 
      'Space'
    ];
    return gameKeys.includes(code);
  }

  private updateInputState(): void {
    if (!this.isEnabled) return;

    const newInput: PlayerInput = {
      MoveUp: this.keys.has('keyw'),
      MoveDown: this.keys.has('keys'),
      MoveLeft: this.keys.has('keya'),
      MoveRight: this.keys.has('keyd'),
      RotateLeft: this.keys.has('keyq'),
      RotateRight: this.keys.has('keye'),
      SwitchWeapon: false,
      Reload: false,
      FastReload: false,
      Fire: this.keys.has('space'),
      Timestamp: Date.now()
    };

    // Update debug info
    gameState.updateDebugInfo({
      keysPressed: Array.from(this.keys).join(', '),
      inputState: newInput
    });

    if (this.hasInputChanged(newInput)) {
      this.currentInput = newInput;
      this.sendInputToServer();
    } else {
      this.currentInput = newInput;
    }
  }

  private hasInputChanged(newInput: PlayerInput): boolean {
    return (
      this.currentInput.MoveUp !== newInput.MoveUp ||
      this.currentInput.MoveDown !== newInput.MoveDown ||
      this.currentInput.MoveLeft !== newInput.MoveLeft ||
      this.currentInput.MoveRight !== newInput.MoveRight ||
      this.currentInput.RotateLeft !== newInput.RotateLeft ||
      this.currentInput.RotateRight !== newInput.RotateRight ||
      this.currentInput.Fire !== newInput.Fire
    );
  }

  private hasActiveInput(): boolean {
    return (
      this.currentInput.MoveUp || this.currentInput.MoveDown ||
      this.currentInput.MoveLeft || this.currentInput.MoveRight ||
      this.currentInput.RotateLeft || this.currentInput.RotateRight ||
      this.currentInput.Fire
    );
  }

  private sendInputToServer(): void {
    if (!this.isEnabled || !this.networkClient.isConnected()) return;

    this.currentInput.Timestamp = Date.now();
    gameState.updateDebugInfo({
      lastInputSent: new Date().toLocaleTimeString()
    });
    
    this.networkClient.sendPlayerInput(this.currentInput);
  }

  private startInputLoop(): void {
    if (this.inputUpdateInterval) {
      clearInterval(this.inputUpdateInterval);
    }

    this.inputUpdateInterval = setInterval(() => {
      if (this.isEnabled && this.hasActiveInput()) {
        this.sendInputToServer();
      }
    }, 16); // ~60 FPS
  }

  private stopInputLoop(): void {
    if (this.inputUpdateInterval) {
      clearInterval(this.inputUpdateInterval);
      this.inputUpdateInterval = null;
    }
  }

  setEnabled(enabled: boolean): void {
    const wasEnabled = this.isEnabled;
    this.isEnabled = enabled;

    if (enabled && !wasEnabled) {
      // Enabling input
      this.startInputLoop();
      console.log('Input enabled');
    } else if (!enabled && wasEnabled) {
      // Disabling input
      this.stopInputLoop();
      this.keys.clear();
      this.updateInputState();
      console.log('Input disabled');
    }
  }

  isInputEnabled(): boolean {
    return this.isEnabled;
  }

  getCurrentInput(): PlayerInput {
    return { ...this.currentInput };
  }

  destroy(): void {
    this.setEnabled(false);
    
    // Remove event listeners - Note: we can't remove specific arrow function listeners
    // This is a limitation, but the listeners will check isEnabled anyway
    
    console.log('InputManager destroyed');
  }
}