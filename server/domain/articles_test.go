package domain

import (
	"context"
	"database/sql"
	"owlet/server/infra/fail"
	"owlet/server/infra/sessions"
	"owlet/server/testinfra"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/fundwit/go-commons/types"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

func TestArticleTableName(t *testing.T) {
	RegisterTestingT(t)

	t.Run("table name of Tag should be correct", func(t *testing.T) {
		r := ArticleRecord{}
		Expect(r.TableName()).To(Equal("article"))
	})
}

func TestQueryArticleMetas_MinArgsWithTagsExtend(t *testing.T) {
	RegisterTestingT(t)

	_, mock := testinfra.SetUpMockSql()
	article := ArticleRecord{
		ArticleMeta: ArticleMeta{
			ID: 100, Type: 1, Title: "title", UID: 1000,
		},
		Content: "content",
	}
	article2 := ArticleRecord{
		ArticleMeta: ArticleMeta{
			ID: 200, Type: 1, Title: "title 2", UID: 1000,
		},
		Content: "content 2",
	}
	rows := sqlmock.NewRows([]string{"id", "type", "title", "uid"}).
		AddRow(article.ID, article.Type, article.Title, article.UID).
		AddRow(article2.ID, article2.Type, article2.Title, article2.UID)

	const sqlExpr = "SELECT id, type, title, uid, create_time, modify_time, status, is_invalid, " +
		"abstracts, source, is_elite, is_top, view_num, comment_num " +
		"FROM `article` WHERE is_invalid = 0 AND (status = 1 || uid = ?) " +
		"ORDER BY is_top DESC, create_time DESC LIMIT 10"
	mock.ExpectQuery(regexp.QuoteMeta(sqlExpr)).
		WithArgs(0).
		WillReturnRows(rows)

	QueryTagAssignmentsFunc = func(resIds []types.ID, s *sessions.Session) ([]TagAssignment, error) {
		return []TagAssignment{
			{ID: 2000, TagID: 20, ResID: article.ID, ResType: 0, TagOrder: 1},
			{ID: 3000, TagID: 30, ResID: article.ID, ResType: 0, TagOrder: 2},
			{ID: 2000, TagID: 30, ResID: article2.ID, ResType: 0, TagOrder: 1},
		}, nil
	}

	tags := []Tag{
		{ID: 20, Name: "tag20", Image: "tag20.png", Note: "tag 20"},
		{ID: 30, Name: "tag30", Image: "tag30.png", Note: "tag 30"},
	}
	QueryTagsFunc = func(q TagQuery, s *sessions.Session) ([]Tag, error) {
		return tags, nil
	}

	result, err := QueryArticles(ArticleQuery{Page: 0}, &sessions.Session{Context: context.TODO()})

	Expect(err).ToNot(HaveOccurred())
	Expect(result).To(Equal([]ArticleMetaExt{
		{ArticleMeta: article.ArticleMeta, Tags: tags},
		{ArticleMeta: article2.ArticleMeta, Tags: []Tag{tags[1]}},
	}))

	Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
}

func TestQueryArticleMetas_MaxArgs(t *testing.T) {
	RegisterTestingT(t)

	_, mock := testinfra.SetUpMockSql()

	article := ArticleRecord{
		ArticleMeta: ArticleMeta{
			ID: 100, Type: 1, Title: "title", UID: 1000,
		},
		Content: "content",
	}
	rows := sqlmock.NewRows([]string{"id", "type", "title", "uid"}).
		AddRow(article.ID, article.Type, article.Title, article.UID)

	const sqlExpr = "SELECT id, type, title, uid, create_time, modify_time, status, is_invalid, " +
		"abstracts, source, is_elite, is_top, view_num, comment_num " +
		"FROM `article` WHERE (is_invalid = 0 AND (status = 1 || uid = ?)) AND title LIKE ? " +
		"ORDER BY is_top DESC, create_time DESC LIMIT 10 OFFSET 20"
	mock.ExpectQuery(regexp.QuoteMeta(sqlExpr)).
		WithArgs(10, "%go%").
		WillReturnRows(rows)

	QueryTagAssignmentsFunc = func(resIds []types.ID, s *sessions.Session) ([]TagAssignment, error) {
		return nil, nil
	}

	result, err := QueryArticles(ArticleQuery{KeyWord: "go", Page: 3}, &sessions.Session{
		Context:  context.TODO(),
		Identity: sessions.Identity{ID: 10},
	})
	Expect(err).ToNot(HaveOccurred())
	Expect(result).To(Equal([]ArticleMetaExt{{ArticleMeta: article.ArticleMeta, Tags: nil}}))

	Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
}

func TestQueryArticleMetas_MinArgsAndNoResult(t *testing.T) {
	RegisterTestingT(t)

	_, mock := testinfra.SetUpMockSql()

	rows := sqlmock.NewRows([]string{"id", "type", "title", "uid"})

	const sqlExpr = "SELECT id, type, title, uid, create_time, modify_time, status, is_invalid, " +
		"abstracts, source, is_elite, is_top, view_num, comment_num " +
		"FROM `article` WHERE is_invalid = 0 AND (status = 1 || uid = ?) " +
		"ORDER BY is_top DESC, create_time DESC LIMIT 10"
	mock.ExpectQuery(regexp.QuoteMeta(sqlExpr)).
		WithArgs(0).
		WillReturnRows(rows)

	result, err := QueryArticles(ArticleQuery{}, &sessions.Session{Context: context.TODO()})
	Expect(err).ToNot(HaveOccurred())
	Expect(result).To(BeNil())

	Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
}

func TestQueryArticleMetas_ErrorOnQueryTags(t *testing.T) {
	RegisterTestingT(t)

	_, mock := testinfra.SetUpMockSql()

	article := ArticleRecord{
		ArticleMeta: ArticleMeta{ID: 100, Type: 1, Title: "title", UID: 1000},
		Content:     "content",
	}
	rows := sqlmock.NewRows([]string{"id", "type", "title", "uid"}).
		AddRow(article.ID, article.Type, article.Title, article.UID)

	const sqlExpr = "SELECT id, type, title, uid, create_time, modify_time, status, is_invalid, " +
		"abstracts, source, is_elite, is_top, view_num, comment_num " +
		"FROM `article` WHERE is_invalid = 0 AND (status = 1 || uid = ?) " +
		"ORDER BY is_top DESC, create_time DESC LIMIT 10"

	mock.ExpectQuery(regexp.QuoteMeta(sqlExpr)).
		WithArgs(0).
		WillReturnRows(rows)

	QueryTagAssignmentsFunc = func(resIds []types.ID, s *sessions.Session) ([]TagAssignment, error) {
		return []TagAssignment{
			{ID: 2000, TagID: 20, ResID: article.ID, ResType: 0, TagOrder: 1},
		}, nil
	}

	QueryTagsFunc = func(q TagQuery, s *sessions.Session) ([]Tag, error) {
		return nil, sql.ErrConnDone
	}

	result, err := QueryArticles(ArticleQuery{}, &sessions.Session{Context: context.TODO()})
	Expect(err).To(Equal(sql.ErrConnDone))
	Expect(result).To(BeEmpty())

	Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
}

func TestQueryArticleMetas_ErrorOnQueryTagAssignments(t *testing.T) {
	RegisterTestingT(t)

	_, mock := testinfra.SetUpMockSql()

	article := ArticleRecord{
		ArticleMeta: ArticleMeta{ID: 100, Type: 1, Title: "title", UID: 1000},
		Content:     "content",
	}
	rows := sqlmock.NewRows([]string{"id", "type", "title", "uid"}).
		AddRow(article.ID, article.Type, article.Title, article.UID)

	const sqlExpr = "SELECT id, type, title, uid, create_time, modify_time, status, is_invalid, " +
		"abstracts, source, is_elite, is_top, view_num, comment_num " +
		"FROM `article` WHERE is_invalid = 0 AND (status = 1 || uid = ?) " +
		"ORDER BY is_top DESC, create_time DESC LIMIT 10"

	mock.ExpectQuery(regexp.QuoteMeta(sqlExpr)).
		WithArgs(0).
		WillReturnRows(rows)

	QueryTagAssignmentsFunc = func(resIds []types.ID, s *sessions.Session) ([]TagAssignment, error) {
		return nil, sql.ErrConnDone
	}

	result, err := QueryArticles(ArticleQuery{}, &sessions.Session{Context: context.TODO()})
	Expect(err).To(Equal(sql.ErrConnDone))
	Expect(result).To(BeEmpty())

	Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
}

func TestQueryArticleMetas_ErrorOnQueryArticles(t *testing.T) {
	RegisterTestingT(t)

	_, mock := testinfra.SetUpMockSql()

	const sqlExpr = "SELECT id, type, title, uid, create_time, modify_time, status, is_invalid, " +
		"abstracts, source, is_elite, is_top, view_num, comment_num " +
		"FROM `article` WHERE is_invalid = 0 AND (status = 1 || uid = ?) " +
		"ORDER BY is_top DESC, create_time DESC LIMIT 10"

	mock.ExpectQuery(regexp.QuoteMeta(sqlExpr)).
		WithArgs(0).
		WillReturnError(sql.ErrConnDone)

	result, err := QueryArticles(ArticleQuery{}, &sessions.Session{Context: context.TODO()})
	Expect(err).To(Equal(sql.ErrConnDone))
	Expect(result).To(BeEmpty())

	Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
}

func TestDetailArticle_FoundArticleWithTags(t *testing.T) {
	RegisterTestingT(t)

	mockTime1 := types.TimestampOfDate(2000, 1, 2, 3, 4, 5, 0, time.UTC)
	mockTime2 := types.TimestampOfDate(2001, 2, 3, 12, 5, 6, 0, time.FixedZone("x", 28800))
	_, mock := testinfra.SetUpMockSql()
	a := ArticleRecord{
		ArticleMeta: ArticleMeta{
			ID: 100, Type: 1, Title: "title", UID: 1000, CreateTime: mockTime1, ModifyTime: mockTime2, Status: 1,
			IsInvalid: true, Abstracts: "abstract 100", Source: 2, IsElite: true, IsTop: true,
			ViewNum: 123, CommentNum: 45,
		},
		Content: "content 100",
	}
	rows := sqlmock.NewRows([]string{"id", "type", "title", "uid", "create_time", "modify_time", "status", "is_invalid",
		"abstracts", "source", "is_elite", "is_top", "view_num", "comment_num", "content"}).
		AddRow(a.ID, a.Type, a.Title, a.UID, a.CreateTime, a.ModifyTime, a.Status,
			a.IsInvalid, a.Abstracts, a.Source, a.IsElite, a.IsTop, a.ViewNum, a.CommentNum, a.Content)

	const sqlExpr = "SELECT * FROM `article` WHERE id = ? AND is_invalid = 0 AND (status = 1 || uid = ?)"
	mock.ExpectQuery(regexp.QuoteMeta(sqlExpr)).
		WithArgs(100, 10).
		WillReturnRows(rows)

	QueryTagAssignmentsFunc = func(resIds []types.ID, s *sessions.Session) ([]TagAssignment, error) {
		return []TagAssignment{
			{ID: 2000, TagID: 20, ResID: a.ID, ResType: 0, TagOrder: 1},
			{ID: 3000, TagID: 30, ResID: a.ID, ResType: 0, TagOrder: 2},
		}, nil
	}

	tags := []Tag{
		{ID: 20, Name: "tag20", Image: "tag20.png", Note: "tag 20"},
		{ID: 30, Name: "tag30", Image: "tag30.png", Note: "tag 30"},
	}
	QueryTagsFunc = func(q TagQuery, s *sessions.Session) ([]Tag, error) {
		return tags, nil
	}

	result, err := DetailArticle(100, &sessions.Session{
		Context:  context.TODO(),
		Identity: sessions.Identity{ID: 10},
	})
	Expect(err).ToNot(HaveOccurred())

	want := ArticleDetail{
		ArticleRecord: ArticleRecord{ArticleMeta: a.ArticleMeta, Content: a.Content},
		Tags:          tags,
	}
	Expect(want.CreateTime.Time().Hour()).To(Equal(3))
	want.CreateTime = types.Timestamp(want.CreateTime.Time().In(result.CreateTime.Time().Location()))
	want.ModifyTime = types.Timestamp(want.ModifyTime.Time().In(result.ModifyTime.Time().Location()))
	Expect(want.CreateTime.Time().Hour()).To(Equal(11))

	Expect(result.ArticleMeta.CreateTime).To(Equal(want.ArticleMeta.CreateTime))
	Expect(result.Tags).To(Equal(want.Tags))
	Expect(*result).To(Equal(want))

	Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
}

func TestDetailArticle_ErrorOnAppendTags(t *testing.T) {
	RegisterTestingT(t)

	_, mock := testinfra.SetUpMockSql()
	a := ArticleRecord{
		ArticleMeta: ArticleMeta{ID: 100, Type: 1, Title: "title", UID: 1000},
		Content:     "content 100",
	}
	rows := sqlmock.NewRows([]string{"id", "type", "title", "uid"}).
		AddRow(a.ID, a.Type, a.Title, a.UID)

	const sqlExpr = "SELECT * FROM `article` WHERE id = ? AND is_invalid = 0 AND (status = 1 || uid = ?)"
	mock.ExpectQuery(regexp.QuoteMeta(sqlExpr)).
		WithArgs(100, 10).
		WillReturnRows(rows)

	QueryTagAssignmentsFunc = func(resIds []types.ID, s *sessions.Session) ([]TagAssignment, error) {
		return nil, sql.ErrConnDone
	}

	result, err := DetailArticle(100, &sessions.Session{
		Context:  context.TODO(),
		Identity: sessions.Identity{ID: 10},
	})
	Expect(err).To(Equal(sql.ErrConnDone))
	Expect(result).To(BeNil())

	Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
}

func TestDetailArticle_ErrorOnQueryArticle(t *testing.T) {
	RegisterTestingT(t)

	_, mock := testinfra.SetUpMockSql()
	const sqlExpr = "SELECT * FROM `article` WHERE id = ? AND is_invalid = 0 AND (status = 1 || uid = ?)"
	mock.ExpectQuery(regexp.QuoteMeta(sqlExpr)).
		WithArgs(100, 10).
		WillReturnError(sql.ErrConnDone)

	QueryTagAssignmentsFunc = func(resIds []types.ID, s *sessions.Session) ([]TagAssignment, error) {
		return nil, sql.ErrConnDone
	}

	result, err := DetailArticle(100, &sessions.Session{Context: context.TODO(), Identity: sessions.Identity{ID: 10}})
	Expect(err).To(Equal(sql.ErrConnDone))
	Expect(result).To(BeNil())

	Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
}

func TestPatchArticle_NullParams(t *testing.T) {
	RegisterTestingT(t)

	ts := types.CurrentTimestamp()
	timestampFunc = func() types.Timestamp {
		return ts
	}
	s := &sessions.Session{Context: context.TODO(), Identity: sessions.Identity{ID: 1}}
	user := &sessions.Session{Context: context.TODO(), Identity: sessions.Identity{ID: 10}}

	t.Run("patch article with null params should return directly", func(t *testing.T) {
		_, mock := testinfra.SetUpMockSql()
		err := PatchArticle(100, nil, s)
		Expect(err).To(BeNil())

		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})

	t.Run("patch article with empty params should return directly", func(t *testing.T) {
		_, mock := testinfra.SetUpMockSql()
		err := PatchArticle(100, &ArticlePatch{}, s)
		Expect(err).To(BeNil())

		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})

	t.Run("only admin can patch article", func(t *testing.T) {
		permCheckFunc = func(tx *gorm.DB, id types.ID, s *sessions.Session) error {
			Expect(id).To(Equal(types.ID(100)))
			Expect(*s).To(Equal(*user))
			return fail.ErrForbidden
		}

		_, mock := testinfra.SetUpMockSql()
		mock.ExpectBegin()
		mock.ExpectRollback()

		err := PatchArticle(100, &ArticlePatch{Title: "test title"}, user)
		Expect(err).To(Equal(fail.ErrForbidden))

		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})

	t.Run("patch article with error on update", func(t *testing.T) {
		permCheckFunc = func(tx *gorm.DB, id types.ID, s *sessions.Session) error {
			return nil
		}

		_, mock := testinfra.SetUpMockSql()
		const sqlExpr = "UPDATE `article` SET `content`=?,`is_elite`=?,`is_top`=?,`modify_time`=?," +
			"`source`=?,`status`=?,`title`=?,`type`=? WHERE id = ?"
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(sqlExpr)).
			WithArgs("test content", true, true, ts, 3, 1, "test title", 2, 100).
			WillReturnError(sql.ErrConnDone)
		mock.ExpectRollback()

		tp := GenericType(2)
		src := ArticleSource(3)
		status := ArticleStatus(1)
		elite := true
		top := true

		param := ArticlePatch{
			Title:   "test title",
			Content: "test content",
			Type:    &tp,
			Source:  &src,
			Status:  &status,
			IsElite: &elite,
			IsTop:   &top,
		}
		err := PatchArticle(100, &param, s)
		Expect(err).To(Equal(sql.ErrConnDone))

		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})

	t.Run("patch article success with max params", func(t *testing.T) {
		permCheckFunc = func(tx *gorm.DB, id types.ID, s *sessions.Session) error {
			return nil
		}

		_, mock := testinfra.SetUpMockSql()
		const sqlExpr = "UPDATE `article` SET `content`=?,`is_elite`=?,`is_top`=?,`modify_time`=?," +
			"`source`=?,`status`=?,`title`=?,`type`=? WHERE id = ?"
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(sqlExpr)).
			WithArgs("test content", true, true, ts, 3, 1, "test title", 2, 100).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tp := GenericType(2)
		src := ArticleSource(3)
		status := ArticleStatus(1)
		elite := true
		top := true

		param := ArticlePatch{
			Title:   "test title",
			Content: "test content",
			Type:    &tp,
			Source:  &src,
			Status:  &status,
			IsElite: &elite,
			IsTop:   &top,
		}
		err := PatchArticle(100, &param, s)
		Expect(err).To(BeNil())

		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})
}

