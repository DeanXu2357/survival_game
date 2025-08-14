import type { PlayerInput } from './state';
import { sendPlayerInput, isConnected } from './network';
import { gameState } from './state';

class InputManager {
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
    Fire: false
  };

  private inputUpdateInterval: number | null = null;

  constructor() {
    this.setupEventListeners();
    this.startInputLoop();
  }

  private setupEventListeners(): void {
    document.addEventListener('keydown', (event) => {
      this.keys.add(event.code.toLowerCase());
      this.updateInputState();
    });

    document.addEventListener('keyup', (event) => {
      this.keys.delete(event.code.toLowerCase());
      this.updateInputState();
    });

    window.addEventListener('focus', () => {
      this.keys.clear();
      this.updateInputState();
    });

    window.addEventListener('blur', () => {
      this.keys.clear();
      this.updateInputState();
    });
  }

  private updateInputState(): void {
    const newInput: PlayerInput = {
      MoveUp: this.keys.has('keys'),
      MoveDown: this.keys.has('keyw'),
      MoveLeft: this.keys.has('keya'),
      MoveRight: this.keys.has('keyd'),
      RotateLeft: this.keys.has('keyq'),
      RotateRight: this.keys.has('keye'),
      SwitchWeapon: false,
      Reload: false,
      FastReload: false,
      Fire: this.keys.has('space')
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
    if (!isConnected()) return;

    gameState.updateDebugInfo({
      lastInputSent: new Date().toLocaleTimeString()
    });
    sendPlayerInput(this.currentInput);
  }

  private startInputLoop(): void {
    this.inputUpdateInterval = setInterval(() => {
      if (this.hasActiveInput()) {
        this.sendInputToServer();
      }
    }, 16); // ~60 FPS
  }

  public destroy(): void {
    if (this.inputUpdateInterval) {
      clearInterval(this.inputUpdateInterval);
      this.inputUpdateInterval = null;
    }

    document.removeEventListener('keydown', this.setupEventListeners);
    document.removeEventListener('keyup', this.setupEventListeners);
    window.removeEventListener('focus', this.setupEventListeners);
    window.removeEventListener('blur', this.setupEventListeners);
  }

  public getCurrentInput(): PlayerInput {
    return { ...this.currentInput };
  }
}

export const inputManager = new InputManager();
