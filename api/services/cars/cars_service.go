package cars_service

import (
	"odo24_mobile_backend/api/services"
	"odo24_mobile_backend/db"

	"github.com/lib/pq"
)

type CarsService struct{}

func NewCarsService() *CarsService {
	return &CarsService{}
}

func (srv *CarsService) GetCarsByUser(userID uint64) ([]CarModel, error) {
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

func (srv *CarsService) GetCarNextServiceInformation(carIDs []uint64) (map[uint64]map[uint64]rowGroup, error) {
	if len(carIDs) == 0 {
		return nil, nil
	}
	pg := db.Conn()

	rows, err := pg.Query(`SELECT g.car_id, g.group_id, g.odo, g.next_distance FROM (
    SELECT s.car_id, s.group_id, s.odo, s.next_distance, row_number()
    OVER (PARTITION BY s.car_id, s.group_id ORDER BY s.odo DESC) AS rownum
		FROM service_book.services s
		WHERE s.car_id = ANY($1)
	) g
	WHERE rownum=1 and g.next_distance is not null;`, pq.Array(carIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rowsMap := make(map[uint64]map[uint64]rowGroup)
	var mapGroupIDs = make(map[uint64]struct{})
	for rows.Next() {
		var row struct {
			CarID   uint64
			GroupID uint64
			Odo     uint32
			NextOdo uint32
		}
		err := rows.Scan(&row.CarID, &row.GroupID, &row.Odo, &row.NextOdo)
		if err != nil {
			return nil, err
		}
		if _, ok := rowsMap[row.CarID]; !ok {
			rowsMap[row.CarID] = make(map[uint64]rowGroup)
		}

		rowsMap[row.CarID][row.GroupID] = rowGroup{
			Odo:     row.Odo,
			NextOdo: row.NextOdo,
		}
		mapGroupIDs[row.GroupID] = struct{}{}
	}

	return rowsMap, nil
}

func (srv *CarsService) Create(userID uint64, carBody CarCreateModel) (*CarModel, error) {
	pg := db.Conn()

	var carID uint64
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

func (srv *CarsService) CheckOwner(carID, userID uint64) error {
	pg := db.Conn()
	var dbUserID uint64
	pg.QueryRow("SELECT user_id FROM service_book.car c WHERE car_id=$1", carID).Scan(&dbUserID)
	if dbUserID != userID {
		return services.ErrorNoPermission
	}
	return nil
}
