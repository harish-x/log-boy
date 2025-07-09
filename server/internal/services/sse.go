package services

import (
	"log"
	"server/internal/models"
	"sync"
	"time"
)

type Client struct {
	ID        string // Unique client ID
	Project   string // Project associated with the client
	Channel   chan models.Log
	closed    bool       // Track if a client is closed
	mu        sync.Mutex // Protect a closed state
	LastSeen  time.Time  // Track last activity
	Connected bool       // Track connection state
}

type SSEService struct {
	mu              sync.RWMutex
	Clients         map[string]*Client
	ProjectClients  map[string]map[string]*Client // Store client pointers instead of channels
	ProjectChannels map[string]chan models.Log
	shutdownChans   map[string]chan struct{} // For graceful goroutine shutdown
}

func NewSSEService() *SSEService {
	return &SSEService{
		Clients:         make(map[string]*Client),
		ProjectClients:  make(map[string]map[string]*Client),
		ProjectChannels: make(map[string]chan models.Log),
		shutdownChans:   make(map[string]chan struct{}),
	}
}

func (s *SSEService) RegisterClient(clientID, project string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// If a client already exists, unregister it first to avoid conflicts
	if existingClient, exists := s.Clients[clientID]; exists {
		existingClient.mu.Lock()
		timeSinceLastSeen := time.Since(existingClient.LastSeen)
		existingClient.mu.Unlock()

		// If a client was seen recently, just reactivate instead of recreating
		if timeSinceLastSeen < 5*time.Second && existingClient.Project == project {
			log.Printf("Reactivating recent client: %s for project: %s", clientID, project)
			existingClient.mu.Lock()
			existingClient.Connected = true
			existingClient.LastSeen = time.Now()
			existingClient.mu.Unlock()
			return
		}

		// Clean up an old client if it's stale or different project
		log.Printf("Client %s exists but is stale or different project, cleaning up", clientID)
		s.unregisterClientUnsafe(clientID, existingClient)
	}

	// Create a new client
	client := &Client{
		ID:        clientID,
		Project:   project,
		Channel:   make(chan models.Log, 100),
		closed:    false,
		LastSeen:  time.Now(),
		Connected: true,
	}
	s.Clients[clientID] = client
	log.Printf("Registered new client: %s for project: %s", clientID, project)

	// Add client to project's client list
	if _, exists := s.ProjectClients[project]; !exists {
		s.ProjectClients[project] = make(map[string]*Client)
	}
	s.ProjectClients[project][clientID] = client

	// If this is the first client for this project, start a fan-out goroutine
	if _, exists := s.ProjectChannels[project]; !exists {
		projectChan := make(chan models.Log, 100)
		shutdownChan := make(chan struct{})
		s.ProjectChannels[project] = projectChan
		s.shutdownChans[project] = shutdownChan
		go s.fanOutLogs(project, projectChan, shutdownChan)
		log.Printf("Started fan-out goroutine for project: %s", project)
	}
}

func (s *SSEService) UnregisterClient(clientID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if client, exists := s.Clients[clientID]; exists {
		client.mu.Lock()
		client.Connected = false
		client.mu.Unlock()

		// Schedule cleanup after a delay to handle reconnections
		go func() {
			time.Sleep(10 * time.Second) // Wait 10 seconds before cleanup
			s.mu.Lock()
			defer s.mu.Unlock()

			// Check if a client is still disconnected
			if client, exists := s.Clients[clientID]; exists {
				client.mu.Lock()
				stillDisconnected := !client.Connected
				client.mu.Unlock()

				if stillDisconnected {
					log.Printf("Cleaning up disconnected client after delay: %s", clientID)
					s.unregisterClientUnsafe(clientID, client)
				}
			}
		}()

		log.Printf("Marked client as disconnected: %s", clientID)
	} else {
		log.Printf("Client not found for unregistration: %s", clientID)
	}
}

