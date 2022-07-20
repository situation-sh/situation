# Modules 

## Introduction

A module is an _independent_ piece of code that can be run during scan. Its job is merely to enrich the store.
It is not fully independent as it may depend on previous modules (some module are likely to need data provided by others).

To develop a module, just init a new `my_new_module.go` source file in the `modules/` subdirectory. The structure of the module should look like the following snippet.

```go
package modules

import (
    // ...
)

func init() {
    m := &MyNewModule{}
    RegisterModule(m)
    // SetDefault
    SetDefault(m, "myparam", value, "the value of myparam")
    // ...
}

type MyNewModule struct {}

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
    // you can grab config with GetConfig.
    // :warning: generics are used to access the config
    // GetConfig[<TYPE>](>MODULE>, <KEY>)
    myParam := GetConfig[time.Duration](m, "myparam")
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
- If you wan to create a module called "awesome stuff":
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

The `Name()` outputs the [unique] name of the module.

The `Dependencies()` returns the name of the modules required to start this module (prior information).

The `Run()` function does the job. This functions is called during the scan. It may have several interactions:

- [config](#configuration) (get extra configuration data)
- [logging](#logging) (output some information about the run)
- [store](#the-store) (retrieve/store collected data)

## Configuration

The configuration is only managed by the flags of [urfave/cli](https://github.com/urfave/cli).

The configuration of the modules are stored in the `modules.module-name.*` namespace. To hide it to the developper, the `modules` package expose two helpers:

```go
// GetConfig is a generic function that returns a value
// associated to a key within the module namespace
func GetConfig[T any](m Module, key string) (T, error) {
	// ...
}

// SetDefault is a helper that defines default module parameter.
// The provided values can be overwritten by CLI flags or config file.
func SetDefault(m Module, key string, value interface{}, usage string) {
    // ...
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
