package hub

import (
	"errors"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	"bandcash/internal/utils"
)

type Client struct {
	ID      string
	SSE     *datastar.ServerSentEventGenerator
	Signals chan struct{}
}

type SSEHub struct {
	mu      sync.RWMutex
	clients map[string]*Client
	views   map[string]string
}

var Hub = New()

func New() *SSEHub {
	return &SSEHub{
		clients: make(map[string]*Client),
		views:   make(map[string]string),
	}
}

func (h *SSEHub) AddClient(id string, sse *datastar.ServerSentEventGenerator) *Client {
	h.mu.Lock()
	defer h.mu.Unlock()

	client := &Client{
		ID:      id,
		SSE:     sse,
		Signals: make(chan struct{}, 1),
	}
	h.clients[id] = client
	return client
}

func (h *SSEHub) RemoveClient(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, id)
	delete(h.views, id)
}

func (h *SSEHub) GetClient(id string) (*Client, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	client, ok := h.clients[id]
	if !ok {
		return nil, errors.New("client not found")
	}
	return client, nil
}

func (h *SSEHub) Render(c echo.Context) error {
	clientID, err := utils.GetClientID(c)
	if err != nil {
		return err
	}

	h.mu.RLock()
	client, ok := h.clients[clientID]
	h.mu.RUnlock()

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

func (h *SSEHub) RenderAll() {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		select {
		case client.Signals <- struct{}{}:
		default:
		}
	}
}

func (h *SSEHub) PatchSignals(c echo.Context, signals any) error {
	clientID, err := utils.GetClientID(c)
	if err != nil {
		return err
	}

	client, err := h.GetClient(clientID)
	if err != nil {
		return err
	}
	return client.SSE.MarshalAndPatchSignals(signals)
}

func (h *SSEHub) SetView(id, view string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.views[id] = view
}

func (h *SSEHub) GetView(id string) (string, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	view, ok := h.views[id]
	return view, ok
}

func (h *SSEHub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, client := range h.clients {
		close(client.Signals)
	}
	h.clients = make(map[string]*Client)
	h.views = make(map[string]string)
}
