package sessions_test

import (
	"owlet/server/infra/authority"
	"owlet/server/infra/sessions"
	session "owlet/server/infra/sessions"
	"testing"

	"github.com/fundwit/go-commons/types"
	. "github.com/onsi/gomega"
)

func TestHasRole(t *testing.T) {
	RegisterTestingT(t)

	t.Run("should work correctly", func(t *testing.T) {
		c := session.Session{}
		Expect(c.Perms.HasRole("aaa")).To(BeFalse())

		c = session.Session{Perms: []string{}}
		Expect(c.Perms.HasRole("aaa")).To(BeFalse())

		c = session.Session{Perms: []string{"bbb", "ccc"}}
		Expect(c.Perms.HasRole("aaa")).To(BeFalse())

		c = session.Session{Perms: []string{"bbb", "ccc"}}
		Expect(c.Perms.HasRole("ccc")).To(BeTrue())
	})
}

func TestVisibleProjects(t *testing.T) {
	RegisterTestingT(t)

	t.Run("should work as expected", func(t *testing.T) {
		c := sessions.Session{Perms: authority.Permissions{"aa_10", "bb_20", "30", "40_abc"}}
		Expect(c.VisibleProjects()).To(Equal([]types.ID{10, 20}))

		c = session.Session{Perms: authority.Permissions{}}
		Expect(c.VisibleProjects()).To(Equal([]types.ID{}))

		c = session.Session{}
		Expect(c.VisibleProjects()).To(Equal([]types.ID{}))
	})
}
