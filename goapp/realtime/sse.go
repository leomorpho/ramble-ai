package realtime

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// Client represents a connected SSE client
type Client struct {
	ID        string
	ProjectID string
	Writer    http.ResponseWriter
	Flusher   http.Flusher
	Done      chan bool
	LastPing  time.Time
}

// NewClient creates a new SSE client
func NewClient(id, projectID string, w http.ResponseWriter) *Client {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil
	}

	return &Client{
		ID:        id,
		ProjectID: projectID,
		Writer:    w,
		Flusher:   flusher,
		Done:      make(chan bool),
		LastPing:  time.Now(),
	}
}

// Send sends an event to the client
func (c *Client) Send(event *Event) error {
	sseData, err := event.ToSSE()
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(c.Writer, sseData)
	if err != nil {
		return err
	}

	c.Flusher.Flush()
	return nil
}

// SendPing sends a ping to keep the connection alive
func (c *Client) SendPing() error {
	_, err := fmt.Fprint(c.Writer, ": ping\n\n")
	if err != nil {
		return err
	}

	c.Flusher.Flush()
	c.LastPing = time.Now()
	return nil
}

// Close closes the client connection
func (c *Client) Close() {
	select {
	case c.Done <- true:
	default:
		// Channel already closed or blocked
	}
}

// SSEManager manages Server-Sent Events connections
type SSEManager struct {
	clients    map[string]*Client
	clientsMux sync.RWMutex

	// Project-specific client groups
	projectClients map[string]map[string]*Client
	projectMux     sync.RWMutex

	// Event channels
	eventChan        chan *Event
	addClientChan    chan *Client
	removeClientChan chan string

	ctx    context.Context
	cancel context.CancelFunc
}

// NewSSEManager creates a new SSE manager
func NewSSEManager() *SSEManager {
	ctx, cancel := context.WithCancel(context.Background())

	manager := &SSEManager{
		clients:          make(map[string]*Client),
		projectClients:   make(map[string]map[string]*Client),
		eventChan:        make(chan *Event, 1000), // Buffered channel
		addClientChan:    make(chan *Client, 100),
		removeClientChan: make(chan string, 100),
		ctx:              ctx,
		cancel:           cancel,
	}

	go manager.run()
	go manager.pingClients()

	return manager
}

// AddClient adds a new client to the manager
func (m *SSEManager) AddClient(client *Client) {
	select {
	case m.addClientChan <- client:
	case <-m.ctx.Done():
		log.Println("SSE manager is shutting down, cannot add client")
	}
}

// RemoveClient removes a client from the manager
func (m *SSEManager) RemoveClient(clientID string) {
	select {
	case m.removeClientChan <- clientID:
	case <-m.ctx.Done():
		// Manager is shutting down
	}
}

// BroadcastToProject sends an event to all clients subscribed to a project
func (m *SSEManager) BroadcastToProject(projectID string, event *Event) {
	select {
	case m.eventChan <- event:
	case <-m.ctx.Done():
		log.Println("SSE manager is shutting down, cannot broadcast event")
	default:
		log.Println("Event channel full, dropping event")
	}
}

// GetClientCount returns the total number of connected clients
func (m *SSEManager) GetClientCount() int {
	m.clientsMux.RLock()
	defer m.clientsMux.RUnlock()
	return len(m.clients)
}

// GetProjectClientCount returns the number of clients for a specific project
func (m *SSEManager) GetProjectClientCount(projectID string) int {
	m.projectMux.RLock()
	defer m.projectMux.RUnlock()

	if clients, exists := m.projectClients[projectID]; exists {
		return len(clients)
	}
	return 0
}

// run is the main event loop for the SSE manager
func (m *SSEManager) run() {
	for {
		select {
		case <-m.ctx.Done():
			log.Println("SSE manager shutting down")
			m.closeAllClients()
			return

		case client := <-m.addClientChan:
			m.addClientInternal(client)

		case clientID := <-m.removeClientChan:
			m.removeClientInternal(clientID)

		case event := <-m.eventChan:
			m.broadcastEventInternal(event)
		}
	}
}

