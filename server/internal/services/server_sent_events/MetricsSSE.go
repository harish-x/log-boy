package serversentevents

import (
	"log"
	"server/internal/models"
	"sync"
	"time"
)

type MetricsClient struct {
	ID        string
	Project   string
	Channel   chan models.Metrics
	closed    bool
	mu        sync.Mutex
	LastSeen  time.Time
	Connected bool
}

type SSEMetricsService struct {
	mu              sync.RWMutex
	Clients         map[string]*MetricsClient
	ProjectClients  map[string]map[string]*MetricsClient
	ProjectChannels map[string]chan models.Metrics
	shutdownChans   map[string]chan struct{}
}

func NewSSEMetricsService() *SSEMetricsService {
	return &SSEMetricsService{
		Clients:         make(map[string]*MetricsClient),
		ProjectClients:  make(map[string]map[string]*MetricsClient),
		ProjectChannels: make(map[string]chan models.Metrics),
		shutdownChans:   make(map[string]chan struct{}),
	}
}

func (s *SSEMetricsService) RegisterMetricsClient(clientID, project string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existingClient, exists := s.Clients[clientID]; exists {
		existingClient.mu.Lock()
		timeSincelastSeen := time.Since(existingClient.LastSeen)
		existingClient.mu.Unlock()

		if timeSincelastSeen < 5*time.Second && existingClient.Project == project {
			existingClient.mu.Lock()
			existingClient.Connected = true
			existingClient.LastSeen = time.Now()
			existingClient.mu.Unlock()
			return
		}
		s.unregisterClientUnsafe(clientID, existingClient)
	}

	client := &MetricsClient{
		ID:        clientID,
		Project:   project,
		Channel:   make(chan models.Metrics),
		closed:    false,
		LastSeen:  time.Now(),
		Connected: true,
	}

	s.Clients[clientID] = client

	if _, exists := s.ProjectClients[project]; !exists {
		s.ProjectClients[project] = make(map[string]*MetricsClient)
	}
	s.ProjectClients[project][clientID] = client

	if _, exist := s.ProjectChannels[project]; !exist {
		projectChan := make(chan models.Metrics, 100)
		shutdownChans := make(chan struct{})
		s.ProjectChannels[project] = projectChan
		s.shutdownChans[project] = shutdownChans
		go s.fanOutMetrics(project, projectChan, shutdownChans)
	}
}

func (s *SSEMetricsService) UnRegisterMetricsClient(clientId string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if client, exists := s.Clients[clientId]; exists {
		client.mu.Lock()
		client.Connected = false
		client.mu.Unlock()

		go func() {
			time.Sleep(10 * time.Second)
			s.mu.Lock()
			defer s.mu.Unlock()

			if client, exists := s.Clients[clientId]; exists {
				client.mu.Lock()
				stillDisconnected := !client.Connected
				client.mu.Unlock()

				if stillDisconnected {
					s.unregisterClientUnsafe(clientId, client)
				}
			}
		}()

	}
}

func (s *SSEMetricsService) unregisterClientUnsafe(clientID string, client *MetricsClient) {
	client.mu.Lock()
	if !client.closed {
		close(client.Channel)
		client.closed = true
	}
	client.mu.Unlock()

	// Remove client from global map
	delete(s.Clients, clientID)

	// Remove the client from a project-specific map
	if projectsClients, ok := s.ProjectClients[client.Project]; ok {
		delete(projectsClients, clientID)

		if len(projectsClients) == 0 {

			if shutdownChan, ok := s.shutdownChans[client.Project]; ok {
				close(shutdownChan)
				delete(s.shutdownChans, client.Project)
			}

			//close projevt ProjectChannels
			if projectChan, ok := s.ProjectChannels[client.Project]; ok {
				close(projectChan)
				delete(s.ProjectChannels, client.Project)
			}

			delete(s.ProjectClients, client.Project)
		}
	}
}

func (s *SSEMetricsService) fanOutMetrics(project string, projectChan chan models.Metrics, shutdownChan chan struct{}) {
	for {
		select {
		case metrics, ok := <-projectChan:
			if !ok {
				return
			}
			s.mu.RLock()
			clientsForProject, exists := s.ProjectClients[project]
			if !exists {
				s.mu.RUnlock()
				return
			}

			for clientID, client := range clientsForProject {
				client.mu.Lock()
				if !client.closed {
					select {
					case client.Channel <- metrics:
					default:
						log.Printf("Client %s channel full for project %s, dropping log", clientID, project)
					}
				}
				client.mu.Unlock()
			}
			s.mu.RUnlock()
		case <-shutdownChan:
			return
		}
	}
}

func (s *SSEMetricsService) BroadcastMetrics(project string, metrics *models.Metrics) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if projectChan, ok := s.ProjectChannels[project]; ok {
		select {
		case projectChan <- *metrics:
		default:
			log.Printf("Project channel for %s full, dropping log for broadcast", project)
		}
	}
}

func (s *SSEMetricsService) GetMetricsClientChannel(clientId string) (chan models.Metrics, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, exists := s.Clients[clientId]

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

func (s *SSEMetricsService) UpdateMetricsClientActivity(clientID string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if client, exists := s.Clients[clientID]; exists {
		client.mu.Lock()
		client.LastSeen = time.Now()
		client.mu.Unlock()
	}
}

func (s *SSEMetricsService) CleanupMetricsStaleClients() {
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
