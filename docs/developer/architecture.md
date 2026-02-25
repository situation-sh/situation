# Architecture

The project is organized around a core library (`pkg/`) and an agent application (`agent/`).

## Package Overview

| Package       | Job                                                                                               |
| ------------- | ------------------------------------------------------------------------------------------------- |
| `agent`       | Agent entrypoint and CLI subcommands (run, id, task, update, version)                             |
| `pkg/models`  | Data structures representing discoverable entities (Machine, Application, NetworkInterface, etc.) |
| `pkg/modules` | All collection modules (plugins)                                                                  |
| `pkg/store`   | Database layer using [Bun ORM](https://bun.uptrace.dev/) (SQLite and PostgreSQL)                  |
| `pkg/utils`   | Extra helpers                                                                                     |

## Core Concepts

The overall architecture is plugin-based. A **scheduler** resolves module dependencies and runs the available modules in order. Each module receives a `context.Context` that carries:

- **logger**: a [logrus](https://github.com/Sirupsen/logrus) field logger scoped to the module
- **storage**: a `BunStorage` instance connected to a database (SQLite or PostgreSQL)
- **agent**: the agent identifier string

All collected data is persisted via the store into a relational database. 

