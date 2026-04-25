package account

import (
	"bandcash/internal/db"
	"bandcash/internal/utils"
)

type AccountData struct {
	Title                 string
	Breadcrumbs           []utils.Crumb
	CurrentLang           string
	UserID                string
	UserEmail             string
	SubscriptionSlots     int
	UsedSlots             int
	RemainingSlots        int
	IsLimitExceeded       bool
	HasAvailableGroupSlot bool
	HasActiveSubscription bool
	PaymentsEnabled       bool
	CheckoutQuantity      int
	ActiveTab             string
	Signals               map[string]any
	IsAuthenticated       bool
	IsSuperAdmin          bool
}

type SessionsData struct {
	Title            string
	Breadcrumbs      []utils.Crumb
	CurrentSessionID string
	UserEmail        string
	Sessions         []db.UserSession
	ActiveTab        string
	Signals          map[string]any
	IsAuthenticated  bool
	IsSuperAdmin     bool
}

type OverLimitGroup struct {
	Name string
}

type OverLimitData struct {
	Title                 string
	Breadcrumbs           []utils.Crumb
	SubscriptionCap       int
	OwnedGroups           int
	ExcessGroups          int
	Groups                []OverLimitGroup
	PaymentsEnabled       bool
	HasActiveSubscription bool
	CheckoutQuantity      int
	IsAuthenticated       bool
	IsSuperAdmin          bool
}
