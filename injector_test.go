package dieb_test

import (
	"github.com/joernlenoch/go-dieb"
	"testing"
)

type (
	Notifier interface {
		Hello(string) string
	}

	SomeService struct {
		Notifier
	}

	SomeOtherService struct {
		Previous Notifier `inject:""`
	}

	SomeController struct {
		Notifier Notifier `dieb:""`
	}
)

func (s SomeService) Hello(n string) string {
	return "Hello " + n
}

func (s SomeOtherService) Hello(n string) string {
	return s.Previous.Hello(n + " elloH")
}

func TestName(t *testing.T) {

	inj := dieb.NewInjector()
	defer inj.Shutdown()

	inj.Provide(
		&SomeService{},
		&SomeOtherService{},
	)

	var ctrl SomeController
	if err := inj.Prepare(&ctrl); err != nil {
		t.Fatal(err)
	}

	if ctrl.Notifier.Hello("World") != "Hello World elloH" {
		t.FailNow()
	}

	t.Log(ctrl.Notifier.Hello("Jones"))
}
