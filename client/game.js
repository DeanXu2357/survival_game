// Game States
const GAME_STATES = {
    MENU: 'menu',
    PLAYING: 'playing',
    RESULT: 'result'
};

class Game {
    constructor() {
        this.canvas = document.getElementById('gameCanvas');
        this.ctx = this.canvas.getContext('2d');
        this.ws = null;
        this.playerID = null;
        this.players = {};
        this.keys = {};
        this.mouse = { x: 0, y: 0, angle: 0 };
        
        this.PLAYER_SIZE = 20;
        this.VISION_RADIUS = 20; // 1 body length
        this.VISION_CONE_DISTANCE = 200; // 10 body lengths
        this.VISION_CONE_ANGLE = Math.PI / 4; // 45 degrees
        
        // Game state management
        this.gameState = GAME_STATES.MENU;
        this.gameMode = null;
        this.gameStartTime = null;
        this.gameStats = {
            enemiesKilled: 0,
            survivalTime: 0,
            finalScore: 0
        };
        
        this.setupEventListeners();
        // Don't auto-connect anymore, wait for game start
    }
    
    // Screen management
    showScreen(screenId) {
        document.querySelectorAll('.screen').forEach(screen => {
            screen.classList.remove('active');
        });
        document.getElementById(screenId).classList.add('active');
    }
    
    // Game flow methods
    startGame(mode) {
        this.gameMode = mode;
        this.gameState = GAME_STATES.PLAYING;
        this.gameStartTime = Date.now();
        this.gameStats = { enemiesKilled: 0, survivalTime: 0, finalScore: 0 };
        
        this.showScreen('gameScreen');
        this.connect();
    }
    
    endGame(result = 'defeat') {
        this.gameState = GAME_STATES.RESULT;
        
        // Calculate survival time
        if (this.gameStartTime) {
            this.gameStats.survivalTime = Math.floor((Date.now() - this.gameStartTime) / 1000);
        }
        
        // Calculate final score (survival time + kills)
        this.gameStats.finalScore = this.gameStats.survivalTime + (this.gameStats.enemiesKilled * 10);
        
        // Disconnect from server
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
        
        // Show results
        this.showResults(result);
    }
    
    showResults(result) {
        const resultTitle = document.getElementById('resultTitle');
        const resultText = document.getElementById('resultText');
        const survivalTimeSpan = document.getElementById('survivalTime');
        const enemiesKilledSpan = document.getElementById('enemiesKilled');
        const finalScoreSpan = document.getElementById('finalScore');
        
        if (result === 'victory') {
            resultTitle.textContent = 'Victory!';
            resultTitle.style.color = '#0f0';
            resultText.textContent = 'Congratulations! You survived the challenge!';
        } else {
            resultTitle.textContent = 'Game Over';
            resultTitle.style.color = '#f00';
            resultText.textContent = `You survived for ${this.formatTime(this.gameStats.survivalTime)}!`;
        }
        
        survivalTimeSpan.textContent = this.formatTime(this.gameStats.survivalTime);
        enemiesKilledSpan.textContent = this.gameStats.enemiesKilled;
        finalScoreSpan.textContent = this.gameStats.finalScore;
        
        this.showScreen('resultScreen');
    }
    
    returnToMenu() {
        this.gameState = GAME_STATES.MENU;
        this.gameMode = null;
        this.playerID = null;
        this.players = {};
        
        this.showScreen('mainMenu');
    }
    
    formatTime(seconds) {
        const minutes = Math.floor(seconds / 60);
        const remainingSeconds = seconds % 60;
        return `${minutes}:${remainingSeconds.toString().padStart(2, '0')}`;
    }
    
    connect() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        this.ws = new WebSocket(`${protocol}//${window.location.host}/ws`);
        
        this.ws.onopen = () => {
            document.getElementById('status').textContent = 'Connected';
            this.startGameLoop();
        };
        
