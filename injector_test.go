package injectables_test

import (
	"testing"
	"github.com/joernlenoch/go-injectables"
)

type (

	NotificationService interface {
		injectables.Injectable
		Hello(string) string
	}

	SomeService struct {
		NotificationService
	}

	SomeOtherService struct {
		NotificationService

		Previous NotificationService
	}

	SomeController struct {
		Notify NotificationService
	}
)

func (s SomeService) Hello(n string) string {
	return "Hello " + n
}

func (s SomeOtherService) Hello(n string) string {
	return s.Previous.Hello(n + " elloH")
}


func TestName(t *testing.T) {

	inj := injectables.NewServiceInjector(
		&SomeService{},
		&SomeOtherService{},
	)

	var ctrl SomeController
	if err := inj.Prepare(&ctrl); err != nil {
		t.Fatal(err)
	}

	if ctrl.Notify.Hello("World") != "Hello World elloH" {
		t.FailNow()
	}

	t.Log(ctrl.Notify.Hello("Jones"))
}
