# Situation

[![pages-build-deployment](https://github.com/situation-sh/situation/actions/workflows/pages/pages-build-deployment/badge.svg)](https://situation-sh.github.io/situation/)
[![Go Report Card](https://goreportcard.com/badge/github.com/situation-sh/situation)](https://goreportcard.com/report/github.com/situation-sh/situation)
[![gosec](https://github.com/situation-sh/situation/actions/workflows/gosec.yaml/badge.svg)](https://github.com/situation-sh/situation/actions/workflows/gosec.yaml)
[![CodeQL](https://github.com/situation-sh/situation/actions/workflows/github-code-scanning/codeql/badge.svg)](https://github.com/situation-sh/situation/actions/workflows/github-code-scanning/codeql)

Situation provides the core infrastructure to automatically collect and consolidate IT data (machines, device, apps, network, flows...), on its own. 
Providing then an up-to-date and reliable view of the current state of your infra (or your home LAN), namely the *graph*.

Now you are ready to build a context-rich IT tool above Situation.

## Installation

The agent currently supports Linux and Windows. 

### Github releases

Pre-built binaries are available through [github releases](https://github.com/situation-sh/situation/releases/latest/).

### From sources

```shell
go install github.com/situation-sh/situation/agent@latest
```

## Quick start

You can run the agent directly (without data persistence) and explore what has been discovered (experimental terminal ui)

```bash
situation run --explore
```

![tui](docs/img/tui.svg)


## Data persistence

Once you gives a db (sqlite or postgres) to `situation` you enable data persistence. 

```bash
situation run --db=situation.sqlite
```

To go further, several agents can collaborate by sharing the same postgres db (you own). 
The **IT data collection starts here**!

```bash
situation run --db="postgres://user:password@example.org:5432/situation"
```

See the [docs](https://situation-sh.github.io/situation/) for more details.