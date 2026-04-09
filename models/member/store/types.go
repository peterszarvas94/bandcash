package store

type GetMemberParams struct {
	ID      string `json:"id"`
	GroupID string `json:"group_id"`
}

type CreateMemberParams struct {
	ID          string `json:"id"`
	GroupID     string `json:"group_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateMemberParams struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ID          string `json:"id"`
	GroupID     string `json:"group_id"`
}

type DeleteMemberParams struct {
	ID      string `json:"id"`
	GroupID string `json:"group_id"`
}
