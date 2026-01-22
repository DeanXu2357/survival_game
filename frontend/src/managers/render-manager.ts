import type { 
  BaseRenderer, 
  RenderManagerConfig, 
  RendererType, 
  RendererConfig 
} from '../types/renderer-types';
import type { ClientGameState, StaticGameData } from '../state';
import { rendererFactory } from '../renderers/renderer-factory';
import { gameState } from '../state';

export class RenderManager {
  private currentRenderer: BaseRenderer | null = null;
  private config: RenderManagerConfig;
  private isActive: boolean = false;
  private renderLoopId: number | null = null;
  private frameCount: number = 0;

  constructor(config: RenderManagerConfig) {
    this.config = config;
  }

  async initializeRenderer(type?: RendererType): Promise<void> {
    const rendererType = type || this.config.defaultRenderer;
    
    try {
      // Clean up existing renderer
      if (this.currentRenderer) {
        await this.destroyRenderer();
      }

      // Create new renderer
      this.currentRenderer = rendererFactory.createRenderer(rendererType);
      
      const rendererConfig: RendererConfig = {
        type: rendererType,
        width: this.config.width,
        height: this.config.height,
        antialias: true,
        backgroundColor: 0x1a1a1a,
        pixelsPerUnit: 40
      };

      await this.currentRenderer.init(this.config.container, rendererConfig);
      console.log(`Renderer initialized: ${rendererType}`);
      
    } catch (error) {
      console.error(`Failed to initialize renderer ${rendererType}:`, error);
      
      // Try fallback renderer if different from requested type
      if (rendererType !== this.config.fallbackRenderer) {
        console.log(`Attempting fallback to ${this.config.fallbackRenderer}`);
        await this.initializeRenderer(this.config.fallbackRenderer);
      } else {
        throw new Error(`Failed to initialize any renderer: ${error}`);
      }
    }
  }

  async switchRenderer(type: RendererType): Promise<void> {
    if (!rendererFactory.isTypeSupported(type)) {
      throw new Error(`Renderer type not supported: ${type}`);
    }

    const wasActive = this.isActive;
    
    if (wasActive) {
      this.stopRenderLoop();
    }

    await this.initializeRenderer(type);
    
    if (wasActive) {
      this.startRenderLoop();
    }
  }

  startRenderLoop(): void {
    if (!this.currentRenderer) {
      console.warn('Cannot start render loop: no renderer initialized');
      return;
    }

    if (this.isActive) {
      console.warn('Render loop already active');
      return;
    }

    this.isActive = true;
    this.frameCount = 0;
    console.log('[Render] Starting render loop');

    const renderFrame = () => {
      if (!this.isActive || !this.currentRenderer) {
        return;
      }

      try {
        const currentGameState = gameState.getState();
        const staticData = gameState.getStaticData();

        if (this.frameCount < 5) {
          console.log('[Render] Frame', this.frameCount,
            '- gameState:', !!currentGameState,
            '- staticData:', !!staticData,
            '- walls:', staticData?.walls?.length ?? 0);
        }
        this.frameCount++;

        if (currentGameState && staticData) {
          this.currentRenderer.render(currentGameState, staticData);
        }
      } catch (error) {
        console.error('Render loop error:', error);
      }

      this.renderLoopId = requestAnimationFrame(renderFrame);
    };

    this.renderLoopId = requestAnimationFrame(renderFrame);
  }

  stopRenderLoop(): void {
    if (!this.isActive) {
      return;
    }

    console.log('[Render] Stopping render loop');
    this.isActive = false;

    if (this.renderLoopId !== null) {
      cancelAnimationFrame(this.renderLoopId);
      this.renderLoopId = null;
    }
  }

  pauseRenderer(): void {
    if (this.currentRenderer) {
      this.currentRenderer.pause();
      console.log('Renderer paused');
    }
  }

  resumeRenderer(): void {
    if (this.currentRenderer) {
      this.currentRenderer.resume();
      console.log('Renderer resumed');
    }
  }

  async destroyRenderer(): Promise<void> {
    this.stopRenderLoop();
    
    if (this.currentRenderer) {
      console.log('Destroying current renderer');
      this.currentRenderer.destroy();
      this.currentRenderer = null;
    }
  }

  resizeRenderer(width: number, height: number): void {
    if (this.currentRenderer) {
      this.currentRenderer.resize(width, height);
      this.config.width = width;
      this.config.height = height;
    }
  }

  getCurrentRendererType(): RendererType | null {
    return this.currentRenderer ? this.currentRenderer.getType() : null;
  }

  getSupportedRendererTypes(): RendererType[] {
    return rendererFactory.getSupportedTypes();
  }

  isRendererActive(): boolean {
    return this.isActive && this.currentRenderer !== null;
  }

  getCurrentRenderer(): BaseRenderer | null {
    return this.currentRenderer;
  }

  // Event handlers for renderer callbacks
  setPlayerClickHandler(handler: (playerId: number) => void): void {
    if (this.currentRenderer) {
      this.currentRenderer.onPlayerClick = handler;
    }
  }

  setWallClickHandler(handler: (wallId: string) => void): void {
    if (this.currentRenderer) {
      this.currentRenderer.onWallClick = handler;
    }
  }
}