package cars_service

type CarModel struct {
	CarID  int64  `json:"car_id"`
	Name   string `json:"name"`
	Odo    uint32 `json:"odo"`
	Avatar bool   `json:"avatar"`
}
