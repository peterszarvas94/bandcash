package auth

import "bandcash/internal/utils"

type AuthPageData struct {
	Title       string
	Breadcrumbs []utils.Crumb
	UserEmail   string
}
