import { writable, derived } from 'svelte/store';
import { browser } from '$app/environment';

// Check if we're running in Wails
const isWails = typeof window !== 'undefined' && window.wails;

/**
 * Real-time SSE connection state
 */
export const CONNECTION_STATES = {
  DISCONNECTED: 'disconnected',
  CONNECTING: 'connecting',
  CONNECTED: 'connected',
  RECONNECTING: 'reconnecting',
  ERROR: 'error'
};

/**
 * Real-time event types
 */
export const EVENT_TYPES = {
  HIGHLIGHTS_UPDATED: 'highlights_updated',
  HIGHLIGHTS_DELETED: 'highlights_deleted',
  HIGHLIGHTS_REORDERED: 'highlights_reordered',
  PROJECT_UPDATED: 'project_updated',
  CONNECTED: 'connected',
  DISCONNECTED: 'disconnected'
};

/**
 * Real-time connection manager
 */
class RealtimeManager {
  constructor() {
    this.eventSource = null;
    this.projectId = null;
    this.connectionState = writable(CONNECTION_STATES.DISCONNECTED);
    this.lastEvent = writable(null);
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 10;
    this.baseReconnectDelay = 1000; // Start with 1 second
    this.reconnectTimer = null;
    this.eventListeners = new Map();
    
    // Connection info
    this.connectedAt = null;
    this.lastPingAt = null;
    this.reconnectDelay = this.baseReconnectDelay;
  }

  /**
   * Connect to real-time updates for a specific project
   */
  connect(projectId) {
    if (!browser) return;
    
    if (this.projectId === projectId) {
      console.log(`Already connected to project ${projectId}`);
      return;
    }

    this.disconnect();
    this.projectId = projectId;
    this.connectionState.set(CONNECTION_STATES.CONNECTING);
    
    if (isWails) {
      // Use Wails events for desktop app
      this.connectWailsEvents(projectId);
    } else {
      // Use SSE for browser
      this.connectSSE(projectId);
    }
  }

  /**
   * Connect using Wails events
   */
  connectWailsEvents(projectId) {
    try {
      console.log(`Connecting to Wails events for project ${projectId}`);
      
      // Import Wails runtime dynamically
      import('$lib/wailsjs/runtime/runtime.js').then((runtime) => {
        // Set up Wails event listeners
        runtime.EventsOn(EVENT_TYPES.HIGHLIGHTS_UPDATED, (data) => {
          console.log('Received Wails event - highlights update:', data);
          this.handleWailsEvent(data);
        });
        
        runtime.EventsOn(EVENT_TYPES.HIGHLIGHTS_DELETED, (data) => {
          console.log('Received Wails event - highlights deleted:', data);
          this.handleWailsEvent(data);
        });
        
        runtime.EventsOn(EVENT_TYPES.HIGHLIGHTS_REORDERED, (data) => {
          console.log('Received Wails event - highlights reordered:', data);
          this.handleWailsEvent(data);
        });
        
        this.connectionState.set(CONNECTION_STATES.CONNECTED);
        this.connectedAt = new Date();
        console.log(`Wails events connected for project ${projectId}`);
      }).catch((error) => {
        console.error('Failed to connect Wails events:', error);
        this.connectionState.set(CONNECTION_STATES.ERROR);
      });
      
    } catch (error) {
      console.error('Failed to set up Wails events:', error);
      this.connectionState.set(CONNECTION_STATES.ERROR);
    }
  }

  /**
   * Connect using SSE
   */
  connectSSE(projectId) {
    try {
      const url = `/api/sse/highlights?projectId=${encodeURIComponent(projectId)}`;
      console.log(`Connecting to SSE: ${url}`);
      
      this.eventSource = new EventSource(url);
      
      this.eventSource.onopen = () => {
        console.log(`SSE connection opened for project ${projectId}`);
        this.connectionState.set(CONNECTION_STATES.CONNECTED);
        this.connectedAt = new Date();
        this.reconnectAttempts = 0;
        this.reconnectDelay = this.baseReconnectDelay;
        
        if (this.reconnectTimer) {
          clearTimeout(this.reconnectTimer);
          this.reconnectTimer = null;
        }
      };
      
      this.eventSource.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          console.log('Received SSE event:', data);
          
          // Update last event
          this.lastEvent.set({
            ...data,
            receivedAt: new Date()
          });
          
          // Handle ping events
          if (event.data === ': ping') {
            this.lastPingAt = new Date();
            return;
          }
          
          // Trigger event listeners
          this.triggerEventListeners(data.type, data);
          
        } catch (error) {
          console.error('Failed to parse SSE message:', error, event.data);
        }
      };
      
