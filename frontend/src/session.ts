interface SessionData {
  sessionId: string;
  clientId: string;
  timestamp: number;
}

export class SessionManager {
  private static readonly STORAGE_PREFIX = 'survival_session_';
  private static readonly SESSION_EXPIRY_MS = 24 * 60 * 60 * 1000; // 24 hours

  static getStoredSession(clientId: string): string | null {
    const key = this.STORAGE_PREFIX + clientId;
    
    try {
      const stored = localStorage.getItem(key);
      if (!stored) return null;

      const sessionData: SessionData = JSON.parse(stored);
      
      // Check if session is expired
      const now = Date.now();
      if (now - sessionData.timestamp > this.SESSION_EXPIRY_MS) {
        this.clearSession(clientId);
        return null;
      }

      return sessionData.sessionId;
    } catch (error) {
      console.warn('Failed to retrieve session from storage:', error);
      this.clearSession(clientId);
      return null;
    }
  }

  static storeSession(clientId: string, sessionId: string): void {
    const key = this.STORAGE_PREFIX + clientId;
    const sessionData: SessionData = {
      sessionId,
      clientId,
      timestamp: Date.now()
    };

    try {
      localStorage.setItem(key, JSON.stringify(sessionData));
      console.log(`Session stored for client ${clientId}: ${sessionId}`);
    } catch (error) {
      console.warn('Failed to store session:', error);
    }
  }

  static clearSession(clientId: string): void {
    const key = this.STORAGE_PREFIX + clientId;
    
    try {
      localStorage.removeItem(key);
      console.log(`Session cleared for client ${clientId}`);
    } catch (error) {
      console.warn('Failed to clear session:', error);
    }
  }

  static cleanupExpiredSessions(): void {
    const now = Date.now();
    
    try {
      for (let i = 0; i < localStorage.length; i++) {
        const key = localStorage.key(i);
        if (!key || !key.startsWith(this.STORAGE_PREFIX)) continue;

        const stored = localStorage.getItem(key);
        if (!stored) continue;

        try {
          const sessionData: SessionData = JSON.parse(stored);
          if (now - sessionData.timestamp > this.SESSION_EXPIRY_MS) {
            localStorage.removeItem(key);
            console.log(`Expired session removed: ${key}`);
          }
        } catch (parseError) {
          localStorage.removeItem(key);
          console.log(`Invalid session data removed: ${key}`);
        }
      }
    } catch (error) {
      console.warn('Failed to cleanup expired sessions:', error);
    }
  }
}