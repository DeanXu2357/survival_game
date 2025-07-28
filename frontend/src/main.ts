import './style.css';
import { connectToServer } from './network';
import { createApp } from './pixi-app';
import './input';

async function initializeGame(): Promise<void> {
  console.log('Game client started');
  
  try {
    createApp();
    console.log('PixiJS app initialized');
    
    await connectToServer();
    console.log('Connected to server');
    
    console.log('Game initialization complete');
    console.log('Controls: WASD to move, QE to rotate, Space to fire');
  } catch (error) {
    console.error('Failed to initialize game:', error);
  }
}

initializeGame();
