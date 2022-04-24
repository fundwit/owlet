package domain

import (
	"owlet/server/infra/idgen"
	"owlet/server/infra/persistence"
	"owlet/server/infra/sessions"

	"github.com/fundwit/go-commons/types"
	"github.com/sony/sonyflake"
	"gorm.io/gorm"
)

type ResType int

const (
	ResTypeArticle = ResType(0)
)

type TagAssignRelation struct {
	ResID types.ID `json:"resId" form:"resId" binding:"required"`
	TagID types.ID `json:"tagId" form:"tagId" binding:"required"`
}

type TagAssignCreate struct {
	ResID   types.ID `json:"resId" binding:"required"`
	TagName string   `json:"tagName" binding:"required"`
}

type TagAssignment struct {
	ID types.ID `json:"id" gorm:"primary_key;type:BIGINT UNSIGNED NOT NULL"`

	ResID   types.ID `json:"resId" gorm:"column:res_id;type:BIGINT NOT NULL;unique_index:uni_res_tag"`
	TagID   types.ID `json:"tagId" gorm:"column:tag;type:BIGINT NOT NULL;unique_index:uni_res_tag"`
	ResType ResType  `json:"resType" gorm:"column:res_type;type:TINYINT NOT NULL DEFAULT '0';unique_index:uni_res_tag"`

	TagOrder int `json:"tagOrder" gorm:"column:tag_order;type:INT NOT NULL DEFAULT '0'"`
}

type TagAssignCreateResponse struct {
	TagAssignment

	TagName  string `json:"tagName"`
	TagNote  string `json:"tagNote"`
	TagImage string `json:"tagImage"`
}

func (r *TagAssignment) TableName() string {
	return "tag_assign"
}

var (
	tagAssignIdWorker = sonyflake.NewSonyflake(sonyflake.Settings{})
	tagAssignIdFunc   = func() types.ID {
		return idgen.NextID(tagAssignIdWorker)
	}

	QueryTagAssignmentsFunc      = QueryTagAssignments
	ClearArticleTagAssignsFunc   = ClearArticleTagAssigns
	CreateTagAssignFunc          = CreateTagAssign
	DeleteTagAssignWithQueryFunc = DeleteTagAssignWithQuery
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

func CreateTagAssign(c *TagAssignCreate, s *sessions.Session) (*TagAssignCreateResponse, error) {
	assign := TagAssignment{
		ID:      tagAssignIdFunc(),
		ResID:   c.ResID,
		ResType: ResTypeArticle,
	}

	var resp *TagAssignCreateResponse

	dbErr := persistence.ActiveGormDB.Transaction(func(tx *gorm.DB) error {
		if err := checkPermFunc(tx, c.ResID, s); err != nil {
			return err
		}

		tag, err := FindOrCreateTagFunc(tx, &TagCreate{Name: c.TagName}, s)
		if err != nil {
			return err
		}
		assign.TagID = tag.ID

		if err := tx.Create(&assign).Error; err != nil {
			return err
		}
		resp = &TagAssignCreateResponse{
			TagAssignment: assign, TagName: tag.Name, TagNote: tag.Note, TagImage: tag.Image}
		return nil
	})

	return resp, dbErr
}

func DeleteTagAssignWithQuery(c *TagAssignRelation, s *sessions.Session) error {
	return persistence.ActiveGormDB.Transaction(func(tx *gorm.DB) error {
		if err := checkPermFunc(tx, c.ResID, s); err != nil {
			return err
		}
		if err := tx.Where("res_id = ? AND tag = ? AND res_type = 0", c.ResID, c.TagID).
			Delete(&TagAssignment{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func ClearArticleTagAssigns(tx *gorm.DB, articleId types.ID, s *sessions.Session) error {
	if err := tx.Where("res_id = ? AND res_type = 0", articleId).Delete(&TagAssignment{}).Error; err != nil {
		return err
	}
	return nil
}
