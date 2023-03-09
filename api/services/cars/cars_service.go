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

func (srv *CarsService) Create(userID int64, carBody CarCreateModel) (*CarModel, error) {
	pg := db.Conn()

	var carID int64
	err := pg.QueryRow(`INSERT INTO service_book.car (user_id,"name",odo,avatar) VALUES ($1,$2,$3,$4) RETURNING car_id`, userID, carBody.Name, carBody.Odo, carBody.Avatar).Scan(&carID)
	if err != nil {
		return nil, err
	}

	return &CarModel{
		CarID:  carID,
		Name:   carBody.Name,
		Odo:    carBody.Odo,
		Avatar: carBody.Avatar,
	}, nil
}

func (srv *CarsService) Update(carBody CarModel) error {
	pg := db.Conn()

	_, err := pg.Exec(`UPDATE service_book.car SET "name"=$1,odo=$2,avatar=$3 WHERE car_id=$4`, carBody.Name, carBody.Odo, carBody.Avatar, carBody.CarID)
	if err != nil {
		return err
	}

	return nil
}
