package groups_service

import (
	"odo24_mobile_backend/db"
)

type GroupsService struct{}

func NewGroupsService() *GroupsService {
	return &GroupsService{}
}

func (srv *GroupsService) GetGroupsByUser(userID int64) ([]GroupModel, error) {
	pg := db.Conn()

	rows, err := pg.Query(`select g.group_id,g."name",g.sort from service_book.service_groups g where g.user_id=$1`, userID)
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

	return groups, nil
}

func (srv *GroupsService) Create(userID int64, groupBody GroupCreateModel) (*GroupModel, error) {
	pg := db.Conn()

	var groupID int64
	err := pg.QueryRow(`INSERT INTO service_book.service_groups (user_id,"name",sort) VALUES ($1,$2,$3) RETURNING group_id`, userID, groupBody.Name, groupBody.Sort).Scan(&groupID)
	if err != nil {
		return nil, err
	}

	return &GroupModel{
		GroupID: groupID,
		Name:    groupBody.Name,
		Sort:    groupBody.Sort,
	}, nil
}
