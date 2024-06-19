package car_services_service

type CarServiceModel struct {
	ServiceID    uint64  `json:"service_id"`
	Odo          *uint32 `json:"odo"`
	NextDistance *uint32 `json:"next_distance"`
	Dt           string  `json:"dt"`
	Description  *string `json:"description"`
	Price        *uint32 `json:"price"`
}

type CarServiceCreateModel struct {
	CarID        uint64
	GroupID      uint64
	Odo          *uint32
	NextDistance *uint32
	Dt           string
	Description  *string
	Price        *uint32
}
type CarServiceUpdateModel struct {
	ServiceID    uint64
	Odo          *uint32
	NextDistance *uint32
	Dt           string
	Description  *string
	Price        *uint32
}
