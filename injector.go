package injectables

import (
	"log"
	"reflect"
	"errors"
	"fmt"
)

const AnnotationTag = "injector"

type (

	Initer interface {
		Init() error
	}

	ServiceInjector interface {
		Get(name reflect.Type) (interface{}, error)
		Prepare(i interface{}) error
	}
)

// Create a new service injector and register all given services for
// later use!
func NewServiceInjector(services ...interface{}) ServiceInjector {

	inj := &defaultInjector{
		Services: []interface{}{},
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

	Services []interface{}
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

	if err := inj.Prepare(v); err != nil {
		return errors.New(fmt.Sprintf("unable to register service '%s': %v", t.String(), err))
	}

	if withInit, ok := v.(Initer); ok {
		if err := withInit.Init(); err != nil {
			return errors.New(fmt.Sprintf("unable to initialize service '%s': %v", t.String(), err))
		}
	}

	// Prepend the new service
	inj.Services = append([]interface{}{v}, inj.Services...)

	return nil
}

func (inj *defaultInjector) Get(t reflect.Type) (interface{}, error) {

	// log.Print("[Services] Request: ", t.String())
	for _, srv := range inj.Services {
		m := reflect.TypeOf(srv)

		// log.Print("Search inside: ", reflect.TypeOf(srv))

		if m.Implements(t) {
			return srv, nil
		}
	}

	return nil, errors.New("unable to find service that fulfills the requirements for :" + t.String())
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

		// Only work with annotated classes
		if _, ok := f.Tag.Lookup(AnnotationTag); !ok {
			continue
		}

		// Throw error on hidden attributes
		if !el.Field(i).CanSet() {
			return errors.New(fmt.Sprintf("the field '%s' is not accessible", f.Name))
		}

		service, err := inj.Get(f.Type)
		if err != nil {
			return err
		}

		log.Print("Found: ", f.Type, " => ", service)
		el.Field(i).Set(reflect.ValueOf(service))
	}

	return nil
}
