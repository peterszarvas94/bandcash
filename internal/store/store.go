package store

import (
	"errors"
	"sync"

	"github.com/starfederation/datastar-go/datastar"
)

type Client struct {
	ID      string
	SSE     *datastar.ServerSentEventGenerator
	Signals chan struct{}
}

type Store struct {
	mu      sync.RWMutex
	clients map[string]*Client
}

func New() *Store {
	return &Store{
		clients: make(map[string]*Client),
	}
}

// Clients

func (s *Store) AddClient(id string, sse *datastar.ServerSentEventGenerator) *Client {
	s.mu.Lock()
	defer s.mu.Unlock()

	client := &Client{
		ID:      id,
		SSE:     sse,
		Signals: make(chan struct{}, 1),
	}
	s.clients[id] = client
	return client
}

func (s *Store) RemoveClient(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clients, id)
}

func (s *Store) GetClient(id string) (*Client, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	client, ok := s.clients[id]
	if !ok {
		return nil, errors.New("client not found")
	}
	return client, nil
}

func (s *Store) SignalClient(id string) error {
	s.mu.RLock()
	client, ok := s.clients[id]
	s.mu.RUnlock()

	if !ok {
		return errors.New("client not found")
	}

	select {
	case client.Signals <- struct{}{}:
		return nil
	default:
		return errors.New("signal channel full")
	}
}

// SignalAll signals all connected clients to update.
func (s *Store) SignalAll() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, client := range s.clients {
		select {
		case client.Signals <- struct{}{}:
		default:
			// Signal channel full, client will get next update
		}
	}
}

// Close closes all client signal channels to trigger graceful disconnection
func (s *Store) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, client := range s.clients {
		close(client.Signals)
	}
	s.clients = make(map[string]*Client)
}
