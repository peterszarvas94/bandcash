package utils

import (
	"context"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

type Notification struct {
	ID      string
	Kind    string
	Message string
	Created time.Time
}

type clientNotifications struct {
	queued []Notification
	active []Notification
}

type notificationStore struct {
	mu      sync.Mutex
	clients map[string]clientNotifications
}

const notificationLifetime = 5 * time.Second

var Notifications = &notificationStore{clients: map[string]clientNotifications{}}

type notificationsContextKey struct{}

func (s *notificationStore) Add(clientID string, n Notification) {
	s.mu.Lock()
	defer s.mu.Unlock()
	state := s.clients[clientID]
	state.queued = append(state.queued, n)
	s.clients[clientID] = state
}

func (s *notificationStore) Drain(clientID string) []Notification {
	return s.DrainForRender(clientID, false)
}

func (s *notificationStore) DrainForRender(clientID string, includeActive bool) []Notification {
	s.mu.Lock()
	defer s.mu.Unlock()
	state, ok := s.clients[clientID]
	if !ok {
		return nil
	}

	now := time.Now()
	active := make([]Notification, 0, len(state.active))
	for _, item := range state.active {
		if now.Sub(item.Created) < notificationLifetime {
			active = append(active, item)
		}
	}

	items := make([]Notification, 0, len(active)+len(state.queued))
	if includeActive {
		items = append(items, active...)
	}
	for _, item := range state.queued {
		if now.Sub(item.Created) < notificationLifetime {
			items = append(items, item)
		}
	}

	if len(items) == 0 {
		delete(s.clients, clientID)
		return nil
	}

	state.active = items
	state.queued = nil
	s.clients[clientID] = state

	result := make([]Notification, len(items))
	copy(result, items)
	return result
}

func Notify(c echo.Context, kind, message string) {
	if message == "" {
		return
	}
	clientID := EnsureClientID(c)
	Notifications.Add(clientID, Notification{
		ID:      GenerateID("ntf"),
		Kind:    kind,
		Message: message,
		Created: time.Now(),
	})
}

func WithNotifications(ctx context.Context, items []Notification) context.Context {
	return context.WithValue(ctx, notificationsContextKey{}, items)
}

func NotificationsFromContext(ctx context.Context) []Notification {
	items, ok := ctx.Value(notificationsContextKey{}).([]Notification)
	if !ok {
		return nil
	}
	return items
}
