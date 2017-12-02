## rocket

Simple hierarchical dependency injector based on go's reflection system and influenced by common dependency injection systems.

## Disclaimer

This is an very early version of an tool that I developed for my REST APIs where I was unable to provide proper
dependency indirection.

This is **not** stable yet and may change in the near future.

## Roadmap

[] 100% test coverage
[] Travis-CI

## Usage

**`Reminder: Services can depend on other services!`**

Create a new service definition. This way, it can be replaced during testing.

```go
NamingService interface {
  Names() []string
}
```

Create the default implementation for this service.
```go
StaticNamingService struct {}

func (s *StaticNamingService) Names() []string {
  return []{"Carl", "Michael", "Susanne"}
}
```

Create a new injector and provide instances of the service implementations.
```go
injector := rocket.NewInjector()
defer injector.Shutdown()

injector.Provide(
  &ConfigurationService{},
  &StaticNamingService{},
)
```

Use the injector to fulfill dependencies.
```go
type NamesController struct{
  namingService NamingService `rocket:""`
}

var ctrl NamesController
if err := injector.Prepare(&ctrl); err != nil {
  panic(err)
}

log.Print(ctrl.namingService.Names())
```

## Hierarchical injection / Overwrite injected services

Adding a new service with existing interfaces overwrites the previously existing one. Consequently, a previously 
used service can be injected in the newest one to be used. 

```go
BetterNamingService struct {
    Previous NamingService `rocket:""`
}

func (s *BetterNamingService) Names() []string {
	return append(s.Previous.Names(), "Vivian", "Marcus", "Nani"}
}
```

Will return ["Carl", "Michael", "Susanne", "Vivian", "Marcus", "Nani"] when used like this.

```go
injector := rocket.NewInjector()
defer injector.Shutdown()

injector.Provide(&StaticNamingService{})

// ... do something else ...

injector.Provide(&BetterNamingSystem{})
```

## Optional dependencies

When declaring dependencies with the `rocket` annotation, the option `rocket:",optional"` can be used to make the injector ignore the dependency if it can not be resolved.

Example
```go
type SomeCriticalController struct{
  logger Logger `rocket:",optional"`
}
```

## Lifecycle event handling

Sometimes, it may be useful to construct or deconstruct services before and after usage. To bring this into the overall workflow

```go
type DatabaseService struct{
  rocket.Initer
  mysql *sqlx.DB
}

func (d *DatabaseService) Init() error {
  // ... connect to the database
  return nil
}

func (d *DatabaseService) Shutdown() {
  // ... close all open connections ...
}

```

**To work, `defer injector.Shutdown()` is required.**

## Usage during testing

**[WIP]**

## Third-Party dependencies

None

## Last Update

02.12 - @joernlenoch

