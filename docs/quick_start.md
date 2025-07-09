---
title: Quick start
summary: What you can do with this CLI tool
order: 20
---
## Run

You guess it. To run the agent, you don't need to provide any extra configuration.

/// tab | Linux

```bash
situation --stdout
```

///

/// tab | Windows

```ps1
situation.exe --stdout
```

///

You will see the json output in the terminal. 



So what you should do next, is to **pipe that json to another tool** (like `jq` see in the [guide](./10_guides/jq-one-liners.md)).

## Update

The agent can update itself. It queries Github releases and check if there is a newer version.

/// tab | Linux

```bash
situation update
```

///


/// tab | Windows

```ps1
situation.exe update
```

///

If the new version has breaking changes (when the major version is different) you should pass the `--force` flag to make the update.


## Other commands

| Command           | Description                                              |
| ----------------- | -------------------------------------------------------- |
| `refresh-id`      | Regenerate the internal ID of the agent                  |
| `defaults`, `def` | Print the default config                                 |
| `id`              | Print the identifier of the agent                        |
| `schema`          | Print the JSON schema of the data exported by this agent |
| `version`         | Print the version of the agent                           |

