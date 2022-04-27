package domain

import (
	"errors"
	"owlet/server/infra/fail"
	"owlet/server/infra/idgen"
	"owlet/server/infra/persistence"
	"owlet/server/infra/sessions"
	"owlet/server/misc"

	"github.com/fundwit/go-commons/types"
	"github.com/sirupsen/logrus"
	"github.com/sony/sonyflake"
	"gorm.io/gorm"
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
	ID types.ID `json:"id" gorm:"primary_key;type:BIGINT UNSIGNED NOT NULL"`

	Title     string `json:"title" gorm:"type:NVARCHAR(255) NOT NULL"`
	Abstracts string `json:"abstracts" gorm:"type:NVARCHAR(1000) NULL"`

	Type   GenericType   `json:"type" gorm:"type:TINYINT NOT NULL"`
	Status ArticleStatus `json:"status" gorm:"type:TINYINT NOT NULL DEFAULT '0'"`
	Source ArticleSource `json:"source" gorm:"type:TINYINT NOT NULL DEFAULT '0'"`

	UID        types.ID        `json:"uid" gorm:"type:BIGINT NOT NULL"`
	CreateTime types.Timestamp `json:"create_time" gorm:"type:DATETIME NOT NULL"`
	ModifyTime types.Timestamp `json:"modify_time" gorm:"type:DATETIME NOT NULL"`

	IsInvalid  bool `json:"is_invalid" gorm:"type:TINYINT NOT NULL DEFAULT '0'"`
	IsElite    bool `json:"is_elite" gorm:"type:TINYINT NOT NULL DEFAULT '0'"`
	IsTop      bool `json:"is_top" gorm:"type:TINYINT NOT NULL DEFAULT '0'"`
	ViewNum    int  `json:"view_num" gorm:"type:INT NOT NULL DEFAULT '0'"`
	CommentNum int  `json:"comment_num" gorm:"type:INT NOT NULL DEFAULT '0'"`
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

	BaseModifyTime types.Timestamp `json:"baseModifyTime"`
}

