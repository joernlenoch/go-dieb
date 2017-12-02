package injectables

import (
	"log"
	"reflect"
	"errors"
	"fmt"
)

const FieldTag = "injector"

type (

	Injectable interface {}

	InjectableWithInit interface {
		Injectable
		Init() error
	}

	ServiceInjector interface {
		Get(name reflect.Type) (Injectable, error)
		Prepare(i interface{}) error
	}
)

// Create a new service injector and register all given services for
// later use!
func NewServiceInjector(services ...Injectable) ServiceInjector {

	inj := &defaultInjector{
		Services: []Injectable{},
	}

	for _, s := range services {
		if err := inj.Register(s); err != nil {
			log.Fatal(err)
		}
	}

	return inj
}


//
//
//
type defaultInjector struct {
	ServiceInjector

	Services []Injectable
}

func (inj *defaultInjector) Register(v interface{}) error {

	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Ptr {
		return errors.New(fmt.Sprintf("services must be given as pointer: %v", t.Kind()))
	}

	t = t.Elem()

	// log.Print("[Services] Register: ", t)

	// if err := v.Init(inj); err != nil {
	//	return errors.New(fmt.Sprintf("unable to register service '%s' : %v", v.ServiceName(), err))
	// }

	if _, ok := v.(Injectable); !ok {
		return errors.New("the given service must implement the 'Injectable' interface")
	}

	if err := inj.Prepare(v); err != nil {
		return errors.New(fmt.Sprintf("unable to register service '%s': %v", t.String(), err))
	}

	if withInit, ok := v.(InjectableWithInit); !ok {
		if err := withInit.Init(); err != nil {
			return errors.New(fmt.Sprintf("unable to initialize service '%s': %v", t.String(), err))
		}
	}


	inj.Services = append(inj.Services, v)

	return nil
}

func (inj *defaultInjector) Get(t reflect.Type) (Injectable, error) {

	// log.Print("[Services] Request: ", t.String())
	var found Injectable
	for _, srv := range inj.Services {
		m := reflect.TypeOf(srv)

		// log.Print("Search inside: ", reflect.TypeOf(srv))

		if m.Implements(t) {
			found = srv
		}
	}

	if found == nil {
		return nil, errors.New("unable to find service that fulfills the requirements for :" + t.String())
	}

	return found, nil
}

func (inj *defaultInjector) Prepare(i interface{}) error {

	if i == nil {
		return errors.New("the given object must not be nil")
	}

	if reflect.TypeOf(i).Kind() != reflect.Ptr {
		return errors.New("you must provide a pointer interface")
	}

	el := reflect.ValueOf(i).Elem()
	val := reflect.Indirect(reflect.ValueOf(i))
	t := val.Type()

	// log.Print("PREPARE: ",  val.Type().Name(), t.NumField())

	for i := 0; i < t.NumField(); i++ {
		f := val.Type().Field(i)

		// Allow users to skip the injection
		if f.Tag.Get(FieldTag) == ",skip" {
			continue
		}

		// Skip all interface inheritance
		if f.Anonymous {
			// log.Print("Skip:", f.Name, f.Type)
			continue
		}

		// Skip all interface inheritance
		if !el.Field(i).CanSet() {
			return errors.New(fmt.Sprintf("the field '%s' is not accessible", f.Name))
		}

		service, err := inj.Get(f.Type)
		if err != nil {
			return err
		}

		// log.Print(f.Type, " => ", reflect.TypeOf(service))
		el.Field(i).Set(reflect.ValueOf(service))
	}

	return nil
}