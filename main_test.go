package main

import (
	"errors"
	"owlet/server/infra/app"
	"testing"

	. "github.com/onsi/gomega"
)

func TestMain(t *testing.T) {
	RegisterTestingT(t)

	t.Run("no panic in main", func(t *testing.T) {
		called := 0
		app.RunAppFunc = func() error {
			called++
			return nil
		}
		main()
		Expect(called).To(Equal(1))
	})

	t.Run("panic occurred in main", func(t *testing.T) {
		called := 0

		defer func() {
			Expect(called).To(Equal(1))
			if r := recover(); r == nil {
				t.Errorf("expected panic not happend")
			}
		}()

		app.RunAppFunc = func() error {
			called++
			return errors.New("some error")
		}
		main()
	})
}