// addClientInternal adds a client to internal maps
func (m *SSEManager) addClientInternal(client *Client) {
	m.clientsMux.Lock()
	m.clients[client.ID] = client
	m.clientsMux.Unlock()

	m.projectMux.Lock()
	if m.projectClients[client.ProjectID] == nil {
		m.projectClients[client.ProjectID] = make(map[string]*Client)
	}
	m.projectClients[client.ProjectID][client.ID] = client
	m.projectMux.Unlock()

	log.Printf("Client %s connected to project %s", client.ID, client.ProjectID)

	// Send connection confirmation
	confirmEvent := NewEvent(EventConnected, client.ProjectID, map[string]string{
		"clientId": client.ID,
		"message":  "Connected to real-time updates",
	})

	if err := client.Send(confirmEvent); err != nil {
		log.Printf("Failed to send connection confirmation to client %s: %v", client.ID, err)
		m.RemoveClient(client.ID)
	}
}

// removeClientInternal removes a client from internal maps
func (m *SSEManager) removeClientInternal(clientID string) {
	m.clientsMux.Lock()
	client, exists := m.clients[clientID]
	if exists {
		delete(m.clients, clientID)
	}
	m.clientsMux.Unlock()

	if !exists {
		return
	}

	m.projectMux.Lock()
	if projectClients, exists := m.projectClients[client.ProjectID]; exists {
		delete(projectClients, clientID)
		if len(projectClients) == 0 {
			delete(m.projectClients, client.ProjectID)
		}
	}
	m.projectMux.Unlock()

	client.Close()
	log.Printf("Client %s disconnected from project %s", clientID, client.ProjectID)
}

// broadcastEventInternal sends an event to all relevant clients
func (m *SSEManager) broadcastEventInternal(event *Event) {
	m.projectMux.RLock()
	projectClients, exists := m.projectClients[event.ProjectID]
	if !exists {
		m.projectMux.RUnlock()
		return
	}

	// Create a copy of the client map to avoid holding the lock during send
	clientsCopy := make(map[string]*Client)
	for id, client := range projectClients {
		clientsCopy[id] = client
	}
	m.projectMux.RUnlock()

	// Send to all clients for this project
	var failedClients []string
	for clientID, client := range clientsCopy {
		if err := client.Send(event); err != nil {
			log.Printf("Failed to send event to client %s: %v", clientID, err)
			failedClients = append(failedClients, clientID)
		}
	}

	// Remove failed clients
	for _, clientID := range failedClients {
		m.RemoveClient(clientID)
	}
}

// pingClients sends periodic pings to keep connections alive
func (m *SSEManager) pingClients() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.sendPingsToAllClients()
		}
	}
}

// sendPingsToAllClients sends pings to all connected clients
func (m *SSEManager) sendPingsToAllClients() {
	m.clientsMux.RLock()
	clients := make([]*Client, 0, len(m.clients))
	for _, client := range m.clients {
		clients = append(clients, client)
	}
	m.clientsMux.RUnlock()

	var failedClients []string
	for _, client := range clients {
		if time.Since(client.LastPing) > 5*time.Minute {
			// Client hasn't responded to pings, consider it dead
			failedClients = append(failedClients, client.ID)
			continue
		}

		if err := client.SendPing(); err != nil {
			failedClients = append(failedClients, client.ID)
		}
	}

	// Remove failed clients
	for _, clientID := range failedClients {
		m.RemoveClient(clientID)
	}
}

// closeAllClients closes all client connections
func (m *SSEManager) closeAllClients() {
	m.clientsMux.Lock()
	defer m.clientsMux.Unlock()

	for _, client := range m.clients {
		client.Close()
	}
}

// Shutdown gracefully shuts down the SSE manager
func (m *SSEManager) Shutdown() {
	m.cancel()
}
