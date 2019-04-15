## Dieb (former 'rocket')

Simple hierarchical dependency injector based on go's reflection system and influenced by common dependency injection handlers.
 
Focuses on a clear and reusable structure, without pitfalls.

## Personal Note

> Currently in production in 3 of my larger projects (35k loc+).
> 
> Works very reliable and we never had any issues.
>
> -- len
 

## Roadmap

- [x] indirect dependancies via interface

- [x] direct dependancies via pointer

- [x] preparation of structs

- [x] constructor methods

- [ ] 100% test coverage

- [ ] Travis-CI

## Usage

**`Reminder: Services can depend on other services and references!`**

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
injector := dieb.NewInjector()
defer injector.Shutdown()

injector.Provide(
  &ConfigurationService{},
  &StaticNamingService{},
)
```

Use the injector to fulfill dependencies.
```go
type NamesController struct{
  NamingService       NamingService `dieb:""` // This is an indirect dependancy
  StaticNamingService *StaticNamingService `dieb:""` // This is a direct dependancy
}

var ctrl NamesController
if err := injector.Prepare(&ctrl); err != nil {
  panic(err)
}

log.Print(ctrl.NamingService.Names())
```

## Init / Constructor methods

Since Version `v2.1.0` it is possible to inject via method.

```go
func ConstructorMethod(/* deps */) error {
  
  // [...]
  // Do something with the deps
  // [...]
  
  return nil
}

err := inj.PrepareFunc(ConstructorMethod)
if err != nil {
  panic(err)
}

```

**The `Init(..) error` method will automatically be called when a service is provided!**


## Hierarchical injection / Overwrite injected services

Adding a new service with existing interfaces overwrites the previously existing one. Consequently, a previously 
used service can be injected in the newest one to be used. 

```go
BetterNamingService struct {
    Previous NamingService `dieb:""`
}

func (s *BetterNamingService) Names() []string {
	return append(s.Previous.Names(), "Vivian", "Marcus", "Nani"}
}
```

Will return ["Carl", "Michael", "Susanne", "Vivian", "Marcus", "Nani"] when used like this.

```go
injector := dieb.NewInjector()
defer injector.Shutdown()

injector.Provide(&StaticNamingService{})

// ... do something else ...

injector.Provide(&BetterNamingSystem{})
```

#### Use-Cases

A typical example could be a `StorageService` and a `CachingService` that provides a storage interface,
but applies a custom caching strategy.

## Optional dependencies

When declaring dependencies with the `dieb` annotation, the option `dieb:",optional"` can be used to make the injector ignore the dependency if it can not be resolved.

Example
```go
type SomeCriticalController struct{
  logger Logger `dieb:",optional"`
}
```

## Lifecycle event handling

Sometimes, it may be useful to construct or deconstruct services before and after usage. To bring this into the overall workflow

```go
type DatabaseService struct{
  mysql *sqlx.DB
}

// dieb.Initer
func (d *DatabaseService) Init() error {
  // ... connect to the database
  return nil
}

// custom Init with dependancies
func (d *DatabaseService) Init(injector dieb.Injector, db *SomeService) error {
  // ... connect to the database
  return nil
}

// dieb.Shutdowner
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

