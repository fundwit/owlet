package domain

import (
	"errors"
	"owlet/server/infra/idgen"
	"owlet/server/infra/persistence"
	"owlet/server/infra/sessions"

	"github.com/fundwit/go-commons/types"
	"github.com/sony/sonyflake"
	"gorm.io/gorm"
)

type Tag struct {
	ID types.ID `json:"id" gorm:"primary_key;type:BIGINT UNSIGNED NOT NULL"`

	Name  string `json:"name" gorm:"column:tname;type:NVARCHAR(255) NOT NULL"`
	Note  string `json:"note" gorm:"type:NVARCHAR(255) NULL"`
	Image string `json:"image" gorm:"column:img;type:NVARCHAR(255) NULL"`
}

func (r *Tag) TableName() string {
	return "tag"
}

type TagWithStat struct {
	Tag

	Count int `json:"count"`
}

var (
	tagIdWorker = sonyflake.NewSonyflake(sonyflake.Settings{})
	tagIdFunc   = func() types.ID {
		return idgen.NextID(tagIdWorker)
	}

	QueryTagsFunc         = QueryTags
	QueryTagsWithStatFunc = QueryTagsWithStat
	ExtendTagsStatFunc    = ExtendTagsStat
	FindOrCreateTagFunc   = FindOrCreateTag
)

type TagQuery struct {
	IDs []types.ID `form:"id" binding:"omitempty"`
}

type TagCreate struct {
	Name string `json:"name" binding:"required"`
}

func QueryTagsWithStat(s *sessions.Session) ([]TagWithStat, error) {
	tags, err := QueryTagsFunc(TagQuery{}, s)
	if err != nil {
		return nil, err
	}
	return ExtendTagsStatFunc(tags, s)
}

func QueryTags(q TagQuery, s *sessions.Session) ([]Tag, error) {
	tags := []Tag{}
	db := persistence.ActiveGormDB.WithContext(s.Context).Model(&Tag{})
	if len(q.IDs) > 0 {
		db.Where("id IN ?", q.IDs)
	}
	if err := db.Scan(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

func FindOrCreateTag(tx *gorm.DB, q *TagCreate, s *sessions.Session) (*Tag, error) {
	var tag Tag
	err := tx.Model(&Tag{}).Where("tname LIKE ?", q.Name).First(&tag).Error
	if err == nil {
		return &tag, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	tag = Tag{
		ID:   tagIdFunc(),
		Name: q.Name,
	}
	if err := tx.Create(&tag).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

func ExtendTagsStat(tags []Tag, s *sessions.Session) ([]TagWithStat, error) {
	cap := len(tags)
	if cap == 0 {
		return []TagWithStat{}, nil
	}

	tagCount := map[types.ID]int{}
	tagIds := make([]types.ID, 0, cap)
	for _, tag := range tags {
		tagIds = append(tagIds, tag.ID)
		tagCount[tag.ID] = 0
	}

	tagsStat := []TagWithStat{}
	db := persistence.ActiveGormDB.WithContext(s.Context).Debug().
		Table("tag_assign").Select("res_id AS id, count(*) AS count").
		Where("res_id IN ? AND res_type = ?", tagIds, 0).
		Group("res_id")

	if err := db.Scan(&tagsStat).Error; err != nil {
		return nil, err
	}

	for _, ts := range tagsStat {
		tagCount[ts.ID] = ts.Count
	}

	tagsWithStat := make([]TagWithStat, 0, cap)
	for _, tag := range tags {
		tagsWithStat = append(tagsWithStat, TagWithStat{Tag: tag, Count: tagCount[tag.ID]})
	}

	return tagsWithStat, nil
}
