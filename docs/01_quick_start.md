---
title: Quick start 
summary: Grab your first data
---

## Installation

The agent currently supports Linux (`armv5`, `armv6`, `armv7`, `arm64` and `amd64`) and Windows (only `amd64`). The binaries are made available through [github releases](https://github.com/situation-sh/situation/releases/latest/).

You can also compile it from sources (once you have have a [go compiler `>=1.18`](https://go.dev/dl/)):

```shell
go install {{ variables.go_module }}/agent
```

## Quick run

You can run the agent without data persistence (in-memory database)

/// tab | Linux

```bash
situation run
```

///

/// tab | Windows

```ps1
situation.exe run
```

///

If you want to output an sqlite database, just add the `--db` flag

/// tab | Linux

```bash
situation run --db=situation.sqlite
```

///

/// tab | Windows

```ps1
situation.exe run --db=situation.sqlite
```

///

## Exploring (experimental)

Situation embeds a minimal terminal ui (tui) that briefly shows the collected data. It can pop up after the run with the `--explore` flag,

/// tab | Linux

```bash
situation run --explore
```

///

/// tab | Windows

```ps1
situation.exe run --explore
```

///

or the `explore` command can also be used in case of data persistence

/// tab | Linux

```bash
situation explore --db=situation.sqlite
```

///

/// tab | Windows

```ps1
situation.exe explore --db=situation.sqlite
```

///

![tui](img/tui.svg)

## Cooperation

Here is where the IT data collection platform starts!
You can let the agents cooperate by providing them a common postgres database:

/// tab | Linux

```bash
situation run --db=postgresql://[user]:[password]@[host]:[port]/[database]
```

///

/// tab | Windows

```ps1
situation.exe run --db=postgresql://[user]:[password]@[host]:[port]/[database]
```

///
