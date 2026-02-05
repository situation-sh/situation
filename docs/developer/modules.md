______________________________________________________________________

## title: Module summary: Independent piece of magic

## Introduction

A module is an _independent_ piece of code that can be run during scan. Its job is merely to enrich the store.
It is not fully independent as it may depend on previous modules (some module are likely to need data provided by others).

To develop a module, just init a new `my_new_module.go` source file in the `modules/` subdirectory. The structure of the module should look like the following snippet.

```go
package modules

import (
    // ...
)

type MyNewModule struct {
    Attribute string
}

func init() {
    m := &MyNewModule{
        Attribute: "defultValue"
    }
    RegisterModule(m)
    // bind attributes with configuration variable (the attribute will be exposed to CLI flags)
    SetDefault(m, "attribute", &m.Attribute, "Custom attribute for my new module")
}



// Name returns the name of the module
func (m * MyNewModule) Name() string {
    // return the name of the module with a dash
    return "my-new-module"
}

// Dependencies return the list of modules
// required to run this one
func (m * MyNewModule) Dependencies() []string {
    // put the name of the modules you depend on here
    return []string{"host-basic"}
}

// Run do the job. It returns error only if it really
// fails, i.e. it cannot be run (like privileges).
// In the other cases, just log the errors
func (m * MyNewModule) Run() error {
    // you can grab your logger (from https://github.com/Sirupsen/logrus)
    logger := GetLogger(m)
    // ...
    // do what you want
    // ...
    // but do not return error except if something
    // prevents the module to be run, just log them:
    // logger.Error(err)
    // ...
    //
    // don't forget to put data into the store
    // ...
    return nil
}


```

## Naming

You are free about the module naming, but obviously there are some constraints:

- the module name must be unique
- the name should describe what the module does (or the ecosystem, like "docker")
- If you want to create a module called "awesome stuff":
  - its name (output of `.Name()`) must be `awesome-stuff`
  - the object that respects the `Module` interface must be `AwesomeStuffModule`
  - the source file must be `awesome_stuff.go`

## Module interface

A module must implement the `Module` interface described below.

```go
// Module is the generic module interface to implement plugins to
// the agent
type Module interface {
	Name() string
	Dependencies() []string
	Run() error
}
```

The `Name()` outputs the **unique** name of the module.

The `Dependencies()` returns the names of the modules required to start this module (prior information).

The `Run()` function does the job. This functions is called during the scan. It may have several interactions:

- [config](#configuration) (get extra configuration data)
- [logging](#logging) (output some information about the run)
- [store](store.md) (retrieve/store collected data)

## Configuration

The configuration is managed by [asiffer/puzzle](https://github.com/asiffer/puzzle). As the example above, you should put the required information into the base module struct, along with a relevant default value. If you want to let the user modify attributes, you should bind your struct attribute with the configuration, through the following helper:

```go
// SetDefault is a helper that defines default module parameter.
// The provided values can be overwritten by CLI flags, env variables or anything
// the asiffer/puzzle library may support.
func SetDefault[T any](m Module, key string, value *T, usage string) {
	// ...
}
```

The configuration of the modules are stored in the `modules.module-name.*` namespace in the `config` module, so it can be accessed by other modules through `config.Get[T](key string)`. In your code (like in the `Run()` function), you should directly access the attributes through the pointer receiver.

```go
func (m *MyNewModule) Run() error {
    // do not get it through the config
    attr, err := config.Get[string]("modules.my-new-module.attribute")
    // rather access it directly
    attr := m.Attribute
}
```

## Logging

The logging is managed by [logrus](https://github.com/Sirupsen/logrus). To log some information, the `modules` package expose a `GetLogger` function that returns a contextual logger (relative to the module).

```go
func (m * MyModule) Run() error {
    // ...
    logger := GetLogger(m)
    // now you can use the classical methods
    logger.Debug("Debug message")
    logger.Info("Info message")
    logger.Warn("Warning message")
    logger.Error("Error message")
    logger.Fatal("Fatal error")
    // you should avoid logger.Panic to prevent the agent from crashing

    // ...
}
```

In addition, the module is likely to collect some information. You can log the collected data in a structured manner with the `logger.WithField` method.

```go
func (m * MyModule) Run() error {
    // ...
    logger := GetLogger(m)
    // ...
    // append the fields you want to show and call Debug/Info method
    logger.WithField("hostname", hostname).Debug("Hostname found!")
}
```

## Big module case

if your module is heavy you can store all the work (namely the material for the `Run` function) inside a submodule and write a short interface in the `modules` directory.

You may have the following layout:

```bash
modules/
    heavy.go
    heavy/
        file1.go
        file2.go
        ...
```

The `heavy.go` file may look like the following:

```go

import (
    // ...

	// load the submodule
	 "github.com/situation-sh/situation/modules/heavy"

)

type HeavyModule struct {}

func init() {
    RegisterModule(&HeavyModule{})
}

func (m * HeavyModule) Name() string {
    // return the name of the module with a dash
    return "heavy"
}

func (m * HeavyModule) Dependencies() []string {
    // put the name of the modules you depend on here
    return []string{}
}

func (m * HeavyModule) Run() error {
    // ...
    // call heavy.Stuff
    // ...
}
```

## Documentation

Documenting a module is mandatory. There are two things to do. The first thing is to document the `Module` object as follows:

```go
// MyNewModule retrieves data from ...
//
// It mainly depends on the following external library:
//  - ... 
//
// On Windows, it collect data by calling...
// On Linux, it reads ...
type MyNewModule struct {}
```

One must have a synospis (first line) and then some details about the module.
One may include how data is collected with regards to the platform and also
other relevant things (edge cases, libraries, privileges, options etc.)

The second point is to fill some standard notes, as follows:

```go
// LINUX(MyNewModule) ok
// WINDOWS(MyNewModule) ok
// MACOS(MyNewModule) ?
// ROOT(MyNewModule) no
package modules
```

The format of the note is given by the [doc](https://pkg.go.dev/go/doc#Note) package. We use it as follows: `<KEY>(<MODULE-NAME>) <VALUE>`

Currently there are 4 attributes to provide: `LINUX`, `WINDOWS`, `MACOS` and `ROOT`. Their corresponding values must be
`yes`/`ok` (meaning "supported"), `no` (meaning "not supported"), or `?` (meaning "don't know").

!!! warning ""
For `ROOT`, `yes`/`ok` means that root privileges are required