func TestIdFunc(t *testing.T) {
	RegisterTestingT(t)

	t.Run("id func work as expected", func(t *testing.T) {
		Expect(idFunc()).ToNot(BeZero())
	})
}

func TestCreateArticle(t *testing.T) {
	RegisterTestingT(t)

	s := &sessions.Session{Context: context.TODO(), Identity: sessions.Identity{ID: 1}}
	ts := types.CurrentTimestamp()
	timestampFunc = func() types.Timestamp {
		return ts
	}

	t.Run("create article should be forbidden without admin", func(t *testing.T) {
		id, err := CreateArticle(nil, &sessions.Session{Context: context.TODO(), Identity: sessions.Identity{ID: 10}})
		Expect(id).To(BeZero())
		Expect(err).To(Equal(fail.ErrForbidden))
	})

	t.Run("create article should be able to expose error of nil param", func(t *testing.T) {
		id, err := CreateArticle(nil, s)
		Expect(id).To(BeZero())
		Expect(err.Error()).To(Equal("bad param"))
	})

	t.Run("create article should be able to expose error of database", func(t *testing.T) {
		idFunc = func() types.ID {
			return 200
		}

		_, mock := testinfra.SetUpMockSql()
		const sqlExpr = "INSERT INTO `article` (`title`,`abstracts`,`type`,`status`,`source`,`uid`," +
			"`create_time`,`modify_time`,`is_invalid`,`is_elite`,`is_top`,`view_num`,`comment_num`,`content`,`id`)" +
			" VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(sqlExpr)).
			WithArgs("test title", "", GenericType(2), ArticleStatus(1), ArticleSource(3),
				s.Identity.ID, ts, ts, false, false, false, 0, 0, "test content", 200).
			WillReturnError(sql.ErrConnDone)
		mock.ExpectRollback()

		param := ArticleCreate{
			Title:   "test title",
			Content: "test content",
			Type:    GenericType(2),
			Source:  ArticleSource(3),
			Status:  ArticleStatus(1),
			IsElite: true,
			IsTop:   true,
		}
		id, err := CreateArticle(&param, s)
		Expect(id).To(BeZero())
		Expect(err).To(Equal(sql.ErrConnDone))

		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})

	t.Run("create article should be able to work as expect", func(t *testing.T) {
		idFunc = func() types.ID {
			return 200
		}

		_, mock := testinfra.SetUpMockSql()
		const sqlExpr = "INSERT INTO `article` (`title`,`abstracts`,`type`,`status`,`source`,`uid`," +
			"`create_time`,`modify_time`,`is_invalid`,`is_elite`,`is_top`,`view_num`,`comment_num`,`content`,`id`)" +
			" VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(sqlExpr)).
			WithArgs("test title", "", GenericType(2), ArticleStatus(1), ArticleSource(3),
				s.Identity.ID, ts, ts, false, false, false, 0, 0, "test content", 200).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		param := ArticleCreate{
			Title:   "test title",
			Content: "test content",
			Type:    GenericType(2),
			Source:  ArticleSource(3),
			Status:  ArticleStatus(1),
			IsElite: true,
			IsTop:   true,
		}
		id, err := CreateArticle(&param, s)
		Expect(id).To(Equal(types.ID(200)))
		Expect(err).To(BeNil())

		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})
}