        this.ws.onmessage = (event) => {
            const message = JSON.parse(event.data);
            this.handleMessage(message);
        };
        
        this.ws.onclose = () => {
            document.getElementById('status').textContent = 'Disconnected';
        };
        
        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            document.getElementById('status').textContent = 'Connection Error';
        };
    }
    
    handleMessage(message) {
        switch (message.type) {
            case 'init':
                this.playerID = message.data.playerID;
                document.getElementById('playerID').textContent = this.playerID;
                break;
                
            case 'gameState':
                this.players = message.data.players;
                document.getElementById('playerCount').textContent = Object.keys(this.players).length;
                break;
        }
    }
    
    setupEventListeners() {
        // Keyboard events
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && this.gameState === GAME_STATES.PLAYING) {
                this.endGame('quit');
                return;
            }
            this.keys[e.key.toLowerCase()] = true;
        });
        
        document.addEventListener('keyup', (e) => {
            this.keys[e.key.toLowerCase()] = false;
        });
        
        // Mouse events
        this.canvas.addEventListener('mousemove', (e) => {
            const rect = this.canvas.getBoundingClientRect();
            this.mouse.x = e.clientX - rect.left;
            this.mouse.y = e.clientY - rect.top;
            
            // Calculate angle relative to player (server will handle smoothing)
            if (this.playerID && this.players[this.playerID]) {
                const player = this.players[this.playerID];
                this.mouse.angle = Math.atan2(
                    this.mouse.y - player.y,
                    this.mouse.x - player.x
                );
            }
        });
        
        this.canvas.addEventListener('click', (e) => {
            this.sendMessage('attack', { x: this.mouse.x, y: this.mouse.y });
        });
        
        // Prevent context menu
        this.canvas.addEventListener('contextmenu', (e) => e.preventDefault());
    }
    
    startGameLoop() {
        const loop = () => {
            if (this.gameState === GAME_STATES.PLAYING) {
                this.sendInput();
                this.render();
            }
            requestAnimationFrame(loop);
        };
        requestAnimationFrame(loop);
    }
    
    sendInput() {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.sendMessage('input', {
                keys: {
                    w: this.keys['w'] || false,
                    a: this.keys['a'] || false,
                    s: this.keys['s'] || false,
                    d: this.keys['d'] || false
                },
                mouse: {
                    x: this.mouse.x,
                    y: this.mouse.y,
                    angle: this.mouse.angle
                }
            });
        }
    }
    
    sendMessage(type, data) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify({ type, data }));
        }
    }
    
    render() {
        // Clear canvas with a lighter background color for better contrast with fog
        this.ctx.fillStyle = '#666';  // Medium gray for better contrast
        this.ctx.fillRect(0, 0, this.canvas.width, this.canvas.height);
        
        if (!this.playerID || !this.players[this.playerID]) {
            return;
        }
        
        const currentPlayer = this.players[this.playerID];
        
        // Render all players first (on the visible background)
        this.renderPlayers(currentPlayer);
        
        // Apply fog of war on top
        this.renderFogOfWar(currentPlayer);
        
        // Render UI elements
        this.renderCrosshair(currentPlayer);
        
        // Check for game end conditions (demo: end after 30 seconds)
        if (this.gameStartTime && Date.now() - this.gameStartTime > 30000) {
            this.endGame('victory'); // Demo: auto-win after 30 seconds
        }
    }
    
    renderFogOfWar(player) {
        // Save context
        this.ctx.save();
        
        // Create a clipping mask for NON-visible areas (where fog should be)
        // First, set composite operation to source-over
        this.ctx.globalCompositeOperation = 'source-over';
        
        // Create a path that covers everything EXCEPT the visible areas
        this.ctx.beginPath();
        
        // Create outer rectangle (entire canvas)
        this.ctx.rect(0, 0, this.canvas.width, this.canvas.height);
        
        // Create holes for visible areas (these will NOT be filled with fog)
        // Circular vision hole
        this.ctx.moveTo(player.x + this.VISION_RADIUS, player.y);
        this.ctx.arc(player.x, player.y, this.VISION_RADIUS, 0, 2 * Math.PI, true); // true = counterclockwise = hole
        
        // Cone vision hole
        const startAngle = player.angle - this.VISION_CONE_ANGLE / 2;
        const endAngle = player.angle + this.VISION_CONE_ANGLE / 2;
        
        this.ctx.moveTo(player.x, player.y);
        this.ctx.arc(player.x, player.y, this.VISION_CONE_DISTANCE, startAngle, endAngle, false);
        this.ctx.closePath();
        
        // Fill everything except the holes with fog
        this.ctx.fillStyle = 'rgba(0, 0, 0, 0.8)';
        this.ctx.fill('evenodd'); // Use even-odd rule to create holes
        
        // Restore context
        this.ctx.restore();
    }
    
    renderPlayers(currentPlayer) {
        Object.values(this.players).forEach(player => {
            if (!player.isAlive) return;
            
            // Render ALL players regardless of visibility - fog will handle visibility
            this.ctx.save();
            
            // Move to player position
            this.ctx.translate(player.x, player.y);
            this.ctx.rotate(player.angle);
            
            // Draw triangle (player) - use brighter colors
            this.ctx.fillStyle = player.id === this.playerID ? '#00ff00' : '#ff0000';
            this.ctx.beginPath();
            this.ctx.moveTo(this.PLAYER_SIZE / 2, 0);
            this.ctx.lineTo(-this.PLAYER_SIZE / 2, -this.PLAYER_SIZE / 2);
            this.ctx.lineTo(-this.PLAYER_SIZE / 2, this.PLAYER_SIZE / 2);
            this.ctx.closePath();
            this.ctx.fill();
            
            // Draw player outline
            this.ctx.strokeStyle = '#fff';
            this.ctx.lineWidth = 1;
            this.ctx.stroke();
            
            this.ctx.restore();
        });
    }
    
    renderCrosshair(player) {
        // Draw aiming line (should be visible on top of fog)
        this.ctx.save();
        this.ctx.strokeStyle = '#ff0';
        this.ctx.lineWidth = 2;
        this.ctx.setLineDash([5, 5]);
        this.ctx.globalCompositeOperation = 'source-over'; // Ensure it renders on top
        
        this.ctx.beginPath();
        this.ctx.moveTo(player.x, player.y);
        this.ctx.lineTo(
            player.x + Math.cos(player.angle) * this.VISION_CONE_DISTANCE,
            player.y + Math.sin(player.angle) * this.VISION_CONE_DISTANCE
        );
        this.ctx.stroke();
        
        this.ctx.setLineDash([]);
        this.ctx.restore();
    }
    
    isVisible(viewer, target) {
        if (viewer.id === target.id) return true;
        
        const dx = target.x - viewer.x;
        const dy = target.y - viewer.y;
        const distance = Math.sqrt(dx * dx + dy * dy);
        
        // Check circular vision
        if (distance <= this.VISION_RADIUS) return true;
        
        // Check cone vision
        if (distance <= this.VISION_CONE_DISTANCE) {
            const angle = Math.atan2(dy, dx);
            const angleDiff = Math.abs(angle - viewer.angle);
            const normalizedDiff = Math.min(angleDiff, 2 * Math.PI - angleDiff);
            
            return normalizedDiff <= this.VISION_CONE_ANGLE / 2;
        }
        
        return false;
    }
}

// Global game instance
let gameInstance = null;

// Global functions for HTML to call
function startGame(mode) {
    if (gameInstance) {
        gameInstance.startGame(mode);
    }
}

function returnToMenu() {
    if (gameInstance) {
        gameInstance.returnToMenu();
    }
}

// Start the game system when page loads
window.addEventListener('load', () => {
    gameInstance = new Game();
    gameInstance.startGameLoop(); // Start the render loop
});