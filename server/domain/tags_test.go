package domain

import (
	"context"
	"database/sql"
	"owlet/server/infra/sessions"
	"owlet/server/testinfra"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/fundwit/go-commons/types"
	. "github.com/onsi/gomega"
)

func TestTagTableName(t *testing.T) {
	RegisterTestingT(t)

	t.Run("table name of Tag should be correct", func(t *testing.T) {
		r := Tag{}
		Expect(r.TableName()).To(Equal("tag"))
	})
}

func TestTagIdFunc(t *testing.T) {
	RegisterTestingT(t)

	t.Run("id func work as expected", func(t *testing.T) {
		Expect(tagIdFunc()).ToNot(BeZero())
	})
}

func TestQueryTags(t *testing.T) {
	RegisterTestingT(t)

	t.Run("should be able to query tags with min arguments", func(t *testing.T) {
		_, mock := testinfra.SetUpMockSql()

		tag := Tag{ID: 100, Name: "golang", Note: "go language", Image: "golang.png"}
		rows := sqlmock.NewRows([]string{"id", "tname", "note", "img"}).
			AddRow(tag.ID, tag.Name, tag.Note, tag.Image)

		const sqlExpr = "SELECT * FROM `tag`"
		mock.ExpectQuery(regexp.QuoteMeta(sqlExpr)).
			WillReturnRows(rows)

		mock.ExpectQuery(regexp.QuoteMeta(sqlExpr)).
			WillReturnError(sql.ErrConnDone)

		result, err := QueryTags(TagQuery{}, &sessions.Session{Context: context.TODO()})
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal([]Tag{tag}))

		result, err = QueryTags(TagQuery{}, &sessions.Session{Context: context.TODO()})
		Expect(err).To(Equal(sql.ErrConnDone))
		Expect(result).To(BeEmpty())

		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})

	t.Run("should be able to query tags with max arguments", func(t *testing.T) {
		_, mock := testinfra.SetUpMockSql()

		tag := Tag{ID: 100, Name: "golang", Note: "go language", Image: "golang.png"}
		rows := sqlmock.NewRows([]string{"id", "tname", "note", "img"}).
			AddRow(tag.ID, tag.Name, tag.Note, tag.Image)

		const sqlExpr = "SELECT * FROM `tag` WHERE id IN (?,?)"
		mock.ExpectQuery(regexp.QuoteMeta(sqlExpr)).
			WillReturnRows(rows)

		mock.ExpectQuery(regexp.QuoteMeta(sqlExpr)).
			WillReturnError(sql.ErrConnDone)

		result, err := QueryTags(TagQuery{IDs: []types.ID{100, 200}}, &sessions.Session{Context: context.TODO()})
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal([]Tag{tag}))

		result, err = QueryTags(TagQuery{IDs: []types.ID{100, 200}}, &sessions.Session{Context: context.TODO()})
		Expect(err).To(Equal(sql.ErrConnDone))
		Expect(result).To(BeEmpty())

		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})
}

func TestQueryTagsWithStat(t *testing.T) {
	RegisterTestingT(t)

	t.Run("should be able to query tags with stat success", func(t *testing.T) {
		mockTags := []Tag{{ID: 100, Name: "golang", Note: "go language", Image: "golang.png"}}
		mockTagsWithStat := []TagWithStat{{Tag: mockTags[0], Count: 10}}

		QueryTagsFunc = func(q TagQuery, s *sessions.Session) ([]Tag, error) {
			return mockTags, nil
		}
		ExtendTagsStatFunc = func(tags []Tag, s *sessions.Session) ([]TagWithStat, error) {
			return mockTagsWithStat, nil
		}

		result, err := QueryTagsWithStat(&sessions.Session{Context: context.TODO()})
		Expect(err).To(BeNil())
		Expect(result).To(Equal(mockTagsWithStat))
	})

	t.Run("should be able to query tags with stat on query tags error", func(t *testing.T) {
		QueryTagsFunc = func(q TagQuery, s *sessions.Session) ([]Tag, error) {
			return nil, sql.ErrConnDone
		}

		result, err := QueryTagsWithStat(&sessions.Session{Context: context.TODO()})
		Expect(err).To(Equal(sql.ErrConnDone))
		Expect(result).To(BeEmpty())
	})

	t.Run("should be able to query tags with stat on extend tag stats error", func(t *testing.T) {
		mockTags := []Tag{{ID: 100, Name: "golang", Note: "go language", Image: "golang.png"}}
		QueryTagsFunc = func(q TagQuery, s *sessions.Session) ([]Tag, error) {
			return mockTags, nil
		}
		ExtendTagsStatFunc = func(tags []Tag, s *sessions.Session) ([]TagWithStat, error) {
			return nil, sql.ErrConnDone
		}

		result, err := QueryTagsWithStat(&sessions.Session{Context: context.TODO()})
		Expect(err).To(Equal(sql.ErrConnDone))
		Expect(result).To(BeEmpty())
	})
}

