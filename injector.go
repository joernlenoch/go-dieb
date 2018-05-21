package dieb

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
)

var AnnotationTags = []string{"rocket", "dieb", "inject"}

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
		MustPrepare(i interface{})

		//
		//
		Prepare(i interface{}) error

		//
		//
		Provide(v ...interface{}) error

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

	inj := &StaticInjector{
		services: []interface{}{},
		debug:    false,
	}

	return inj
}

//
//
func NewInjectorWithConfig(cfg *Config) Injector {

	inj := &StaticInjector{
		services: []interface{}{},
		debug:    cfg.Debug,
	}

	return inj
}

//
//
//
type StaticInjector struct {
	Injector

	debug    bool
	services []interface{}
}

//
//
//
func (inj *StaticInjector) Provide(r ...interface{}) error {

	for _, v := range r {

		t := reflect.TypeOf(v)
		if t.Kind() != reflect.Ptr {
			return errors.New(fmt.Sprintf("services must be given as pointer: %v", t.Kind()))
		}

		t = t.Elem()

		// log.Print("[Services] Provide: ", t)

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
func (inj *StaticInjector) Shutdown() {
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
func (inj *StaticInjector) Get(t reflect.Type) (interface{}, error) {

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
func (inj *StaticInjector) MustPrepare(i interface{}) {
	if err := inj.Prepare(i); err != nil {
		panic(err)
	}
}

//
//
//
func (inj *StaticInjector) Prepare(i interface{}) error {

	if i == nil {
		return errors.New("the given object must not be nil")
	}

	if reflect.TypeOf(i).Kind() != reflect.Ptr {
		return errors.New("you must provide a pointer interface")
	}

	el := reflect.ValueOf(i).Elem()
	val := reflect.Indirect(reflect.ValueOf(i))
	t := val.Type()

	for i := 0; i < t.NumField(); i++ {
		f := val.Type().Field(i)

		var ok bool
		var tag string

		// Only work with annotated classes
		for _, tagName := range AnnotationTags {
			tag, ok = f.Tag.Lookup(tagName)
			if ok {
				break
			}

		}

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
