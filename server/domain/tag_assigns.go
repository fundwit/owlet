package domain

import (
	"owlet/server/infra/persistence"
	"owlet/server/infra/sessions"

	"github.com/fundwit/go-commons/types"
)

type ResType int

const (
	ResTypeArticle = ResType(0)
)

type TagAssignCreate struct {
	ResID types.ID `json:"resId"`
	TagID types.ID `json:"tagId"`
}

type TagAssignment struct {
	ID types.ID `json:"id" gorm:"primary_key;type:BIGINT UNSIGNED NOT NULL"`

	ResID   types.ID `json:"resId" gorm:"column:res_id;type:BIGINT NOT NULL;unique_index:uni_res_tag"`
	TagID   types.ID `json:"tagId" gorm:"column:tag;type:BIGINT NOT NULL;unique_index:uni_res_tag"`
	ResType ResType  `json:"restype" gorm:"column:res_type;type:TINYINT NOT NULL DEFAULT '0';unique_index:uni_res_tag"`

	TagOrder int `json:"tagOrder" gorm:"column:tag_order;type:INT NOT NULL DEFAULT '0'"`
}

func (r *TagAssignment) TableName() string {
	return "tag_assign"
}

var (
	QueryTagAssignmentsFunc = QueryTagAssignments
)

func QueryTagAssignments(resIds []types.ID, s *sessions.Session) ([]TagAssignment, error) {
	tagAssigns := []TagAssignment{}
	if len(resIds) == 0 {
		return tagAssigns, nil
	}

	db := persistence.ActiveGormDB.WithContext(s.Context).Model(&TagAssignment{}).
		Where("res_id IN ? AND res_type = 0", resIds)

	if err := db.Scan(&tagAssigns).Error; err != nil {
		return nil, err
	}
	return tagAssigns, nil
}

// var (
// 	idWorker = sonyflake.NewSonyflake(sonyflake.Settings{})
// )

// func CreateTagAssign(c *TagAssignCreate, s *sessions.Session) error {
// 	assign := TagAssignment{
// 		ID:      idgen.NextID(idWorker),
// 		ResID:   c.ResID,
// 		TagID:   c.TagID,
// 		ResType: ResTypeArticle,
// 	}

// }
