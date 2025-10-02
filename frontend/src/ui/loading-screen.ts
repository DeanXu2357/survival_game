import type { UIScreen, LoadingScreenConfig } from '../types/ui-types';

export class LoadingScreen implements UIScreen {
  private config: LoadingScreenConfig;
  private container: HTMLElement;
  private visible: boolean = false;
  
  // UI elements
  private messageElement: HTMLElement | null = null;
  private spinnerElement: HTMLElement | null = null;

  constructor(config: LoadingScreenConfig) {
    this.config = config;
    this.container = config.container;
    
    this.createUI();
  }

  private createUI(): void {
    // Clear container
    this.container.innerHTML = '';
    this.container.className = 'loading-screen';

    // Create main loading container
    const loadingContainer = document.createElement('div');
    loadingContainer.className = 'loading-container';

    // Create spinner
    this.spinnerElement = document.createElement('div');
    this.spinnerElement.className = 'loading-spinner';
    this.spinnerElement.innerHTML = `
      <div class="spinner-circle"></div>
    `;

    // Create message
    this.messageElement = document.createElement('div');
    this.messageElement.className = 'loading-message';
    this.messageElement.textContent = this.config.message || 'Loading...';

    // Create game title/logo area (optional)
    const titleElement = document.createElement('div');
    titleElement.className = 'loading-title';
    titleElement.innerHTML = `
      <h1>Survival Game</h1>
      <p>Multiplayer Battle Arena</p>
    `;

    // Assemble loading screen
    loadingContainer.appendChild(titleElement);
    loadingContainer.appendChild(this.spinnerElement);
    loadingContainer.appendChild(this.messageElement);

    this.container.appendChild(loadingContainer);

    console.log('Loading screen created');
  }

  show(): void {
    this.visible = true;
    this.container.classList.remove('hidden');
    console.log('Loading screen shown');
  }

  hide(): void {
    this.visible = false;
    this.container.classList.add('hidden');
    console.log('Loading screen hidden');
  }

  isVisible(): boolean {
    return this.visible;
  }

  updateMessage(message: string): void {
    if (this.messageElement) {
      this.messageElement.textContent = message;
      console.log('Loading screen message updated:', message);
    }
  }

  setProgress(progress: number): void {
    // For future use - add a progress bar
    // progress should be between 0 and 100
    console.log('Loading progress:', progress + '%');
    
    // If we add a progress bar later, update it here
    // const progressBar = this.container.querySelector('.progress-bar') as HTMLElement;
    // if (progressBar) {
    //   progressBar.style.width = progress + '%';
    // }
  }

  showError(error: string): void {
    if (this.messageElement) {
      this.messageElement.textContent = `Error: ${error}`;
      this.messageElement.classList.add('error');
      
      // Hide spinner on error
      if (this.spinnerElement) {
        this.spinnerElement.style.display = 'none';
      }
    }
  }

  clearError(): void {
    if (this.messageElement) {
      this.messageElement.classList.remove('error');
      
      // Show spinner again
      if (this.spinnerElement) {
        this.spinnerElement.style.display = 'block';
      }
      
      // Reset to default message
      this.updateMessage(this.config.message || 'Loading...');
    }
  }

  destroy(): void {
    // Clear container
    this.container.innerHTML = '';
    
    console.log('Loading screen destroyed');
  }
}