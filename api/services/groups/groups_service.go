package groups_service

import (
	"fmt"
	"odo24_mobile_backend/api/services"
	"odo24_mobile_backend/db"
	"strings"
)

type GroupsService struct{}

func NewGroupsService() *GroupsService {
	return &GroupsService{}
}

func (srv *GroupsService) GetGroupsByUser(userID int64) ([]GroupModel, error) {
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

	return groups, nil
}

func (srv *GroupsService) Create(userID int64, groupBody GroupCreateModel) (*GroupModel, error) {
	pg := db.Conn()

	var sort uint32 = 0
	row := pg.QueryRow(`select max(sg.sort) from service_book.service_groups sg where sg.user_id=$1`, userID)
	if row != nil {
		row.Scan(&sort)
		sort += 1
	}

	var groupID int64
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

func (srv *GroupsService) Update(userID int64, groupBody GroupModel) error {
	pg := db.Conn()

	_, err := pg.Exec(`UPDATE service_book.service_groups SET "name"=$1 WHERE group_id=$2`, groupBody.Name, groupBody.GroupID)
	if err != nil {
		return err
	}

	return nil
}

func (srv *GroupsService) UpdateSort(userID int64, groupIDs []int64) error {
	pg := db.Conn()

	sortedIDs := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(groupIDs)), ","), "[]")

	var query = fmt.Sprintf(`UPDATE service_book.service_groups SET sort=s.idx
		FROM (SELECT * FROM LATERAL unnest(array[%s]) WITH ORDINALITY AS t(id,idx)) s
		WHERE user_id=$1 AND group_id=s.id`, sortedIDs)

	_, err := pg.Exec(query, userID)
	return err
}

func (srv *GroupsService) Delete(userID int64, groupID int64) error {
	pg := db.Conn()

	_, err := pg.Exec(`DELETE FROM service_book.service_groups WHERE group_id=$1`, groupID)
	if err != nil {
		return err
	}

	return nil
}

func (srv *GroupsService) CheckOwner(groupID, userID int64) error {
	pg := db.Conn()
	var dbUserID int64
	pg.QueryRow("SELECT user_id FROM service_book.service_groups c WHERE group_id=$1", groupID).Scan(&dbUserID)
	if dbUserID != userID {
		return services.ErrorNoPermission
	}
	return nil
}
