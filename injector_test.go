package dieb_test

import (
	"github.com/joernlenoch/go-dieb"
	"github.com/stretchr/testify/assert"
	"log"
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

func (s *SomeOtherService) Init(injector dieb.Injector, someSvc Notifier) error {

	log.Print(someSvc.Hello("User"))

	return nil
}

func (s SomeService) Hello(n string) string {
	return "Hello " + n
}

func (s SomeOtherService) Hello(n string) string {
	return s.Previous.Hello(n + " elloH")
}

func TestName(t *testing.T) {

	inj := dieb.NewInjector()
	defer inj.Shutdown()

	err := inj.Provide(
		&SomeService{},
		&SomeOtherService{},
	)

	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
		return
	}

	var ctrl SomeController
	if err := inj.Prepare(&ctrl); err != nil {
		t.Fatal(err)
	}

	if ctrl.Notifier.Hello("World") != "Hello World elloH" {
		t.FailNow()
	}

	t.Log(ctrl.Notifier.Hello("Jones"))
}
