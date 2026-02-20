package utils

import (
	"context"
	"sync"

	"github.com/labstack/echo/v4"
)

type Notification struct {
	ID      string
	Kind    string
	Message string
}

type notificationStore struct {
	mu      sync.Mutex
	clients map[string][]Notification
}

var Notifications = &notificationStore{clients: map[string][]Notification{}}

type notificationsContextKey struct{}

func (s *notificationStore) Add(clientID string, n Notification) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[clientID] = append(s.clients[clientID], n)
}

func (s *notificationStore) Drain(clientID string) []Notification {
	s.mu.Lock()
	defer s.mu.Unlock()
	items := s.clients[clientID]
	if len(items) == 0 {
		return nil
	}
	delete(s.clients, clientID)
	return items
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
