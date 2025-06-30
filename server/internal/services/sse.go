package services

import (
	"log"
	"server/internal/models"
	"sync"
)

type Client struct {
	ID      string // Unique client ID
	Project string // Project associated with the client
	Channel chan models.Log
	closed  bool       // Track if a client is closed
	mu      sync.Mutex // Protect a closed state
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
		log.Printf("Client %s already exists, cleaning up old registration", clientID)
		s.unregisterClientUnsafe(clientID, existingClient)
	}

	// Create a new client
	client := &Client{
		ID:      clientID,
		Project: project,
		Channel: make(chan models.Log, 100),
		closed:  false,
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
		s.unregisterClientUnsafe(clientID, client)
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
						// Successfully sent
					default:
						log.Printf("Client %s channel full for project %s, dropping log", clientID, project)
						// Consider cleaning up stalled clients here
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

// CleanupStaleClients removes clients that might be stalled
func (s *SSEService) CleanupStaleClients() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for clientID, client := range s.Clients {
		client.mu.Lock()
		if client.closed {
			// Client is already marked as closed, clean it up
			s.unregisterClientUnsafe(clientID, client)
		}
		client.mu.Unlock()
	}
}
