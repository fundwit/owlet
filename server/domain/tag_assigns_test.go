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
	"gorm.io/gorm"
)

func TestTagAssignmentTableName(t *testing.T) {
	RegisterTestingT(t)

	t.Run("table name of TagAssignment should be correct", func(t *testing.T) {
		r := TagAssignment{}
		Expect(r.TableName()).To(Equal("tag_assign"))
	})
}

func TestTagAssignIdFunc(t *testing.T) {
	RegisterTestingT(t)

	t.Run("id func work as expected", func(t *testing.T) {
		Expect(tagAssignIdFunc()).ToNot(BeZero())
	})
}

func TestQueryTagAssignments(t *testing.T) {
	RegisterTestingT(t)

	t.Run("should be able to query tag assignments", func(t *testing.T) {
		_, mock := testinfra.SetUpMockSql()

		tagAssign := TagAssignment{ID: 10000, ResID: 100, TagID: 10, ResType: 0, TagOrder: 1}
		rows := sqlmock.NewRows([]string{"id", "res_id", "tag", "res_type", "tag_order"}).
			AddRow(tagAssign.ID, tagAssign.ResID, tagAssign.TagID, tagAssign.ResType, tagAssign.TagOrder)

		const sqlExpr = "SELECT * FROM `tag_assign` WHERE res_id IN (?,?,?) AND res_type = 0"
		mock.ExpectQuery(regexp.QuoteMeta(sqlExpr)).
			WillReturnRows(rows)

		result, err := QueryTagAssignments([]types.ID{100, 200, 300}, &sessions.Session{Context: context.TODO()})
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal([]TagAssignment{tagAssign}))

		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})

	t.Run("should be able to query tag assignments on database error", func(t *testing.T) {
		_, mock := testinfra.SetUpMockSql()

		const sqlExpr = "SELECT * FROM `tag_assign` WHERE res_id IN (?,?) AND res_type = 0"
		mock.ExpectQuery(regexp.QuoteMeta(sqlExpr)).
			WillReturnError(sql.ErrConnDone)

		result, err := QueryTagAssignments([]types.ID{100, 200}, &sessions.Session{Context: context.TODO()})
		Expect(err).To(Equal(sql.ErrConnDone))
		Expect(result).To(BeEmpty())

		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})

	t.Run("should be able to query tag assignments on empty args", func(t *testing.T) {
		_, mock := testinfra.SetUpMockSql()

		result, err := QueryTagAssignments([]types.ID{}, &sessions.Session{Context: context.TODO()})
		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).To(BeEmpty())

		result, err = QueryTagAssignments(nil, &sessions.Session{Context: context.TODO()})
		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).To(BeEmpty())

		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})
}

func TestClearArticleTagAssigns(t *testing.T) {
	RegisterTestingT(t)

	t.Run("should be able to clear tag assignments of article", func(t *testing.T) {
		db, mock := testinfra.SetUpMockSql()

		const sqlExpr = "DELETE FROM `tag_assign` WHERE res_id = ? AND res_type = 0"
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(sqlExpr)).
			WithArgs(100).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := ClearArticleTagAssigns(db, 100, &sessions.Session{Context: context.TODO()})
		Expect(err).ToNot(HaveOccurred())

		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})

	t.Run("should raise error of delete sql", func(t *testing.T) {
		db, mock := testinfra.SetUpMockSql()

		const sqlExpr = "DELETE FROM `tag_assign` WHERE res_id = ? AND res_type = 0"
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(sqlExpr)).
			WithArgs(100).
			WillReturnError(sql.ErrConnDone)
		mock.ExpectRollback()

		err := ClearArticleTagAssigns(db, 100, &sessions.Session{Context: context.TODO()})
		Expect(err).To(Equal(err))

		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})
}

