package groups_service

type GroupModel struct {
	GroupID      uint64        `json:"group_id"`
	Name         string        `json:"name"`
	Sort         uint32        `json:"sort"`
	GroupExtData *GroupExtData `json:"group_ext_data,omitempty"`
}

type GroupExtData struct {
	LeftOdo uint32 `json:"left_odo"`
}

type GroupCreateModel struct {
	Name string
	Sort uint32
}
