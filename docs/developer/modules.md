---
title: Module 
summary: Independent piece of magic
---

## Introduction

A module is an _independent_ piece of code that can be run during scan. Its job is merely to enrich the store.
It is not fully independent as it may depend on previous modules (some module are likely to need data provided by others).

To develop a module, create a new `my_new_module.go` source file in the `pkg/modules/` directory. The structure of the module should look like the following snippet.

```go
package modules

import (
    "context"
    // ...
)

func init() {
    registerModule(&MyNewModule{
        attribute: "defaultValue",
    })
}

// Module definition ------------------------------------------------

type MyNewModule struct {
    BaseModule

    attribute string
}

// Name returns the name of the module
func (m *MyNewModule) Name() string {
    return "my-new-module"
}

// Dependencies return the list of modules
// required to run this one
func (m *MyNewModule) Dependencies() []string {
    return []string{"host-basic"}
}

// Run does the job. It returns error only if it really
// fails, i.e. it cannot be run (like privileges).
// In the other cases, just log the errors
func (m *MyNewModule) Run(ctx context.Context) error {
    // extract the logger and storage from the context
    logger := getLogger(ctx, m)
    storage := getStorage(ctx)

    // ...
    // do what you want
    // ...
    // but do not return error except if something
    // prevents the module to be run, just log them:
    // logger.
    //      WithError(err).
    //      Warn("something wrong but not critical happens")
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
    Run(ctx context.Context) error
}
```

The `Name()` outputs the **unique** name of the module.

The `Dependencies()` returns the names of the modules required to start this module (prior information).

The `Run(ctx)` function does the job. This function is called during the scan by the scheduler. The `context.Context` carries the logger, storage, and agent identifier. The function may have several interactions:

- [config](#configuration) (get extra configuration data)
- [logging](#logging) (output some information about the run)
- [store](store.md) (retrieve/store collected data)

### BaseModule

All modules should embed `BaseModule`:

```go
type MyNewModule struct {
    BaseModule
    // your fields here
}
```

Currently this object does not provide extra attribute/methods. But this is where we could inject ones in the future.

### Registration

Modules must be registered via `init()` using the unexported `registerModule` function. 
This adds the module to the internal map. It panics if two modules share the same name.

```go
func init() {
    registerModule(&MyNewModule{})
}
```

## Context

The `Run` method receives a `context.Context` prepared by the scheduler.
In this catch-all parameter, we provide all the runtile needs of the module.
Currently, three helpers extract what you need:

| Helper              | Returns              | Description                      |
| ------------------- | -------------------- | -------------------------------- |
| `getLogger(ctx, m)` | `logrus.FieldLogger` | Logger scoped to the module name |
| `getStorage(ctx)`   | `*store.BunStorage`  | Database storage instance        |
| `getAgent(ctx)`     | `string`             | Agent identifier                 |

```go
func (m *MyNewModule) Run(ctx context.Context) error {
    logger  := getLogger(ctx, m)
    storage := getStorage(ctx)
    agent   := getAgent(ctx)
    // ...
}
```

## Configuration

The configuration is managed by [asiffer/puzzle](https://github.com/asiffer/puzzle). If your module needs configurable attributes, put them in the module struct with a default value and implement the `Configurable` interface defined in `agent/config`:

```go
type Configurable interface {
    Bind(config *puzzle.Config) error
}
```

Inside `Bind`, use the `setDefault` helper to register parameters:

```go
func (m *MyNewModule) Bind(config *puzzle.Config) error {
    return setDefault(config, m, "attribute", &m.Attribute,
        "Custom attribute for my new module")
}
```

The parameters are stored in the `modules.module-name.*` namespace and are automatically exposed as CLI flags. In your `Run()` function, access attributes directly through the pointer receiver:

```go
func (m *MyNewModule) Run(ctx context.Context) error {
    // access it directly 
    attr := m.attribute
    // ...
}
```

## Logging

The logging is managed by [logrus](https://github.com/Sirupsen/logrus). To log some information, extract the logger from the context with `getLogger`. This returns a logger automatically scoped to the module name.

```go
func (m *MyNewModule) Run(ctx context.Context) error {
    logger := getLogger(ctx, m)
    // now you can use the classical methods
    logger.Debug("Debug message")
    logger.Info("Info message")
    logger.Warn("Warning message")
    logger.Error("Error message")
    // you should avoid logger.Panic to prevent the agent from crashing
    // ...
}
```

You should log collected data in a structured manner with `logger.WithField`:

```go
func (m *MyNewModule) Run(ctx context.Context) error {
    logger := getLogger(ctx, m)
    // ...
    logger.WithField("hostname", hostname).Debug("Hostname found!")
}
```

## Big module case

If your module is heavy you can store the implementation inside a sub-package and write a short interface in the `modules` directory.

You may have the following layout:

```bash
pkg/modules/
    heavy.go
    heavy/
        file1.go
        file2.go
        ...
```

The `heavy.go` file may look like the following:

```go
package modules

import (
    "context"

    // load the sub-package
    heavy "github.com/situation-sh/situation/pkg/modules/heavy"
)

type HeavyModule struct {
    BaseModule
}

func init() {
    registerModule(&HeavyModule{})
}

func (m *HeavyModule) Name() string {
    return "heavy"
}

func (m *HeavyModule) Dependencies() []string {
    return []string{}
}

func (m *HeavyModule) Run(ctx context.Context) error {
    logger := getLogger(ctx, m)
    storage := getStorage(ctx)
    // delegate to the sub-package
    return heavy.DoWork(ctx, logger, storage)
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
type MyNewModule struct {
    BaseModule
}
```

One must have a synopsis (first line) and then some details about the module.
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
