import { writable, derived } from 'svelte/store';
import { onRealtimeEvent, EVENT_TYPES } from './realtime.js';

/**
 * Chatbot real-time message store
 * Manages real-time synchronization of chat messages across different endpoints
 */

// Store for tracking connected chatbot sessions by project and endpoint
const chatbotSessions = writable(new Map()); // Map<`${projectId}_${endpointId}`, { messages: [], sessionId: string }>

// Store for real-time chatbot events
export const lastChatbotEvent = writable(null);

// Store for chatbot progress messages
const chatbotProgress = writable(new Map()); // Map<`${projectId}_${endpointId}`, string>

/**
 * Connect to real-time updates for a specific chatbot session
 * @param {number} projectId - The project ID
 * @param {string} endpointId - The chatbot endpoint ID
 * @param {Array} initialMessages - Initial messages to populate
 * @param {string} initialSessionId - Initial session ID
 * @returns {function} Unsubscribe function
 */
export function connectChatbotSession(projectId, endpointId, initialMessages = [], initialSessionId = null) {
  const sessionKey = `${projectId}_${endpointId}`;
  
  // Initialize or update session
  chatbotSessions.update(sessions => {
    const newSessions = new Map(sessions);
    newSessions.set(sessionKey, {
      messages: [...initialMessages],
      sessionId: initialSessionId,
      projectId,
      endpointId
    });
    return newSessions;
  });
  
  console.log(`Connected chatbot session for project ${projectId}, endpoint ${endpointId}`);
  
  // Set up event listeners for this session
  const unsubscribers = [
    onRealtimeEvent(EVENT_TYPES.CHAT_MESSAGE_ADDED, (data) => {
      handleChatMessageAdded(projectId, endpointId, data);
    }),
    
    onRealtimeEvent(EVENT_TYPES.CHAT_HISTORY_CLEARED, (data) => {
      handleChatHistoryCleared(projectId, endpointId, data);
    }),
    
    onRealtimeEvent(EVENT_TYPES.CHAT_SESSION_UPDATED, (data) => {
      handleChatSessionUpdated(projectId, endpointId, data);
    }),
    
    onRealtimeEvent(EVENT_TYPES.CHAT_PROGRESS, (data) => {
      handleChatProgress(projectId, endpointId, data);
    })
  ];
  
  // Return combined unsubscribe function
  return () => {
    unsubscribers.forEach(unsubscribe => unsubscribe());
    disconnectChatbotSession(projectId, endpointId);
  };
}

/**
 * Disconnect from real-time updates for a specific chatbot session
 * @param {number} projectId - The project ID  
 * @param {string} endpointId - The chatbot endpoint ID
 */
export function disconnectChatbotSession(projectId, endpointId) {
  const sessionKey = `${projectId}_${endpointId}`;
  
  chatbotSessions.update(sessions => {
    const newSessions = new Map(sessions);
    newSessions.delete(sessionKey);
    return newSessions;
  });
  
  console.log(`Disconnected chatbot session for project ${projectId}, endpoint ${endpointId}`);
}

/**
 * Get messages for a specific chatbot session
 * @param {number} projectId - The project ID
 * @param {string} endpointId - The chatbot endpoint ID
 * @returns {import('svelte/store').Readable<Array>} Readable store with messages
 */
export function getChatbotMessages(projectId, endpointId) {
  const sessionKey = `${projectId}_${endpointId}`;
  
  return derived(chatbotSessions, $sessions => {
    const session = $sessions.get(sessionKey);
    return session ? session.messages : [];
  });
}

/**
 * Get session ID for a specific chatbot session
 * @param {number} projectId - The project ID
 * @param {string} endpointId - The chatbot endpoint ID
 * @returns {import('svelte/store').Readable<string|null>} Readable store with session ID
 */
export function getChatbotSessionId(projectId, endpointId) {
  const sessionKey = `${projectId}_${endpointId}`;
  
  return derived(chatbotSessions, $sessions => {
    const session = $sessions.get(sessionKey);
    return session ? session.sessionId : null;
  });
}

/**
 * Get progress message for a specific chatbot session
 * @param {number} projectId - The project ID
 * @param {string} endpointId - The chatbot endpoint ID
 * @returns {import('svelte/store').Readable<string|null>} Readable store with progress message
 */
export function getChatbotProgress(projectId, endpointId) {
  const sessionKey = `${projectId}_${endpointId}`;
  
  return derived(chatbotProgress, $progress => {
    return $progress.get(sessionKey) || null;
  });
}

/**
 * Add a message to a specific chatbot session
 * @param {number} projectId - The project ID
 * @param {string} endpointId - The chatbot endpoint ID
 * @param {Object} message - The message to add
 */
export function addChatbotMessage(projectId, endpointId, message) {
  const sessionKey = `${projectId}_${endpointId}`;
  
  chatbotSessions.update(sessions => {
    const newSessions = new Map(sessions);
    const session = newSessions.get(sessionKey);
    
    if (session) {
      session.messages = [...session.messages, message];
      newSessions.set(sessionKey, session);
    }
    
    return newSessions;
  });
}

/**
 * Clear all messages for a specific chatbot session
 * @param {number} projectId - The project ID
 * @param {string} endpointId - The chatbot endpoint ID
 */
