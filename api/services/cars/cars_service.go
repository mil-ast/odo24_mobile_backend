package cars_service

import (
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/mil-ast/odo24_mobile_backend/api/services"
	"github.com/mil-ast/odo24_mobile_backend/db"
)

type CarsService struct{}

func NewCarsService() *CarsService {
	return &CarsService{}
}

func (srv *CarsService) GetCarsByUser(userID uint64) ([]CarModel, error) {
	pg := db.Conn()

	rows, err := pg.Query(`SELECT 
		c.car_id, 
		c."name", 
		c.odo, 
		c.avatar,
		COALESCE(cnt.services_total, 0) AS services_total,
		ext.car_ext_data
		FROM service_book.car c
		LEFT JOIN (
			SELECT s.car_id, count(*) AS services_total
			FROM service_book.services s
			GROUP BY s.car_id
		) cnt ON cnt.car_id = c.car_id
		LEFT JOIN (
			SELECT 
				g.car_id,
				jsonb_agg(
					jsonb_build_object(
						'odo', g.odo,
						'next_odo', g.next_distance,
						'group_name', COALESCE(sg."name", 'Group ' || g.group_id::text)
					) ORDER BY g.group_id
				) AS car_ext_data
			FROM (
				SELECT DISTINCT ON (s.car_id, s.group_id)
					s.car_id, s.group_id, s.odo, s.next_distance
				FROM service_book.services s
				WHERE s.next_distance IS NOT NULL -- Перенесли сюда, чтобы NULL-записи не перекрывали актуальные данные
				ORDER BY s.car_id, s.group_id, s.odo DESC
			) g
			LEFT JOIN service_book.service_groups sg ON sg.group_id = g.group_id
			GROUP BY g.car_id
		) ext ON ext.car_id = c.car_id
		WHERE c.user_id = $1
		ORDER BY c.car_id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cars []CarModel
	for rows.Next() {
		var car CarModel
		var rawExt []byte
		if err := rows.Scan(&car.CarID, &car.Name, &car.Odo, &car.Avatar, &car.ServicesTotal, &rawExt); err != nil {
			return nil, err
		}

		if len(rawExt) > 0 {
			if err := json.Unmarshal(rawExt, &car.CarExtData); err != nil {
				return nil, err
			}
		}

		cars = append(cars, car)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cars, nil
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

func (srv *CarsService) UpdateODO(carID uint64, odo uint32) error {
	pg := db.Conn()

	_, err := pg.Exec(`UPDATE service_book.car SET odo=$1 WHERE car_id=$2`, odo, carID)
	if err != nil {
		return err
	}
	return nil
}

func (srv *CarsService) Delete(carID uint64) error {
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
	err := pg.QueryRow("SELECT user_id FROM service_book.car WHERE car_id=$1", carID).Scan(&dbUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return services.ErrorNoPermission
		}
		return err
	}

	if dbUserID != userID {
		return services.ErrorNoPermission
	}
	return nil
}