func TestFindOrCreateTag(t *testing.T) {
	RegisterTestingT(t)

	t.Run("should be able to find tag", func(t *testing.T) {
		db, mock := testinfra.SetUpMockSql()
		q := TagCreate{
			Name: "test",
		}
		s := &sessions.Session{Context: context.TODO()}

		const sqlExpr = "SELECT * FROM `tag` WHERE tname LIKE ? ORDER BY `tag`.`id` LIMIT 1"
		mock.ExpectQuery(regexp.QuoteMeta(sqlExpr)).
			WithArgs("test").
			WillReturnRows(sqlmock.NewRows([]string{"id", "tname", "note", "img"}).
				AddRow(100, "test", "Test", "test.png"))

		tag, err := FindOrCreateTag(db, &q, s)
		Expect(err).ToNot(HaveOccurred())
		Expect(*tag).To(Equal(Tag{ID: 100, Name: "test", Note: "Test", Image: "test.png"}))

		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})

	t.Run("should raise error of find tag sql", func(t *testing.T) {
		db, mock := testinfra.SetUpMockSql()

		q := TagCreate{
			Name: "test",
		}
		s := &sessions.Session{Context: context.TODO()}

		const sqlExpr = "SELECT * FROM `tag` WHERE tname LIKE ? ORDER BY `tag`.`id` LIMIT 1"
		mock.ExpectQuery(regexp.QuoteMeta(sqlExpr)).
			WithArgs("test").
			WillReturnError(sql.ErrConnDone)

		idObj, err := FindOrCreateTag(db, &q, s)
		Expect(err).To(Equal(sql.ErrConnDone))
		Expect(idObj).To(BeNil())

		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})

	t.Run("should be create new tag when name not found", func(t *testing.T) {
		db, mock := testinfra.SetUpMockSql()
		q := TagCreate{
			Name: "test",
		}
		s := &sessions.Session{Context: context.TODO()}
		id := types.ID(200)
		tagIdFunc = func() types.ID {
			return id
		}

		const sqlExpr = "SELECT * FROM `tag` WHERE tname LIKE ? ORDER BY `tag`.`id` LIMIT 1"
		mock.ExpectQuery(regexp.QuoteMeta(sqlExpr)).
			WithArgs("test").
			WillReturnRows(sqlmock.NewRows([]string{"id", "tname", "note", "img"}))
		mock.ExpectBegin()
		const createSqlExpr = "INSERT INTO `tag` (`tname`,`note`,`img`,`id`) VALUES (?,?,?,?)"
		mock.ExpectExec(regexp.QuoteMeta(createSqlExpr)).
			WithArgs("test", "", "", id).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tag, err := FindOrCreateTag(db, &q, s)
		Expect(err).ToNot(HaveOccurred())
		Expect(*tag).To(Equal(Tag{ID: id, Name: "test"}))

		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})

	t.Run("should raise error of tag create sql", func(t *testing.T) {
		db, mock := testinfra.SetUpMockSql()

		q := TagCreate{
			Name: "test",
		}
		s := &sessions.Session{Context: context.TODO()}
		id := types.ID(200)
		tagIdFunc = func() types.ID {
			return id
		}

		const sqlExpr = "SELECT * FROM `tag` WHERE tname LIKE ? ORDER BY `tag`.`id` LIMIT 1"
		mock.ExpectQuery(regexp.QuoteMeta(sqlExpr)).
			WithArgs("test").
			WillReturnRows(sqlmock.NewRows([]string{"id", "tname", "note", "img"}))
		mock.ExpectBegin()
		const createSqlExpr = "INSERT INTO `tag` (`tname`,`note`,`img`,`id`) VALUES (?,?,?,?)"
		mock.ExpectExec(regexp.QuoteMeta(createSqlExpr)).
			WithArgs("test", "", "", id).
			WillReturnError(sql.ErrConnDone)
		mock.ExpectRollback()

		idObj, err := FindOrCreateTag(db, &q, s)
		Expect(err).To(Equal(sql.ErrConnDone))
		Expect(idObj).To(BeNil())

		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})
}

func TestExtendTagsStat(t *testing.T) {
	RegisterTestingT(t)

	t.Run("should be able to extend tags stat with nil or empty input", func(t *testing.T) {
		result, err := ExtendTagsStat(nil, &sessions.Session{Context: context.TODO()})
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal([]TagWithStat{}))

		result, err = ExtendTagsStat([]Tag{}, &sessions.Session{Context: context.TODO()})
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal([]TagWithStat{}))
	})

	t.Run("should be able to extend tags stat", func(t *testing.T) {
		_, mock := testinfra.SetUpMockSql()

		tag := Tag{ID: 100, Name: "golang", Note: "go language", Image: "golang.png"}
		tagNoAssign := Tag{ID: 200, Name: "javascript", Note: "javascript language", Image: "javascript.png"}

		rows := sqlmock.NewRows([]string{"id", "count"}).
			AddRow(tag.ID, "30")

		const sqlExpr = "SELECT res_id AS id, count(*) AS count FROM `tag_assign` WHERE res_id IN (?,?) AND res_type = ? GROUP BY `res_id`"
		mock.ExpectQuery(regexp.QuoteMeta(sqlExpr)).
			//WithArgs(sqlmock.AnyArg, 0).
			WillReturnRows(rows)

		mock.ExpectQuery(regexp.QuoteMeta(sqlExpr)).
			//WithArgs(sqlmock.AnyArg, 0).
			WillReturnError(sql.ErrConnDone)

		result, err := ExtendTagsStat([]Tag{tag, tagNoAssign}, &sessions.Session{Context: context.TODO()})
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal([]TagWithStat{{Tag: tag, Count: 30}, {Tag: tagNoAssign, Count: 0}}))

		result, err = ExtendTagsStat([]Tag{tag, tagNoAssign}, &sessions.Session{Context: context.TODO()})
		Expect(err).To(Equal(sql.ErrConnDone))
		Expect(result).To(BeEmpty())

		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})
}