func TestCreateTagAssign(t *testing.T) {
	RegisterTestingT(t)

	t.Run("should raise error of perm check", func(t *testing.T) {
		_, mock := testinfra.SetUpMockSql()

		q := TagAssignCreate{
			ResID:   100,
			TagName: "test",
		}
		s := &sessions.Session{Context: context.TODO()}
		tagAssignIdFunc = func() types.ID {
			return 123
		}
		checkPermFunc = func(tx *gorm.DB, id types.ID, s0 *sessions.Session) error {
			Expect(id).To(Equal(types.ID(100)))
			Expect(*s).To(Equal(*s0))
			return sql.ErrConnDone
		}

		mock.ExpectBegin()
		mock.ExpectRollback()

		idObj, err := CreateTagAssign(&q, s)
		Expect(err).To(Equal(sql.ErrConnDone))
		Expect(idObj).To(BeNil())

		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})

	t.Run("should raise error of tag find or create error", func(t *testing.T) {
		_, mock := testinfra.SetUpMockSql()

		q := TagAssignCreate{
			ResID:   100,
			TagName: "test",
		}
		s := &sessions.Session{Context: context.TODO()}
		tagAssignIdFunc = func() types.ID {
			return 123
		}
		checkPermFunc = func(tx *gorm.DB, id types.ID, s0 *sessions.Session) error {
			Expect(id).To(Equal(types.ID(100)))
			Expect(*s).To(Equal(*s0))
			return nil
		}
		FindOrCreateTagFunc = func(tx *gorm.DB, q *TagCreate, s *sessions.Session) (*Tag, error) {
			return nil, sql.ErrConnDone
		}

		mock.ExpectBegin()
		mock.ExpectRollback()

		idObj, err := CreateTagAssign(&q, s)
		Expect(err).To(Equal(sql.ErrConnDone))
		Expect(idObj).To(BeNil())

		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})

	t.Run("should raise error of database error", func(t *testing.T) {
		_, mock := testinfra.SetUpMockSql()

		q := TagAssignCreate{
			ResID:   100,
			TagName: "test",
		}
		s := &sessions.Session{Context: context.TODO()}
		tagAssignIdFunc = func() types.ID {
			return 123
		}
		checkPermFunc = func(tx *gorm.DB, id types.ID, s0 *sessions.Session) error {
			Expect(id).To(Equal(types.ID(100)))
			Expect(*s).To(Equal(*s0))
			return nil
		}
		FindOrCreateTagFunc = func(tx *gorm.DB, q *TagCreate, s *sessions.Session) (*Tag, error) {
			return &Tag{ID: 2000, Name: "t2000", Note: "Tag2000", Image: "t2000.png"}, nil
		}

		const sqlExpr = "INSERT INTO `tag_assign` (`res_id`,`tag`,`res_type`,`tag_order`,`id`)" +
			" VALUES (?,?,?,?,?)"
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(sqlExpr)).
			WithArgs(100, 2000, 0, 0, 123).
			WillReturnError(sql.ErrConnDone)
		mock.ExpectRollback()

		resp, err := CreateTagAssign(&q, s)
		Expect(err).To(Equal(sql.ErrConnDone))
		Expect(resp).To(BeNil())

		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})

	t.Run("should be able to create tag assignment", func(t *testing.T) {
		_, mock := testinfra.SetUpMockSql()

		q := TagAssignCreate{
			ResID:   100,
			TagName: "test",
		}
		s := &sessions.Session{Context: context.TODO()}
		tagAssignIdFunc = func() types.ID {
			return 123
		}
		checkPermFunc = func(tx *gorm.DB, id types.ID, s0 *sessions.Session) error {
			Expect(id).To(Equal(types.ID(100)))
			Expect(*s).To(Equal(*s0))
			return nil
		}
		FindOrCreateTagFunc = func(tx *gorm.DB, q *TagCreate, s *sessions.Session) (*Tag, error) {
			return &Tag{ID: 2000, Name: "t2000", Note: "Tag2000", Image: "t2000.png"}, nil
		}

		const sqlExpr = "INSERT INTO `tag_assign` (`res_id`,`tag`,`res_type`,`tag_order`,`id`)" +
			" VALUES (?,?,?,?,?)"
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(sqlExpr)).
			WithArgs(100, 2000, 0, 0, 123).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		resp, err := CreateTagAssign(&q, s)
		Expect(err).ToNot(HaveOccurred())
		wantResp := TagAssignCreateResponse{
			TagAssignment: TagAssignment{
				ID: 123, ResID: 100, ResType: 0, TagID: 2000, TagOrder: 0,
			},
			TagName: "t2000", TagNote: "Tag2000", TagImage: "t2000.png",
		}
		Expect(*resp).To(Equal(wantResp))

		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})
}

func TestDeleteTagAssignWithQuery(t *testing.T) {
	RegisterTestingT(t)

	t.Run("should be able to delete tag assignment with query", func(t *testing.T) {
		_, mock := testinfra.SetUpMockSql()

		q := TagAssignRelation{
			ResID: 100,
			TagID: 10,
		}
		s := &sessions.Session{Context: context.TODO()}
		checkPermFunc = func(tx *gorm.DB, id types.ID, s0 *sessions.Session) error {
			Expect(id).To(Equal(types.ID(100)))
			Expect(*s).To(Equal(*s0))
			return nil
		}

		const sqlExpr = "DELETE FROM `tag_assign` WHERE res_id = ? AND tag = ? AND res_type = 0"
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(sqlExpr)).
			WithArgs(100, 10).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := DeleteTagAssignWithQuery(&q, s)
		Expect(err).ToNot(HaveOccurred())

		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})

	t.Run("should raise error of perm check", func(t *testing.T) {
		_, mock := testinfra.SetUpMockSql()

		q := TagAssignRelation{
			ResID: 100,
			TagID: 10,
		}
		s := &sessions.Session{Context: context.TODO()}
		checkPermFunc = func(tx *gorm.DB, id types.ID, s0 *sessions.Session) error {
			Expect(id).To(Equal(types.ID(100)))
			Expect(*s).To(Equal(*s0))
			return sql.ErrConnDone
		}

		mock.ExpectBegin()
		mock.ExpectRollback()

		err := DeleteTagAssignWithQuery(&q, s)
		Expect(err).To(Equal(sql.ErrConnDone))

		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})

	t.Run("should raise error of database error", func(t *testing.T) {
		_, mock := testinfra.SetUpMockSql()

		q := TagAssignRelation{
			ResID: 100,
			TagID: 10,
		}
		s := &sessions.Session{Context: context.TODO()}
		checkPermFunc = func(tx *gorm.DB, id types.ID, s0 *sessions.Session) error {
			Expect(id).To(Equal(types.ID(100)))
			Expect(*s).To(Equal(*s0))
			return nil
		}

		const sqlExpr = "DELETE FROM `tag_assign` WHERE res_id = ? AND tag = ? AND res_type = 0"
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(sqlExpr)).
			WithArgs(100, 10).
			WillReturnError(sql.ErrConnDone)
		mock.ExpectRollback()

		err := DeleteTagAssignWithQuery(&q, s)
		Expect(err).To(Equal(sql.ErrConnDone))

		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})
}
