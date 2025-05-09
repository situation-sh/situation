---
title: Configuration
summary: Going beyond the defaults
order: 30
---

What??? You said **zero-config**!

Yes by default, there is no configuration. However one may tweak a little bit the agent so as to fulfil some specific needs.

## Scans

You can configure the number of scans to perform and the waiting time between two scans using the following options:

- `--scans value`: Number of scans to perform (default: 1)
- `--period value`: Waiting time between two scans (default: 1m0s)

## Disabling modules

All the module can be disabled through the following pattern `--no-module-<module-name>` (see the list of [available modules](modules/index.md))

!!! note
    As some modules may depend on others, disabling a module may lead to a cascasding effect. To force modules that depend on it to run, you must pass the `--skip-missing-deps` flag.

        :::shell
        situation --stdout --no-module-ping --skip-missing-deps


## Module configuration

Some modules may expose specific option through flags. Do not hesitate to look at them in the help. For example:

```shell
situation --stdout --snmp-community=local
```