// unregisterClientUnsafe must be called with mutex held
func (s *SSEService) unregisterClientUnsafe(clientID string, client *Client) {
	log.Printf("Unregistering client: %s", clientID)

	// Close client channel safely
	client.mu.Lock()
	if !client.closed {
		close(client.Channel)
		client.closed = true
		log.Printf("Closed client channel for %s", clientID)
	}
	client.mu.Unlock()

	// Remove client from global map
	delete(s.Clients, clientID)

	// Remove the client from a project-specific map
	if projectClients, ok := s.ProjectClients[client.Project]; ok {
		delete(projectClients, clientID)
		log.Printf("Removed client %s from project %s", clientID, client.Project)

		// If no more clients for this project, cleanup project resources
		if len(projectClients) == 0 {
			log.Printf("No more clients for project %s. Cleaning up project resources.", client.Project)

			// Signal shutdown to fan-out goroutine
			if shutdownChan, ok := s.shutdownChans[client.Project]; ok {
				close(shutdownChan)
				delete(s.shutdownChans, client.Project)
			}

			// Close project channel
			if projectChan, ok := s.ProjectChannels[client.Project]; ok {
				close(projectChan)
				delete(s.ProjectChannels, client.Project)
				log.Printf("Closed project channel for %s", client.Project)
			}

			// Clean up an empty project map
			delete(s.ProjectClients, client.Project)
		}
	}
	log.Printf("Successfully unregistered client: %s", clientID)
}

func (s *SSEService) fanOutLogs(project string, projectChan chan models.Log, shutdownChan chan struct{}) {
	log.Printf("Fan-out goroutine started for project: %s", project)
	defer log.Printf("Fan-out goroutine stopped for project: %s", project)

	for {
		select {
		case logEntry, ok := <-projectChan:
			if !ok {
				// Project channel closed, exit goroutine
				return
			}

			s.mu.RLock()
			clientsForProject, exists := s.ProjectClients[project]
			if !exists {
				s.mu.RUnlock()
				log.Printf("Project client map for %s no longer exists, stopping fan-out", project)
				return
			}

			// Send it to all clients
			for clientID, client := range clientsForProject {
				client.mu.Lock()
				if !client.closed {
					select {
					case client.Channel <- logEntry:
					default:
						log.Printf("Client %s channel full for project %s, dropping log", clientID, project)
					}
				}
				client.mu.Unlock()
			}
			s.mu.RUnlock()

		case <-shutdownChan:
			// Graceful shutdown requested
			return
		}
	}
}

func (s *SSEService) BroadcastLogs(project string, logEntry *models.Log) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if projectChan, ok := s.ProjectChannels[project]; ok {
		select {
		case projectChan <- *logEntry:
			// Successfully broadcast
		default:
			log.Printf("Project channel for %s full, dropping log for broadcast", project)
		}
	}
}

func (s *SSEService) GetClientChannel(clientID string) (chan models.Log, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, exists := s.Clients[clientID]
	if !exists {
		return nil, false
	}

	client.mu.Lock()
	defer client.mu.Unlock()

	if client.closed {
		return nil, false
	}

	return client.Channel, true
}

// UpdateClientActivity updates the last seen time for a client
func (s *SSEService) UpdateClientActivity(clientID string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if client, exists := s.Clients[clientID]; exists {
		client.mu.Lock()
		client.LastSeen = time.Now()
		client.mu.Unlock()
	}
}

// CleanupStaleClients removes clients that might be stalled
func (s *SSEService) CleanupStaleClients() {
	s.mu.Lock()
	defer s.mu.Unlock()
	staleThreshold := 5 * time.Minute
	now := time.Now()
	for clientID, client := range s.Clients {
		client.mu.Lock()
		isStale := now.Sub(client.LastSeen) > staleThreshold
		client.mu.Unlock()

		if isStale {
			log.Printf("Cleaning up stale client: %s", clientID)
			s.unregisterClientUnsafe(clientID, client)
		}
	}
}