func TestDeleteArticle(t *testing.T) {
	RegisterTestingT(t)

	admin := &sessions.Session{Context: context.TODO(), Identity: sessions.Identity{ID: 1}}
	ts := types.CurrentTimestamp()
	timestampFunc = func() types.Timestamp {
		return ts
	}

	t.Run("error raised when perm check failed", func(t *testing.T) {
		_, mock := testinfra.SetUpMockSql()
		permCheckFunc = func(tx *gorm.DB, id types.ID, s *sessions.Session) error {
			Expect(id).To(Equal(types.ID(10)))
			Expect(*s).To(Equal(*admin))
			return sql.ErrConnDone
		}
		mock.ExpectBegin()
		mock.ExpectRollback()

		Expect(DeleteArticle(10, admin)).To(Equal(sql.ErrConnDone))

		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})

	t.Run("error raised when delete sql failed", func(t *testing.T) {
		_, mock := testinfra.SetUpMockSql()
		permCheckFunc = func(tx *gorm.DB, id types.ID, s *sessions.Session) error {
			return nil
		}
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `article` WHERE id = ?")).
			WithArgs(10).
			WillReturnError(sql.ErrConnDone)
			//WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectRollback()

		Expect(DeleteArticle(10, admin)).To(Equal(sql.ErrConnDone))

		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})

	t.Run("error raised when failed to delete tag assign", func(t *testing.T) {
		_, mock := testinfra.SetUpMockSql()
		permCheckFunc = func(tx *gorm.DB, id types.ID, s *sessions.Session) error {
			return nil
		}
		ClearArticleTagAssignsFunc = func(tx *gorm.DB, articleId types.ID, s *sessions.Session) error {
			return sql.ErrConnDone
		}
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `article` WHERE id = ?")).
			WithArgs(10).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectRollback()

		Expect(DeleteArticle(10, admin)).To(Equal(sql.ErrConnDone))

		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})

	t.Run("success if user is admin", func(t *testing.T) {
		_, mock := testinfra.SetUpMockSql()
		permCheckFunc = func(tx *gorm.DB, id types.ID, s *sessions.Session) error {
			return nil
		}
		ClearArticleTagAssignsFunc = func(tx *gorm.DB, articleId types.ID, s *sessions.Session) error {
			return nil
		}
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `article` WHERE id = ?")).
			WithArgs(10).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		Expect(DeleteArticle(10, admin)).To(BeNil())

		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})

	t.Run("success if user is not admin", func(t *testing.T) {
		user := &sessions.Session{Context: context.TODO(), Identity: sessions.Identity{ID: 100}}

		_, mock := testinfra.SetUpMockSql()
		permCheckFunc = func(tx *gorm.DB, id types.ID, s *sessions.Session) error {
			return nil
		}
		ClearArticleTagAssignsFunc = func(tx *gorm.DB, articleId types.ID, s *sessions.Session) error {
			return nil
		}
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `article` WHERE id = ? AND uid = ?")).
			WithArgs(10, 100).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		Expect(DeleteArticle(10, user)).To(BeNil())

		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})
}

