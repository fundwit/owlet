package domain

import (
	"context"
	"database/sql"
	"owlet/server/infra/sessions"
	"owlet/server/testinfra"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/fundwit/go-commons/types"
	. "github.com/onsi/gomega"
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
		WithArgs(0, "%go%").
		WillReturnRows(rows)

	QueryTagAssignmentsFunc = func(resIds []types.ID, s *sessions.Session) ([]TagAssignment, error) {
		return nil, nil
	}

	result, err := QueryArticles(ArticleQuery{KeyWord: "go", Page: 3}, &sessions.Session{Context: context.TODO()})
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

	const sqlExpr = "SELECT * FROM `article` WHERE id = ?"
	mock.ExpectQuery(regexp.QuoteMeta(sqlExpr)).
		WithArgs(100).
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

	result, err := DetailArticle(100, &sessions.Session{Context: context.TODO()})
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

	const sqlExpr = "SELECT * FROM `article` WHERE id = ?"
	mock.ExpectQuery(regexp.QuoteMeta(sqlExpr)).
		WithArgs(100).
		WillReturnRows(rows)

	QueryTagAssignmentsFunc = func(resIds []types.ID, s *sessions.Session) ([]TagAssignment, error) {
		return nil, sql.ErrConnDone
	}

	result, err := DetailArticle(100, &sessions.Session{Context: context.TODO()})
	Expect(err).To(Equal(sql.ErrConnDone))
	Expect(result).To(BeNil())

	Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
}

func TestDetailArticle_ErrorOnQueryArticle(t *testing.T) {
	RegisterTestingT(t)

	_, mock := testinfra.SetUpMockSql()
	const sqlExpr = "SELECT * FROM `article` WHERE id = ?"
	mock.ExpectQuery(regexp.QuoteMeta(sqlExpr)).
		WithArgs(100).
		WillReturnError(sql.ErrConnDone)

	QueryTagAssignmentsFunc = func(resIds []types.ID, s *sessions.Session) ([]TagAssignment, error) {
		return nil, sql.ErrConnDone
	}

	result, err := DetailArticle(100, &sessions.Session{Context: context.TODO()})
	Expect(err).To(Equal(sql.ErrConnDone))
	Expect(result).To(BeNil())

	Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
}
