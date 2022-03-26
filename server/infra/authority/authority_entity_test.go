package authority

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestPermissionsHasRole(t *testing.T) {
	RegisterTestingT(t)

	t.Run("HasRole", func(t *testing.T) {
		Expect(Permissions{"foo", "bar"}.HasRole("foo")).To(BeTrue())
		Expect(Permissions{"foo", "bar"}.HasRole("zoo")).To(BeFalse())
		Expect(Permissions{}.HasRole("foo")).To(BeFalse())
	})
}

func TestPermissionsHasAnyProjectRole(t *testing.T) {
	RegisterTestingT(t)

	t.Run("HasAnyProjectRole", func(t *testing.T) {
		Expect(Permissions{"foo_100", "bar_200"}.HasAnyProjectRole(100)).To(BeTrue())
		Expect(Permissions{"foo_100", "bar_200"}.HasAnyProjectRole(300)).To(BeFalse())
		Expect(Permissions{"100", "bar_200"}.HasAnyProjectRole(100)).To(BeFalse())
		Expect(Permissions{"_100", "bar_200"}.HasAnyProjectRole(100)).To(BeTrue())
	})
}

func TestPermissionsHasProjectRole(t *testing.T) {
	RegisterTestingT(t)

	t.Run("HasProjectRole", func(t *testing.T) {
		Expect(Permissions{"foo_100", "bar_200"}.HasProjectRole("foo", 100)).To(BeTrue())
		Expect(Permissions{"foo_100", "bar_200"}.HasProjectRole("foo", 200)).To(BeFalse())
	})
}

func TestPermissionsHasGlobalViewRole(t *testing.T) {
	RegisterTestingT(t)

	t.Run("HasGlobalViewRole", func(t *testing.T) {
		Expect(Permissions{"system:xxx", "bar_200"}.HasGlobalViewRole()).To(BeTrue())
		Expect(Permissions{"system:", "bar_200"}.HasGlobalViewRole()).To(BeTrue())
		Expect(Permissions{"foo_100", "bar_200"}.HasGlobalViewRole()).To(BeFalse())
		Expect(Permissions{}.HasGlobalViewRole()).To(BeFalse())
	})
}

func TestPermissionsHasProjectViewPerm(t *testing.T) {
	RegisterTestingT(t)

	t.Run("HasGlobalViewRole", func(t *testing.T) {
		Expect(Permissions{"system:xxx", "bar_200"}.HasProjectViewPerm(100)).To(BeTrue())
		Expect(Permissions{"system:", "bar_200"}.HasProjectViewPerm(200)).To(BeTrue())
		Expect(Permissions{"foo_100", "bar_200"}.HasProjectViewPerm(200)).To(BeTrue())
		Expect(Permissions{"foo_100", "bar_200"}.HasProjectViewPerm(300)).To(BeFalse())
		Expect(Permissions{}.HasProjectViewPerm(100)).To(BeFalse())
	})
}

func TestProjectRolesHasProject(t *testing.T) {
	RegisterTestingT(t)

	t.Run("HasGlobalViewRole", func(t *testing.T) {
		Expect(ProjectRoles{{ProjectID: 100}, {ProjectID: 200}}.HasProject(100)).To(BeTrue())
		Expect(ProjectRoles{{ProjectID: 100}, {ProjectID: 200}}.HasProject(300)).To(BeFalse())
		Expect(ProjectRoles{}.HasProject(100)).To(BeFalse())
	})
}
