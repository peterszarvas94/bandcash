package utils

import (
	"errors"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"
)

type Client struct {
	ID  string
	SSE *datastar.ServerSentEventGenerator
}

type Hub struct {
	mu      sync.RWMutex
	clients map[string]*Client
}

var SSEHub = NewHub()

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]*Client),
	}
}

func (h *Hub) AddClient(id string, sse *datastar.ServerSentEventGenerator) *Client {
	h.mu.Lock()
	defer h.mu.Unlock()

	client := &Client{
		ID:  id,
		SSE: sse,
	}
	h.clients[id] = client
	return client
}

func (h *Hub) RemoveClient(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, id)
}

func (h *Hub) GetClient(id string) (*Client, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	client, ok := h.clients[id]
	if !ok {
		return nil, errors.New("client not found")
	}
	return client, nil
}

func (h *Hub) PatchHTML(c echo.Context, html string) error {
	clientID, err := GetClientID(c)
	if err != nil {
		return err
	}

	client, err := h.GetClient(clientID)
	if err != nil {
		return err
	}

	return client.SSE.PatchElements(html)
}

func (h *Hub) PatchSignals(c echo.Context, signals any) error {
	clientID, err := GetClientID(c)
	if err != nil {
		return err
	}

	client, err := h.GetClient(clientID)
	if err != nil {
		return err
	}
	return client.SSE.MarshalAndPatchSignals(signals)
}

func (h *Hub) Redirect(c echo.Context, url string) error {
	clientID, err := GetClientID(c)
	if err != nil {
		return err
	}

	client, err := h.GetClient(clientID)
	if err != nil {
		return err
	}
	return client.SSE.Redirect(url)
}

func (h *Hub) ExecuteScript(c echo.Context, script string) error {
	clientID, err := GetClientID(c)
	if err != nil {
		return err
	}

	client, err := h.GetClient(clientID)
	if err != nil {
		return err
	}

	return client.SSE.ExecuteScript(script)
}

func (h *Hub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients = make(map[string]*Client)
}
