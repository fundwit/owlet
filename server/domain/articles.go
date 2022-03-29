package domain

import (
	"owlet/server/infra/persistence"
	"owlet/server/infra/sessions"

	"github.com/fundwit/go-commons/types"
)

type ArticleStatus int

const (
	ArticleStatusDraft     = ArticleStatus(0)
	ArticleStatusPublished = ArticleStatus(1)
)

type ArticleSource int

const (
	ArticleSourceOriginal  = ArticleSource(1)
	ArticleSourceTranslate = ArticleSource(2)
	ArticleSourceNote      = ArticleSource(3)
	ArticleSourceReference = ArticleSource(4)
)

type SyncChannel int

const (
	SyncChannelCnblog  = SyncChannel(0)
	SyncChannelCSDN    = SyncChannel(1)
	SyncChannelJianshu = SyncChannel(2)
)

type GenericType int

const (
	GenericTypeUnClassify = GenericType(1)
	GenericTypeIT         = GenericType(2)
	GenericTypeOther      = GenericType(3)
)

type ArticleMeta struct {
	ID    types.ID    `json:"id" gorm:"primary_key;type:BIGINT UNSIGNED NOT NULL"`
	Type  GenericType `json:"type" gorm:"type:TINYINT NOT NULL"`
	Title string      `json:"title" gorm:"type:NVARCHAR(255) NOT NULL"`

	UID        types.ID        `json:"uid" gorm:"type:BIGINT NOT NULL"`
	CreateTime types.Timestamp `json:"create_time" gorm:"type:DATETIME NOT NULL"`
	ModifyTime types.Timestamp `json:"modify_time" gorm:"type:DATETIME NOT NULL"`
	Status     ArticleStatus   `json:"status" gorm:"type:TINYINT NOT NULL DEFAULT '0'"`
	IsInvalid  bool            `json:"is_invalid" gorm:"type:TINYINT NOT NULL DEFAULT '0'"`

	Abstracts  string        `json:"abstracts" gorm:"type:NVARCHAR(1000) NULL"`
	Source     ArticleSource `json:"source" gorm:"type:TINYINT NOT NULL DEFAULT '0'"`
	IsElite    bool          `json:"is_elite" gorm:"type:TINYINT NOT NULL DEFAULT '0'"`
	IsTop      bool          `json:"is_top" gorm:"type:TINYINT NOT NULL DEFAULT '0'"`
	ViewNum    int           `json:"view_num" gorm:"type:INT NOT NULL DEFAULT '0'"`
	CommentNum int           `json:"comment_num" gorm:"type:INT NOT NULL DEFAULT '0'"`
}

type ArticleRecord struct {
	ArticleMeta

	Content string `json:"content" gorm:"type:TEXT NOT NULL"`
}

type ArticleMetaExt struct {
	ArticleMeta
	Tags []Tag `json:"tags"  gorm:"-"`
}

type ArticleDetail struct {
	ArticleRecord

	Tags []Tag `json:"tags"  gorm:"-"`
}

func (r *ArticleRecord) TableName() string {
	return "article"
}

type ArticleQuery struct {
	KeyWord string `form:"kw" binding:"omitempty,lte=200"`
	Page    int    `form:"page"` // base 1
}

type ArticlePatch struct {
	Title   string `json:"title"`
	Content string `json:"content"`

	Type    *GenericType   `json:"type" binding:"omitempty,oneof=1 2 3"`
	Status  *ArticleStatus `json:"status" binding:"omitempty,oneof=0 1"`
	Source  *ArticleSource `json:"source" binding:"omitempty,oneof=1 2 3 4"`
	IsElite *bool          `json:"is_elite"`
	IsTop   *bool          `json:"is_top"`
}

var (
	PageSize          = 10
	QueryArticlesFunc = QueryArticles
	DetailArticleFunc = DetailArticle
	PatchArticleFunc  = PatchArticle

	timestampFunc = types.CurrentTimestamp
)

