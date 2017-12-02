## rocket

Simple hierarchical dependency injector.

## Disclaimer

This is an very early version of an tool that I developed for my REST APIs where I was unable to provide proper
dependency indirection.

This is **not** stable yet.

## Usage

Create a new service definition.
```
NamingService interface {
    Names() []string
}
```

Create the default implementation for this service.
```
StaticNamingService struct {
    ConfigurationService
}

func (s *StaticNamingService) Names() []string {
	return []{"Carl", "Michael", "Susanne"}
}
```

Create a new injector and provide instances of the service implementations.
```
injector := rocket.NewInjector()
defer injector.Shutdown()

injector.Register(
    &StaticNamingService{},
)
```

Use the injector to fulfill dependencies.
```
type NamesController struct{
    namingService NamingService `rocket:""`
}

var ctrl NamesController
if err := injector.Prepare(&ctrl); err != nil {
    panic(err)
}

log.Print(ctrl.namingService.Names())
```

## Hierarchical Injection

TBD

## Third-Party dependencies

None

## Last Update

02.12 - @joernlenoch

