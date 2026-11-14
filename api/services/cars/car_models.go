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
	ServiceOdo uint32 `json:"odo"`
	NextOdo    uint32 `json:"next_odo"`
	GroupName  string `json:"group_name"`
}

type CarExtDataV2 struct {
	ServiceOdo uint32 `json:"service_odo"`
	LeftOdo    uint32 `json:"left_odo"`
	GroupName  string `json:"group_name"`
}

type CarModelV2 struct {
	CarID         uint64         `json:"car_id"`
	Name          string         `json:"name"`
	Odo           uint32         `json:"odo"`
	Avatar        bool           `json:"avatar"`
	ServicesTotal uint32         `json:"services_total"`
	CarExtData    []CarExtDataV2 `json:"car_ext_data"`
}
