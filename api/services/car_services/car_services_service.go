package car_services_service

import (
	"odo24_mobile_backend/api/services"
	"odo24_mobile_backend/db"
)

type CarServicesService struct{}

func NewCarServicesService() *CarServicesService {
	return &CarServicesService{}
}

func (srv *CarServicesService) GetServices(carID, groupID int64) ([]CarServiceModel, error) {
	pg := db.Conn()

	rows, err := pg.Query(`SELECT s.service_id,s.odo,s.next_distance,s.dt,s.description,s.price FROM service_book.services s WHERE s.car_id=$1 AND s.group_id=$2`, carID, groupID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []CarServiceModel

	for rows.Next() {
		var model CarServiceModel
		err := rows.Scan(&model.ServiceID, &model.Odo, &model.NextDistance, &model.Dt, &model.Description, &model.Price)
		if err != nil {
			return nil, err
		}

		result = append(result, model)
	}

	return result, nil
}

func (srv *CarServicesService) Create(body CarServiceCreateModel) (*CarServiceModel, error) {
	pg := db.Conn()

	var carServiceID int64
	err := pg.QueryRow(`INSERT INTO service_book.services (car_id,group_id,odo,next_distance,dt,description,price) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING service_id`, body.CarID, body.GroupID, body.Odo, body.NextDistance, body.Dt, body.Description, body.Price).Scan(&carServiceID)
	if err != nil {
		return nil, err
	}

	return &CarServiceModel{
		ServiceID:    carServiceID,
		Odo:          body.Odo,
		NextDistance: body.NextDistance,
		Dt:           body.Dt,
		Description:  body.Description,
		Price:        body.Price,
	}, nil
}

func (srv *CarServicesService) Update(body CarServiceUpdateModel) error {
	pg := db.Conn()

	_, err := pg.Exec(`UPDATE service_book.services SET odo=$1,next_distance=$2,dt=$3,description=$4,price=$5 WHERE service_id=$6`, body.Odo, body.NextDistance, body.Dt, body.Description, body.Price, body.ServiceID)
	if err != nil {
		return err
	}

	return nil
}

func (srv *CarServicesService) Delete(userID int64, serviceID int64) error {
	pg := db.Conn()

	_, err := pg.Exec(`DELETE FROM service_book.services WHERE service_id=$1`, serviceID)
	if err != nil {
		return err
	}

	return nil
}

func (srv *CarServicesService) CheckOwner(userID, serviceID int64) error {
	pg := db.Conn()
	var dbUserID int64
	pg.QueryRow("SELECT c.user_id FROM service_book.services s INNER JOIN service_book.car c ON c.car_id=s.car_id WHERE s.service_id=$1;", serviceID).Scan(&dbUserID)
	if dbUserID != userID {
		return services.ErrorNoPermission
	}
	return nil
}
