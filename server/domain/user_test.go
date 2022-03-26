package domain

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestUserTableName(t *testing.T) {
	RegisterTestingT(t)

	t.Run("table name of user should be correct", func(t *testing.T) {
		r := User{}
		Expect(r.TableName()).To(Equal("user"))
	})
}

func TestUserIdentityTableName(t *testing.T) {
	RegisterTestingT(t)

	t.Run("table name of user identity should be correct", func(t *testing.T) {
		r := UserIdentity{}
		Expect(r.TableName()).To(Equal("user_identity"))
	})
}
