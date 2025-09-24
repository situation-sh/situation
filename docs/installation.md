---
title: Installation
summary: How to get Situation
order: 10
---

## Releases

The agent currently supports Linux and Windows on x86_64 architectures. The binaries are made available through [github releases](https://github.com/situation-sh/situation/releases/tag/{{ latest_tag }}).

/// tab | `wget`

```bash
wget -qO situation {{ github_repo }}/releases/download/{{ latest_tag }}/{{ latest_linux_binary }}
chmod +x ./situation
```
///

/// tab | `curl`

```bash
curl -sLo ./situation {{ github_repo }}/releases/download/{{ latest_tag }}/{{ latest_linux_binary }}
chmod +x ./situation
```

///

/// tab | `PowerShell`

```ps1
Invoke-RestMethod -OutFile situation.exe -Uri {{ github_repo }}/releases/download/{{ latest_tag }}/{{ latest_windows_binary }}
```
///

## From sources

As the agent makes use of generics, you need to have the [go compiler `>=1.18`](https://go.dev/dl/)

```shell
go install {{ variables.go_module }}
```


!!! warning ""
    [Pre-built binaries](#releases) are compiled with extra flags to reduce the binary size and also set the version inside the binary. See the [Makefile](https://{{ variables.go_module }}/-/blob/main/Makefile) for more details.


!!! warning ""
    The `$GOPATH/bin` folder must be in your PATH to run `situation` directly from the command line