      this.eventSource.onerror = (error) => {
        console.error('SSE connection error:', error);
        
        if (this.eventSource?.readyState === EventSource.CLOSED) {
          console.log('SSE connection closed, attempting to reconnect...');
          this.connectionState.set(CONNECTION_STATES.RECONNECTING);
          this.scheduleReconnect();
        } else {
          this.connectionState.set(CONNECTION_STATES.ERROR);
        }
      };
      
    } catch (error) {
      console.error('Failed to create SSE connection:', error);
      this.connectionState.set(CONNECTION_STATES.ERROR);
      this.scheduleReconnect();
    }
  }

  /**
   * Handle Wails event data
   */
  handleWailsEvent(data) {
    try {
      console.log('Processing Wails event:', data);
      
      // Update last event
      this.lastEvent.set({
        ...data,
        receivedAt: new Date()
      });
      
      // Trigger event listeners
      this.triggerEventListeners(data.type, data);
      
    } catch (error) {
      console.error('Failed to process Wails event:', error, data);
    }
  }

  /**
   * Disconnect from real-time updates
   */
  disconnect() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    
    console.log(`Disconnecting from project ${this.projectId}`);
    
    if (isWails) {
      // Disconnect Wails events - Note: Wails doesn't have EventsOff, so events will remain
      // This is generally okay as they'll be filtered by project ID
    } else if (this.eventSource) {
      // Disconnect SSE
      this.eventSource.close();
      this.eventSource = null;
    }
    
    this.projectId = null;
    this.connectionState.set(CONNECTION_STATES.DISCONNECTED);
    this.connectedAt = null;
    this.lastPingAt = null;
    this.reconnectAttempts = 0;
  }

  /**
   * Schedule a reconnection attempt
   */
  scheduleReconnect() {
    if (!this.projectId || this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.log('Max reconnection attempts reached or no project ID');
      this.connectionState.set(CONNECTION_STATES.ERROR);
      return;
    }
    
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
    }
    
    this.reconnectAttempts++;
    console.log(`Scheduling reconnect attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts} in ${this.reconnectDelay}ms`);
    
    this.reconnectTimer = setTimeout(() => {
      if (this.projectId) {
        console.log(`Attempting reconnection ${this.reconnectAttempts}/${this.maxReconnectAttempts}`);
        this.connect(this.projectId);
      }
    }, this.reconnectDelay);
    
    // Exponential backoff with jitter
    this.reconnectDelay = Math.min(
      this.reconnectDelay * 2 + Math.random() * 1000,
      30000 // Max 30 seconds
    );
  }

  /**
   * Add an event listener for specific event types
   */
  addEventListener(eventType, callback) {
    if (!this.eventListeners.has(eventType)) {
      this.eventListeners.set(eventType, new Set());
    }
    this.eventListeners.get(eventType).add(callback);
    
    // Return unsubscribe function
    return () => {
      const listeners = this.eventListeners.get(eventType);
      if (listeners) {
        listeners.delete(callback);
        if (listeners.size === 0) {
          this.eventListeners.delete(eventType);
        }
      }
    };
  }

  /**
   * Remove an event listener
   */
  removeEventListener(eventType, callback) {
    const listeners = this.eventListeners.get(eventType);
    if (listeners) {
      listeners.delete(callback);
      if (listeners.size === 0) {
        this.eventListeners.delete(eventType);
      }
    }
  }

  /**
   * Trigger event listeners for a specific event type
   */
  triggerEventListeners(eventType, data) {
    const listeners = this.eventListeners.get(eventType);
    if (listeners) {
      listeners.forEach(callback => {
        try {
          callback(data);
        } catch (error) {
          console.error(`Error in event listener for ${eventType}:`, error);
        }
      });
    }
    
    // Also trigger listeners for 'all' events
    const allListeners = this.eventListeners.get('all');
    if (allListeners) {
      allListeners.forEach(callback => {
        try {
          callback(eventType, data);
        } catch (error) {
          console.error(`Error in 'all' event listener:`, error);
        }
      });
    }
  }

  /**
   * Get connection statistics
   */
  getStats() {
    return {
      connectionState: this.connectionState,
      projectId: this.projectId,
      connectedAt: this.connectedAt,
      lastPingAt: this.lastPingAt,
      reconnectAttempts: this.reconnectAttempts,
      hasEventSource: !!this.eventSource,
      readyState: this.eventSource?.readyState,
      eventListenerCount: Array.from(this.eventListeners.values())
        .reduce((total, listeners) => total + listeners.size, 0)
    };
  }
}

// Create singleton instance
const realtimeManager = new RealtimeManager();

// Export stores and manager
export const connectionState = realtimeManager.connectionState;
export const lastEvent = realtimeManager.lastEvent;

// Derived store for connection status
export const isConnected = derived(
  connectionState,
  $state => $state === CONNECTION_STATES.CONNECTED
);

export const isConnecting = derived(
  connectionState,
  $state => $state === CONNECTION_STATES.CONNECTING || $state === CONNECTION_STATES.RECONNECTING
);

export const hasError = derived(
  connectionState,
  $state => $state === CONNECTION_STATES.ERROR
);

/**
 * Connect to real-time updates for a project
 */
export function connectToProject(projectId) {
  realtimeManager.connect(projectId?.toString());
}

/**
 * Disconnect from real-time updates
 */
export function disconnect() {
  realtimeManager.disconnect();
}

/**
 * Subscribe to specific event types
 */
export function onRealtimeEvent(eventType, callback) {
  return realtimeManager.addEventListener(eventType, callback);
}

/**
 * Subscribe to all events
 */
export function onAllRealtimeEvents(callback) {
  return realtimeManager.addEventListener('all', callback);
}

/**
 * Get real-time connection statistics
 */
export function getRealtimeStats() {
  return realtimeManager.getStats();
}

// Cleanup on page unload
if (browser) {
  window.addEventListener('beforeunload', () => {
    realtimeManager.disconnect();
  });
}

// Export manager for advanced usage
export { realtimeManager };