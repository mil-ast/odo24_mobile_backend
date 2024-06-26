package cars_service

import (
	"odo24_mobile_backend/api/services"
	"odo24_mobile_backend/db"
)

type CarsService struct{}

func NewCarsService() *CarsService {
	return &CarsService{}
}

func (srv *CarsService) GetCarsByUser(userID int64) ([]CarModel, error) {
	pg := db.Conn()

	rows, err := pg.Query(`SELECT c.car_id, c."name", c.odo, c.avatar, count(s.service_id) services_total
		FROM service_book.car c
		LEFT JOIN service_book.services s ON s.car_id = c.car_id
		WHERE c.user_id=$1
		GROUP BY c.car_id`, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var cars []CarModel

	for rows.Next() {
		var car CarModel
		err := rows.Scan(&car.CarID, &car.Name, &car.Odo, &car.Avatar, &car.ServicesTotal)
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

func (srv *CarsService) UpdateODO(carID int64, odo uint32) error {
	pg := db.Conn()

	_, err := pg.Exec(`UPDATE service_book.car SET odo=$1 WHERE car_id=$2`, odo, carID)
	if err != nil {
		return err
	}
	return nil
}

func (srv *CarsService) Delete(carID int64) error {
	pg := db.Conn()

	_, err := pg.Exec(`DELETE FROM service_book.car WHERE car_id=$1`, carID)
	if err != nil {
		return err
	}

	return nil
}

func (srv *CarsService) CheckOwner(carID, userID int64) error {
	pg := db.Conn()
	var dbUserID int64
	pg.QueryRow("SELECT user_id FROM service_book.car c WHERE car_id=$1", carID).Scan(&dbUserID)
	if dbUserID != userID {
		return services.ErrorNoPermission
	}
	return nil
}
