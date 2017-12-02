package injectables

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
)

const AnnotationTag = "injector"

type (
	Initer interface {
		Init() error
	}

	Shutdowner interface {
		Shutdown()
	}

	//
	//
	//
	Injector interface {
		//
		//
		Get(name reflect.Type) (interface{}, error)

		//
		//
		Prepare(i interface{}) error

		//
		//
		Register(v ...interface{}) error

		//
		//
		Shutdown()
	}

	Config struct {
		Debug bool
	}
)

// Create a new service injector and register all given services for
// later use!
func NewInjector() Injector {

	inj := &defaultInjector{
		services: []interface{}{},
		debug:    false,
	}

	return inj
}

//
//
func NewInjectorWithConfig(cfg *Config) Injector {

	inj := &defaultInjector{
		services: []interface{}{},
		debug:    cfg.Debug,
	}

	return inj
}

//
//
//
type defaultInjector struct {
	Injector

	debug    bool
	services []interface{}
}

//
//
//
func (inj *defaultInjector) Register(r ...interface{}) error {

	for _, v := range r {

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
		inj.services = append([]interface{}{v}, inj.services...)
	}

	return nil
}

//
//
//
func (inj *defaultInjector) Shutdown() {

	log.Print("SHUTDOWN")

	for _, srv := range inj.services {
		if withShutdown, ok := srv.(Shutdowner); ok {

			if inj.debug {
				log.Printf("[Injectables] Shutdown '%s'", reflect.TypeOf(srv))
			}

			withShutdown.Shutdown()
		}
	}
}

//
//
//
func (inj *defaultInjector) Get(t reflect.Type) (interface{}, error) {

	// log.Print("[Services] Request: ", t.String())
	for _, srv := range inj.services {
		m := reflect.TypeOf(srv)

		// log.Print("Search inside: ", reflect.TypeOf(srv))

		if m.Implements(t) {
			return srv, nil
		}
	}

	return nil, errors.New("unable to find service that fulfills the requirements for :" + t.String())
}

//
//
//
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
		tag, ok := f.Tag.Lookup(AnnotationTag)
		if !ok {
			if inj.debug {
				log.Printf("[Injectables] Skip '%s': No Annotation found", f.Name)
			}
			continue
		}

		// Throw error on hidden attributes
		if !el.Field(i).CanSet() {
			return errors.New(fmt.Sprintf("the field '%s' is not accessible", f.Name))
		}

		service, err := inj.Get(f.Type)
		if err != nil {
			// Ignore errors when marked as optional
			if strings.Contains(tag, ",optional") {

				if inj.debug {
					log.Printf("[Injectables] Skip '%s': Not found; Marked as optional", f.Name)
				}

				continue
			}

			return err
		}

		if inj.debug {
			log.Printf("[Injectables] Resolve '%s' with '%s'", f.Name, reflect.TypeOf(service))
		}

		el.Field(i).Set(reflect.ValueOf(service))
	}

	return nil
}
