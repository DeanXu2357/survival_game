import type { BaseRenderer, RendererFactory } from '../types/renderer-types';
import { RendererType } from '../types/renderer-types';
import { PixiRenderer } from './pixi-renderer';
// Future renderers can be imported here
// import { Three3DRenderer } from './three-3d-renderer';
// import { Canvas2DRenderer } from './canvas-2d-renderer';
// import { WebGLRenderer } from './webgl-renderer';

export class DefaultRendererFactory implements RendererFactory {
  createRenderer(type: RendererType): BaseRenderer {
    switch (type) {
      case RendererType.PIXI_2D:
        return new PixiRenderer();
        
      case RendererType.THREE_3D:
        // TODO: Implement Three.js renderer
        console.warn('Three.js renderer not yet implemented, falling back to PixiJS');
        return new PixiRenderer();
        
      case RendererType.CANVAS_2D:
        // TODO: Implement Canvas 2D renderer
        console.warn('Canvas 2D renderer not yet implemented, falling back to PixiJS');
        return new PixiRenderer();
        
      case RendererType.WEBGL:
        // TODO: Implement WebGL renderer
        console.warn('WebGL renderer not yet implemented, falling back to PixiJS');
        return new PixiRenderer();
        
      default:
        console.warn(`Unknown renderer type: ${type}, falling back to PixiJS`);
        return new PixiRenderer();
    }
  }

  getSupportedTypes(): RendererType[] {
    return [
      RendererType.PIXI_2D,
      // Add other types when implemented
      // RendererType.THREE_3D,
      // RendererType.CANVAS_2D,
      // RendererType.WEBGL
    ];
  }

  getDefaultType(): RendererType {
    return RendererType.PIXI_2D;
  }

  getFallbackType(): RendererType {
    return RendererType.PIXI_2D;
  }

  isTypeSupported(type: RendererType): boolean {
    return this.getSupportedTypes().includes(type);
  }
}

// Export singleton instance
export const rendererFactory = new DefaultRendererFactory();