---
title: CLI 
summary: What the agent can do
---


Here is what the agent can do in addition to collecting data.

| Command           | Description                                      |
| ----------------- | ------------------------------------------------ |
| `run`             | Run the agent                                    |
| `explore`         | Run the experimental terminal UI                 |
| `refresh-id`      | Regenerate the internal ID of the agent          |
| `defaults`, `def` | Print the default config                         |
| `id`              | Print the identifier of the agent                |
| `update`          | Update the agent                                 |
| `version`         | Print the version of the agent                   |
| `task`, `cron`    | Install a scheduled task                         |
| `help`, `h`       | Shows a list of commands or help for one command |

## Agent identifier

Every agent binary can be identified through a **16 bytes id** (`fc097e65503cb3ad9eb8e10f5a617611` by default).

!!! info ""
    Currently you can't see this id in the database. In the future, it may be present in an attribute like `updated_by`.

You can display the current id through the eponym command.

/// tab | Linux

```bash
situation id
```

///

/// tab | Windows

```ps1
situation.exe id
```

///

In different scenarios you may need to customize this id (naming, multi-deployment...). For these purpioses, you can generate a new random ID (or provide a new one in hex format):

/// tab | Linux

```bash
situation refresh-id 
```

///

/// tab | Windows

```ps1
situation.exe refresh-id
```

///

## Run configuration

By design, you can run the agent as-is but it is also possible to tune modules.

### Module configuration

Some modules may expose specific option through flags. Do not hesitate to look at them in the help. For example:

/// tab | Linux

```bash
situation run --ping-timeout=1s
```

///

/// tab | Windows

```ps1
situation.exe run --ping-timeout=1s
```

///

### Disabling modules

All the module can be disabled through the following pattern `--no-module-<module-name>` (see the list of [available modules](modules/index.md))

!!! note
    As some modules may depend on others, disabling a module may lead to a cascasding effect. To force modules that depend on it to run, you must pass the `--ignore-missing-deps` flag.

        :::shell
        situation run --no-module-ping --ignore-missing-deps
    