export function clearChatbotMessages(projectId, endpointId) {
  const sessionKey = `${projectId}_${endpointId}`;
  
  chatbotSessions.update(sessions => {
    const newSessions = new Map(sessions);
    const session = newSessions.get(sessionKey);
    
    if (session) {
      session.messages = [];
      session.sessionId = null;
      newSessions.set(sessionKey, session);
    }
    
    return newSessions;
  });
}

/**
 * Update session ID for a specific chatbot session
 * @param {number} projectId - The project ID
 * @param {string} endpointId - The chatbot endpoint ID
 * @param {string} sessionId - The new session ID
 */
export function updateChatbotSessionId(projectId, endpointId, sessionId) {
  const sessionKey = `${projectId}_${endpointId}`;
  
  chatbotSessions.update(sessions => {
    const newSessions = new Map(sessions);
    const session = newSessions.get(sessionKey);
    
    if (session) {
      session.sessionId = sessionId;
      newSessions.set(sessionKey, session);
    }
    
    return newSessions;
  });
}

// Event handlers

function handleChatMessageAdded(projectId, endpointId, eventData) {
  console.log('Processing chat message added event:', eventData);
  
  // Only process if this event is for the current project
  if (eventData.projectId === projectId?.toString()) {
    try {
      const { data } = eventData;
      
      // Check if this message is for the current endpoint
      if (data.endpointId === endpointId) {
        const message = data.message;
        
        if (message) {
          console.log(`Adding message to chatbot session ${projectId}_${endpointId}:`, message);
          addChatbotMessage(projectId, endpointId, message);
          
          // Update session ID if provided
          if (data.sessionId) {
            updateChatbotSessionId(projectId, endpointId, data.sessionId);
          }
          
          // Update last event
          lastChatbotEvent.set({
            type: EVENT_TYPES.CHAT_MESSAGE_ADDED,
            projectId,
            endpointId,
            message,
            timestamp: new Date()
          });
        }
      }
    } catch (error) {
      console.error('Error processing chat message added event:', error);
    }
  }
}

function handleChatHistoryCleared(projectId, endpointId, eventData) {
  console.log('Processing chat history cleared event:', eventData);
  
  // Only process if this event is for the current project
  if (eventData.projectId === projectId?.toString()) {
    try {
      const { data } = eventData;
      
      // Check if this clear is for the current endpoint
      if (data.endpointId === endpointId) {
        console.log(`Clearing chat history for session ${projectId}_${endpointId}`);
        clearChatbotMessages(projectId, endpointId);
        
        // Update last event
        lastChatbotEvent.set({
          type: EVENT_TYPES.CHAT_HISTORY_CLEARED,
          projectId,
          endpointId,
          timestamp: new Date()
        });
      }
    } catch (error) {
      console.error('Error processing chat history cleared event:', error);
    }
  }
}

function handleChatSessionUpdated(projectId, endpointId, eventData) {
  console.log('Processing chat session updated event:', eventData);
  
  // Only process if this event is for the current project
  if (eventData.projectId === projectId?.toString()) {
    try {
      const { data } = eventData;
      
      // Check if this update is for the current endpoint
      if (data.endpointId === endpointId) {
        const { sessionId, messages } = data;
        
        if (sessionId) {
          updateChatbotSessionId(projectId, endpointId, sessionId);
        }
        
        if (messages && Array.isArray(messages)) {
          // Replace all messages with the updated list
          const sessionKey = `${projectId}_${endpointId}`;
          
          chatbotSessions.update(sessions => {
            const newSessions = new Map(sessions);
            const session = newSessions.get(sessionKey);
            
            if (session) {
              session.messages = [...messages];
              if (sessionId) {
                session.sessionId = sessionId;
              }
              newSessions.set(sessionKey, session);
            }
            
            return newSessions;
          });
          
          console.log(`Updated chat session ${projectId}_${endpointId} with ${messages.length} messages`);
        }
        
        // Update last event
        lastChatbotEvent.set({
          type: EVENT_TYPES.CHAT_SESSION_UPDATED,
          projectId,
          endpointId,
          sessionId,
          messageCount: messages?.length || 0,
          timestamp: new Date()
        });
      }
    } catch (error) {
      console.error('Error processing chat session updated event:', error);
    }
  }
}

function handleChatProgress(projectId, endpointId, eventData) {
  console.log('Processing chat progress event:', eventData);
  
  // Only process if this event is for the current project
  if (eventData.projectId === projectId?.toString()) {
    try {
      const { data } = eventData;
      
      // Check if this progress is for the current endpoint
      if (data.endpointId === endpointId) {
        const sessionKey = `${projectId}_${endpointId}`;
        const message = data.message;
        
        chatbotProgress.update(progress => {
          const newProgress = new Map(progress);
          newProgress.set(sessionKey, message);
          return newProgress;
        });
        
        console.log(`Updated chat progress for ${projectId}_${endpointId}: ${message}`);
        
        // Clear progress after 5 seconds to avoid stale messages
        setTimeout(() => {
          chatbotProgress.update(progress => {
            const newProgress = new Map(progress);
            if (newProgress.get(sessionKey) === message) {
              newProgress.delete(sessionKey);
            }
            return newProgress;
          });
        }, 5000);
      }
    } catch (error) {
      console.error('Error processing chat progress event:', error);
    }
  }
}

// Export chatbot sessions store for debugging
export { chatbotSessions };