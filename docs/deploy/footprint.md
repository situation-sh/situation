---
title: Limiting footprint
summary: Bound CPU and memory usage
---
Here are two environment variables that can be used to limit the CPU and memory usage of the agent. It 

The `GOMAXPROCS` variable limits the number of operating system threads that can execute user-level Go code simultaneously[^1].

/// tab | Linux

```shell
GOMAXPROCS=1 ./situation
```

///

/// tab | Windows

```ps1
& { $env:GOMAXPROCS = "1"; .\situation.exe }
```

///

The `GOMEMLIMIT` variable sets a soft memory limit for the runtime. A zero limit or a limit that's lower than the amount of memory used by the Go runtime may cause the garbage collector to run nearly continuously. However, the application may still make progress[^1].

/// tab | Linux

```shell
GOMEMLIMIT=10MiB ./situation
```

///

/// tab | Windows

```ps1
& { $env:GOMEMLIMIT = "10MiB"; .\situation.exe }
```

///

[^1]: See [documentation](https://pkg.go.dev/runtime).