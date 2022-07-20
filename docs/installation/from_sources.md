---
title: From sources
---

As the agent makes use of generics, you need to have the [Go compiler `>=1.18`](https://go.dev/dl/)

```shell
go install {{ variables.go_module }}
```

<!-- prettier-ignore -->
!!! warning
    [Pre-built binaries](pre_built_binaries.md) are compiled with extra flags to reduce the binary size and also set the version inside the binary. See the [Makefile](https://{{ variables.go_module }}/-/blob/main/Makefile) for more details.

<!-- prettier-ignore -->
!!! warning
    The `$GOPATH/bin` folder must be in your PATH to run `situation` directly from the command line
