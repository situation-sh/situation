# Configuration

What??? You said **zero-config**!

Yes by default, there is no configuration. However one may tweak a little bit the agent so as to fulfil some specific needs.

## Scans

You can configure the number of scans to perform and the waiting time between two scans using the following options:

- `--scans value`, `-s value`: Number of scans to perform (default: 1)
- `--period value`, `-p value`: Waiting time between two scans (default: 2m0s)

## Disabling modules

All the module can be disabled through the following pattern `--modules.<MODULE-NAME>.disabled=1` (see the list of [available modules](modules/index.md))

!!! warning 
    As some modules may depend on others, disabling a module may lead to a cascasde effect.