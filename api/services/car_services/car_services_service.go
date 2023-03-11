package car_services_service

import (
	"odo24_mobile_backend/api/services"
	"odo24_mobile_backend/db"
)

type CarServicesService struct{}

func NewCarServicesService() *CarServicesService {
	return &CarServicesService{}
}

func (srv *CarServicesService) GetServices(userID, carID, groupID int64) ([]CarServiceModel, error) {
	pg := db.Conn()

	err := srv.checkGroupOwner(groupID, userID)
	if err != nil {
		return nil, err
	}

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

func (srv *CarServicesService) checkGroupOwner(groupID, userID int64) error {
	pg := db.Conn()
	var dbUserID int64
	pg.QueryRow("SELECT user_id FROM service_book.service_groups c WHERE group_id=$1", groupID).Scan(&dbUserID)
	if dbUserID != userID {
		return services.ErrorNoPermission
	}
	return nil
}
