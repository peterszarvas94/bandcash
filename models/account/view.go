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
	HasActiveSubscription bool
	ActiveTab             string
	Signals               map[string]any
	IsAuthenticated       bool
	IsSuperAdmin          bool
}

type SessionsData struct {
	Title            string
	Breadcrumbs      []utils.Crumb
	CurrentSessionID string
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
	Title           string
	Breadcrumbs     []utils.Crumb
	SubscriptionCap int
	OwnedGroups     int
	ExcessGroups    int
	Groups          []OverLimitGroup
	IsAuthenticated bool
	IsSuperAdmin    bool
}
