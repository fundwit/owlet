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

func TestTagAssignmentTableName(t *testing.T) {
	RegisterTestingT(t)

	t.Run("table name of TagAssignment should be correct", func(t *testing.T) {
		r := TagAssignment{}
		Expect(r.TableName()).To(Equal("tag_assign"))
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