func QueryArticles(q ArticleQuery, s *sessions.Session) ([]ArticleMetaExt, error) {
	offset := (q.Page - 1) * PageSize
	if offset < 0 {
		offset = 0
	}

	db := persistence.ActiveGormDB.Model(&ArticleRecord{}).
		Select("id, type, title, uid, create_time, modify_time, status, is_invalid, "+
			"abstracts, source, is_elite, is_top, view_num, comment_num").
		Where("is_invalid = 0 AND (status = 1 || uid = ?)", s.Identity.ID).
		Order("is_top DESC, create_time DESC").
		Offset(offset).
		Limit(PageSize)

	if len(q.KeyWord) > 0 {
		db.Where("title LIKE ?", "%"+q.KeyWord+"%")
	}

	var articleMetaExtList []ArticleMetaExt
	if err := db.Scan(&articleMetaExtList).Error; err != nil {
		return nil, err
	}

	if err := appendTags(articleMetaExtList, s); err != nil {
		return nil, err
	}

	return articleMetaExtList, nil
}

func PatchArticle(id types.ID, p *ArticlePatch, s *sessions.Session) error {
	if p == nil || (*p == ArticlePatch{}) {
		return nil
	}
	db := persistence.ActiveGormDB.Model(&ArticleRecord{}).Where("id = ?", id)

	changes := map[string]interface{}{}
	if p.Content != "" {
		changes["content"] = p.Content
	}
	if p.Title != "" {
		changes["title"] = p.Title
	}
	if p.Type != nil {
		changes["type"] = p.Type
	}
	if p.Source != nil {
		changes["source"] = p.Source
	}
	if p.Status != nil {
		changes["statue"] = p.Status
	}
	if p.IsTop != nil {
		changes["is_top"] = p.IsTop
	}
	if p.IsElite != nil {
		changes["is_elite"] = p.IsElite
	}
	if len(changes) > 0 {
		changes["modify_time"] = timestampFunc()
	}
	if err := db.Save(&changes).Error; err != nil {
		return err
	}
	return nil
}

func DetailArticle(id types.ID, s *sessions.Session) (*ArticleDetail, error) {
	var detail ArticleDetail
	db := persistence.ActiveGormDB.Model(&ArticleRecord{}).Select("*").Where("id = ?", id)
	if err := db.First(&detail).Error; err != nil {
		return nil, err
	}

	articleMetaExts := []ArticleMetaExt{{ArticleMeta: detail.ArticleMeta}}
	if err := appendTags(articleMetaExts, s); err != nil {
		return nil, err
	}
	detail.Tags = articleMetaExts[0].Tags
	return &detail, nil
}

func appendTags(articleMetaExtList []ArticleMetaExt, s *sessions.Session) error {
	articleNum := len(articleMetaExtList)
	if articleNum == 0 {
		return nil
	}

	articleIds := make([]types.ID, 0, articleNum)
	articleIdIndexMap := make(map[types.ID]int, articleNum)
	for idx := 0; idx < articleNum; idx++ {
		articleID := articleMetaExtList[idx].ID
		articleIds = append(articleIds, articleID)
		articleIdIndexMap[articleID] = idx
	}
	tagAssigns, err := QueryTagAssignmentsFunc(articleIds, s)
	if err != nil {
		return err
	}
	tagAssignNum := len(tagAssigns)
	if tagAssignNum == 0 {
		return nil
	}

	tagIdArticleIdsMap := map[types.ID][]types.ID{}
	tagIds := []types.ID{}
	for idx := 0; idx < tagAssignNum; idx++ {
		tagAssign := tagAssigns[idx]
		articleIDsOfTag, found := tagIdArticleIdsMap[tagAssign.TagID] // a copy of value?
		if !found {
			articleIDsOfTag = []types.ID{tagAssign.ResID}
			tagIds = append(tagIds, tagAssign.TagID)
		} else {
			articleIDsOfTag = append(articleIDsOfTag, tagAssign.ResID)
		}
		tagIdArticleIdsMap[tagAssign.TagID] = articleIDsOfTag
	}

	tags, err := QueryTagsFunc(TagQuery{IDs: tagIds}, s)
	if err != nil {
		return err
	}

	for _, tag := range tags {
		for _, articleId := range tagIdArticleIdsMap[tag.ID] {
			articleIndex := articleIdIndexMap[articleId]
			if articleMetaExtList[articleIndex].Tags == nil {
				articleMetaExtList[articleIndex].Tags = []Tag{}
			}
			articleMetaExtList[articleIndex].Tags = append(articleMetaExtList[articleIndex].Tags, tag)
		}
	}
	return nil
}
