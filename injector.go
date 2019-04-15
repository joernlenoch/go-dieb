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

	// Initer describes all services that have an constructor-style method when
	// resolved.
	Initer interface {
		Init() error
	}

	// Shutdowner describe all services that have extra desconstruct-style methods
	// when shut down.
	Shutdowner interface {
		Shutdown()
	}

	// Injector is a interface for an service provider.
	Injector interface {

		// Get returns a instance of the requested type.
		Get(name reflect.Type) (interface{}, error)

		// GetByName returns the service by the provided name
		GetByName(s string) (interface{}, error)

		// MustPrepare tries to resolve all dependencies or panics.
		MustPrepare(i interface{})

		// Prepare tries to resolve all dependencies for the given struct
		Prepare(i interface{}) error

		// PrepareFunc calls the method will all dependencies resolved
		PrepareFunc(i interface{}) error

		// Provide add additional services, resolve their dependencies, and call the init method
		// of given services.
		Provide(v ...interface{}) error

		// Shutdown destroy the injector and calls all shutdown methods
		Shutdown()
	}

	// Config is a configuration for the static injector
	Config struct {
		Debug bool
	}
)

type StaticInjector struct {
	debug    bool
	services []interface{}
}

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
			return fmt.Errorf("unable to register service '%s': %v", t.String(), err)
		}

		var err error
		if withInit, ok := v.(Initer); ok {
			err = withInit.Init()
		} else if initMethod := reflect.ValueOf(v).MethodByName("Init"); initMethod.IsValid() {
			err = inj.PrepareFunc(initMethod)
		}

		if err != nil {
			return fmt.Errorf("unable to initialize service '%s': %v", t.String(), err)
		}

		// Prepend the new service
		inj.services = append([]interface{}{v}, inj.services...)
	}

	return nil
}

func (inj *StaticInjector) Shutdown() {
	for _, srv := range inj.services {

		// Skip the injector
		if srv == inj {
			continue
		}

		if withShutdown, ok := srv.(Shutdowner); ok {

			if inj.debug {
				log.Printf("[Injectables] Shutdown '%s'", reflect.TypeOf(srv))
			}

			withShutdown.Shutdown()
		}
	}
}

func (inj *StaticInjector) Get(t reflect.Type) (interface{}, error) {

	// log.Print("[Services] Request: ", t.String())
	for _, srv := range inj.services {
		m := reflect.TypeOf(srv)

		// log.Print("Search inside: ", reflect.TypeOf(srv))

		if m.Implements(t) {
			return srv, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("unable to find service that fulfills the requirements for '%s'", t.String()))
}

func (inj *StaticInjector) GetByName(s string) (interface{}, error) {

	// log.Print("[Services] Request: ", t.String())
	for _, srv := range inj.services {
		m := reflect.TypeOf(srv)

		log.Print("Search inside: ", m.String())

		if strings.HasSuffix(m.String(), s) {
			return srv, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("unable to find service that has the name '%s'", s))
}

func (inj *StaticInjector) MustPrepare(i interface{}) {
	if err := inj.Prepare(i); err != nil {
		panic(err)
	}
}

func (inj *StaticInjector) Prepare(target interface{}) error {

	if target == nil {
		return errors.New("the given object must not be nil")
	}

	if reflect.TypeOf(target).Kind() != reflect.Ptr {
		return errors.New("you must provide a pointer interface")
	}

	el := reflect.ValueOf(target).Elem()
	val := reflect.Indirect(reflect.ValueOf(target))
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

			return errors.New(fmt.Sprintf("unable to resolve dependency for '%s': %s", t.Name(), err.Error()))
		}

		if inj.debug {
			log.Printf("[Injectables] Resolve '%s' with '%s'", f.Name, reflect.TypeOf(service))
		}

		el.Field(i).Set(reflect.ValueOf(service))
	}

	return nil
}

func (inj *StaticInjector) PrepareFunc(fn interface{}) error {

	var method reflect.Value

	if reflected, ok := fn.(reflect.Value); ok {
		method = reflected
	} else {
		method = reflect.ValueOf(fn)
	}

	if method.Kind() != reflect.Func {
		return fmt.Errorf("can only prepare methods, %s given", method.Kind())
	}

	tm := method.Type()

	if tm.NumOut() != 1 {
		return fmt.Errorf("method '%s' must return a single error, got %d return values", tm.String(), tm.NumOut())
	}

	if tm.Out(0).Kind() != reflect.Interface {
		return fmt.Errorf("method '%s' must return a single error, not '%s'", tm.String(), tm.Out(0).Kind())
	}

	params := make([]reflect.Value, tm.NumIn())
	for i := 0; i < tm.NumIn(); i++ {
		typeIn := tm.In(i)

		if typeIn.Kind() != reflect.Ptr && typeIn.Kind() != reflect.Interface {
			return fmt.Errorf("required type must be a pointer or interface: %s", typeIn.String())
		}

		param, err := inj.Get(typeIn)
		if err != nil {
			return fmt.Errorf("unable to fullfil method '%s': %v", tm.String(), err)
		}
		params[i] = reflect.ValueOf(param)
	}

	if ret := method.Call(params); !ret[0].IsNil() {
		return errors.New(fmt.Sprintf("unable to fullfil method '%s': %v", tm.String(), ret[0]))
	}

	return nil
}

var _ Injector = (*StaticInjector)(nil)

// NewInjector creates a new empty injector.
func NewInjector() Injector {

	inj := &StaticInjector{
		services: []interface{}{},
		debug:    false,
	}

	// Provide itself as injector
	if err := inj.Provide(inj); err != nil {
		panic(err)
	}

	return inj
}

// NewInjectorWithConfig retcreates a new empty injector with the given configuration.
func NewInjectorWithConfig(cfg *Config) Injector {

	inj := &StaticInjector{
		services: []interface{}{},
		debug:    cfg.Debug,
	}

	return inj
}