func TestPermCheck(t *testing.T) {
	RegisterTestingT(t)

	admin := &sessions.Session{Context: context.TODO(), Identity: sessions.Identity{ID: 1}}
	user := &sessions.Session{Context: context.TODO(), Identity: sessions.Identity{ID: 10}}
	ts := types.CurrentTimestamp()
	timestampFunc = func() types.Timestamp {
		return ts
	}

	t.Run("success if user is admin", func(t *testing.T) {
		Expect(permCheck(nil, 10, admin)).To(BeNil())
	})

	t.Run("get error article not found", func(t *testing.T) {
		db, mock := testinfra.SetUpMockSql()
		rows := sqlmock.NewRows([]string{"uid"})

		const sqlExpr = "SELECT `uid` FROM `article` WHERE id = ? ORDER BY `article`.`id` LIMIT 1"
		mock.ExpectQuery(regexp.QuoteMeta(sqlExpr)).
			WithArgs(10).
			WillReturnRows(rows)

		err := permCheck(db, 10, user)
		Expect(err).To(Equal(gorm.ErrRecordNotFound))

		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})

	t.Run("forbidden if user not the author", func(t *testing.T) {
		db, mock := testinfra.SetUpMockSql()
		rows := sqlmock.NewRows([]string{"uid"}).AddRow("20")

		const sqlExpr = "SELECT `uid` FROM `article` WHERE id = ? ORDER BY `article`.`id` LIMIT 1"
		mock.ExpectQuery(regexp.QuoteMeta(sqlExpr)).
			WithArgs(10).
			WillReturnRows(rows)

		err := permCheck(db, 10, user)
		Expect(err).To(Equal(fail.ErrForbidden))

		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})

	t.Run("success if user is the author", func(t *testing.T) {
		db, mock := testinfra.SetUpMockSql()
		rows := sqlmock.NewRows([]string{"uid"}).AddRow("10")

		const sqlExpr = "SELECT `uid` FROM `article` WHERE id = ? ORDER BY `article`.`id` LIMIT 1"
		mock.ExpectQuery(regexp.QuoteMeta(sqlExpr)).
			WithArgs(10).
			WillReturnRows(rows)

		err := permCheck(db, 10, user)
		Expect(err).To(BeNil())

		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})

}
