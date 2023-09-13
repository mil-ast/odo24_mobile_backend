package groups_service

type GroupModel struct {
	GroupID int64  `json:"group_id"`
	Name    string `json:"name"`
	Sort    uint32 `json:"sort"`
}

type GroupCreateModel struct {
	Name string
	Sort uint32
}
