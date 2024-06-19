package cars_service

type CarModel struct {
	CarID         uint64       `json:"car_id"`
	Name          string       `json:"name"`
	Odo           uint32       `json:"odo"`
	Avatar        bool         `json:"avatar"`
	ServicesTotal uint32       `json:"services_total"`
	CarExtData    []CarExtData `json:"car_ext_data"`
}

type CarCreateModel struct {
	Name   string
	Odo    uint32
	Avatar bool
}

type CarExtData struct {
	Odo       uint32 `json:"odo"`
	NextOdo   uint32 `json:"next_odo"`
	GroupName string `json:"group_name"`
}

type rowGroup struct {
	Odo     uint32
	NextOdo uint32
}
