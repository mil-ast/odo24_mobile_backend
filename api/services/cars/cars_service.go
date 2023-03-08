package cars_service

import "odo24_mobile_backend/db"

type CarsService struct{}

func NewCarsService() *CarsService {
	return &CarsService{}
}

func (srv *CarsService) GetCarsByUser(userID int64) ([]CarModel, error) {
	pg := db.Conn()

	rows, err := pg.Query(`select c.car_id,c."name", c.odo, c.avatar from service_book.car c where c.user_id=$1`, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var cars []CarModel

	for rows.Next() {
		var car CarModel
		err := rows.Scan(&car.CarID, &car.Name, &car.Odo, &car.Avatar)
		if err != nil {
			return nil, err
		}

		cars = append(cars, car)
	}

	return cars, nil
}