type ArticleCreate struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`

	IsElite bool `json:"is_elite"`
	IsTop   bool `json:"is_top"`

	Type   GenericType   `json:"type" binding:"required,oneof=1 2 3"`
	Source ArticleSource `json:"source" binding:"required,oneof=1 2 3 4"`
	Status ArticleStatus `json:"status" binding:"omitempty,oneof=0 1"`
}

var (
	PageSize          = 10
	QueryArticlesFunc = QueryArticles
	CreateArticleFunc = CreateArticle
	DetailArticleFunc = DetailArticle
	PatchArticleFunc  = PatchArticle
	DeleteArticleFunc = DeleteArticle

	timestampFunc = types.CurrentTimestamp
	idWorker      = sonyflake.NewSonyflake(sonyflake.Settings{})
	idFunc        = func() types.ID {
		return idgen.NextID(idWorker)
	}
	checkPermFunc         = checkPerm
	checkModifyBehindFunc = checkModifyBehind
)

func QueryArticles(q ArticleQuery, s *sessions.Session) ([]ArticleMetaExt, int64, error) {
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

	dbCount := persistence.ActiveGormDB.Model(&ArticleRecord{}).
		Where("is_invalid = 0 AND (status = 1 || uid = ?)", s.Identity.ID)

	if len(q.KeyWord) > 0 {
		db.Where("title LIKE ?", "%"+q.KeyWord+"%")
		dbCount.Where("title LIKE ?", "%"+q.KeyWord+"%")
	}

	var articleMetaExtList []ArticleMetaExt
	if err := db.Scan(&articleMetaExtList).Error; err != nil {
		return nil, 0, err
	}
	var count int64
	if err := dbCount.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if err := appendTags(articleMetaExtList, s); err != nil {
		return nil, 0, err
	}

	return articleMetaExtList, count, nil
}

func CreateArticle(q *ArticleCreate, s *sessions.Session) (types.ID, error) {
	if !s.IsAdmin() {
		return 0, fail.ErrForbidden
	}
	if q == nil {
		return 0, &fail.ErrBadParam{Cause: errors.New("bad param")}
	}

	ts := timestampFunc()
	r := ArticleRecord{
		ArticleMeta: ArticleMeta{
			ID:    idFunc(),
			Title: q.Title,

			Type:   q.Type,
			Status: q.Status,
			Source: q.Source,

			UID:        s.Identity.ID,
			CreateTime: ts,
			ModifyTime: ts,
		},
		Content: q.Content,
	}
	db := persistence.ActiveGormDB.Model(&ArticleRecord{})
	if err := db.Create(&r).Error; err != nil {
		return 0, err
	}

	return r.ID, nil
}

func PatchArticle(id types.ID, p *ArticlePatch, s *sessions.Session) (*types.Timestamp, error) {
	if p == nil || (*p == ArticlePatch{}) {
		return nil, nil
	}

	ts := timestampFunc()
	err := persistence.ActiveGormDB.Transaction(func(tx *gorm.DB) error {
		if err := checkPermFunc(tx, id, s); err != nil {
			return err
		}

		if err := checkModifyBehindFunc(tx, id, p.BaseModifyTime); err != nil {
			return err
		}

		tx = tx.Model(&ArticleRecord{}).Where("id = ?", id)

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
			changes["status"] = p.Status
		}
		if p.IsTop != nil {
			changes["is_top"] = p.IsTop
		}
		if p.IsElite != nil {
			changes["is_elite"] = p.IsElite
		}
		if len(changes) > 0 {
			changes["modify_time"] = ts
		}
		if err := tx.Save(&changes).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &ts, nil
}

func DetailArticle(id types.ID, s *sessions.Session) (*ArticleDetail, error) {
	var detail ArticleDetail
	db := persistence.ActiveGormDB.Model(&ArticleRecord{}).Select("*").
		Where("id = ? AND is_invalid = 0 AND (status = 1 || uid = ?)", id, s.Identity.ID)

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

func DeleteArticle(id types.ID, s *sessions.Session) error {
	err := persistence.ActiveGormDB.Transaction(func(tx *gorm.DB) error {
		if err := checkPermFunc(tx, id, s); err != nil {
			return err
		}

		db := tx.Where("id = ?", id)
		if !s.IsAdmin() {
			db = db.Where("uid = ?", s.Identity.ID)
		}
		if err := db.Delete(&ArticleRecord{}).Error; err != nil {
			return err
		}

		if err := ClearArticleTagAssignsFunc(tx, id, s); err != nil {
			return err
		}
		return nil
	})

	return err
}

// if the user is not admin and not the author, forbidden
// if record of the given id not found, sql.ErrNoRows?
func checkPerm(tx *gorm.DB, id types.ID, s *sessions.Session) error {
	if s.IsAdmin() {
		return nil
	}

	uidObj := misc.UidObject{}
	tx = tx.Model(&ArticleRecord{}).Select("uid").Where("id = ?", id)
	if err := tx.First(&uidObj).Error; err != nil {
		return err
	}
	if uidObj.UID != s.Identity.ID {
		return fail.ErrForbidden
	}
	return nil
}

func checkModifyBehind(tx *gorm.DB, id types.ID, baseTimestamp types.Timestamp) error {
	if baseTimestamp.IsZero() {
		return nil
	}
	a := ArticleRecord{}
	db := tx.Model(&ArticleRecord{}).Select("modify_time").Where("id = ?", id)
	if err := db.First(&a).Error; err != nil {
		return err
	}
	if a.ModifyTime.Time().After(baseTimestamp.Time()) {
		logrus.Warnf("expected last modified at %s, behind the actual last modified at %s\n", baseTimestamp, a.ModifyTime)
		return fail.ErrModifyBehind
	}
	return nil
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
