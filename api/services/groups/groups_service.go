package groups_service

import (
	"github.com/lib/pq"
	"github.com/mil-ast/odo24_mobile_backend/api/services"
	"github.com/mil-ast/odo24_mobile_backend/db"
)

type GroupsService struct{}

func NewGroupsService() *GroupsService {
	return &GroupsService{}
}

func (srv *GroupsService) GetGroupsByUser(userID uint64) ([]GroupModel, error) {
	pg := db.Conn()

	rows, err := pg.Query(`SELECT g.group_id,g."name",g.sort FROM service_book.service_groups g WHERE g.user_id=$1`, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var groups []GroupModel
	for rows.Next() {
		var group GroupModel
		err := rows.Scan(&group.GroupID, &group.Name, &group.Sort)
		if err != nil {
			return nil, err
		}

		groups = append(groups, group)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return groups, nil
}

func (srv *GroupsService) GetGroupsByCar(userID, carID uint64) ([]GroupModel, error) {
	groups, err := srv.GetGroupsByUser(userID)
	if err != nil {
		return nil, err
	}
	if len(groups) == 0 {
		return groups, nil
	}

	pg := db.Conn()

	// next_distance - дельта: через сколько км требуется сервис.
	// left_odo = (последний сервис этого авто в группе: service_odo + next_distance) - текущий пробег авто.
	rows, err := pg.Query(`SELECT g.group_id, g.odo, g.next_distance, c.odo
		FROM (
			SELECT DISTINCT ON (s.group_id) s.group_id, s.odo, s.next_distance
			FROM service_book.services s
			WHERE s.car_id = $1
			ORDER BY s.group_id, s.odo DESC
		) g
		CROSS JOIN service_book.car c
		WHERE g.next_distance IS NOT NULL AND c.car_id = $1`, carID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	leftOdoByGroup := make(map[uint64]uint32)
	serviceOdoByGroup := make(map[uint64]uint32)
	for rows.Next() {
		var groupID uint64
		var serviceOdo uint32
		var nextDistance uint32
		var carOdo uint32
		if err := rows.Scan(&groupID, &serviceOdo, &nextDistance, &carOdo); err != nil {
			return nil, err
		}

		nextRecommendedOdo := uint64(serviceOdo) + uint64(nextDistance)
		remaining := min(max(int64(nextRecommendedOdo)-int64(carOdo), 0), int64(^uint32(0)))
		leftOdoByGroup[groupID] = uint32(remaining)
		serviceOdoByGroup[groupID] = serviceOdo
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range groups {
		if leftOdo, ok := leftOdoByGroup[groups[i].GroupID]; ok {
			groups[i].GroupExtData = &GroupExtData{
				LeftOdo:    leftOdo,
				ServiceOdo: serviceOdoByGroup[groups[i].GroupID],
			}
		}
	}

	return groups, nil
}

func (srv *GroupsService) GetGroupsByIDs(groupIDs []uint64) ([]GroupModel, error) {
	if len(groupIDs) == 0 {
		return []GroupModel{}, nil
	}

	pg := db.Conn()
	rows, err := pg.Query(`SELECT g.group_id,g."name",g.sort FROM service_book.service_groups g WHERE g.group_id=ANY($1)`, pq.Array(groupIDs))
	if err != nil {
		return []GroupModel{}, err
	}

	defer rows.Close()

	var groups []GroupModel
	for rows.Next() {
		var group GroupModel
		err := rows.Scan(&group.GroupID, &group.Name, &group.Sort)
		if err != nil {
			return []GroupModel{}, err
		}
		groups = append(groups, group)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return groups, nil
}

func (srv *GroupsService) Create(userID uint64, groupBody GroupCreateModel) (*GroupModel, error) {
	pg := db.Conn()

	var sort uint32 = 0
	row := pg.QueryRow(`select max(sg.sort) from service_book.service_groups sg where sg.user_id=$1`, userID)
	if row != nil {
		row.Scan(&sort)
		sort += 1
	}

	var groupID uint64
	err := pg.QueryRow(`INSERT INTO service_book.service_groups (user_id,"name",sort) VALUES ($1,$2,$3) RETURNING group_id`, userID, groupBody.Name, sort).Scan(&groupID)
	if err != nil {
		return nil, err
	}

	return &GroupModel{
		GroupID: groupID,
		Name:    groupBody.Name,
		Sort:    sort,
	}, nil
}

func (srv *GroupsService) Update(userID uint64, groupBody GroupModel) error {
	pg := db.Conn()

	_, err := pg.Exec(`UPDATE service_book.service_groups SET "name"=$1 WHERE group_id=$2`, groupBody.Name, groupBody.GroupID)
	if err != nil {
		return err
	}

	return nil
}

func (srv *GroupsService) UpdateSort(userID uint64, groupIDs []int64) error {
	if len(groupIDs) == 0 {
		return nil
	}

	pg := db.Conn()

	var query = `UPDATE service_book.service_groups SET sort = s.idx
		FROM (SELECT * FROM unnest($2::int8[]) WITH ORDINALITY AS t(id, idx)) s
		WHERE user_id = $1 AND group_id = s.id`
	_, err := pg.Exec(query, userID, pq.Array(groupIDs))
	return err
}

func (srv *GroupsService) Delete(userID uint64, groupID uint64) error {
	pg := db.Conn()

	_, err := pg.Exec(`DELETE FROM service_book.service_groups WHERE group_id=$1`, groupID)
	if err != nil {
		return err
	}

	return nil
}

func (srv *GroupsService) CheckOwner(groupID, userID uint64) error {
	pg := db.Conn()
	var dbUserID uint64
	pg.QueryRow("SELECT user_id FROM service_book.service_groups c WHERE group_id=$1", groupID).Scan(&dbUserID)
	if dbUserID != userID {
		return services.ErrorNoPermission
	}
	return nil
}
