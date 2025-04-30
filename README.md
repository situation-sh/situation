# Situation

[![test](https://github.com/situation-sh/situation/actions/workflows/test.yaml/badge.svg)](https://github.com/situation-sh/situation/actions/workflows/test.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/situation-sh/situation)](https://goreportcard.com/report/github.com/situation-sh/situation)

The autonomous IT data collector.

Situation is a project that aims to **discover** everything on information systems, on its own. In a way, it lies between [nmap](https://nmap.org/), [telegraf](https://www.influxdata.com/time-series-platform/telegraf/) and [osquery](https://osquery.io/). However it mainly differs from them on the following aspect: **user do not declare what to collect or where**.

When we run tools like `nmap` or `telegraf`, we know the targets (ex: a subnetwork, a specific service...) and we must configure the tool in this way. `situation` aims to run without prior knowledge and this philosophy has two advantages:

- frictionless deployment (single binary, just download and run)
- no blind spots (who knows exactly what runs on his/her system?)

Situation is bound to collect data, nothing more. To go further, `situation` provides a [json schema](https://github.com/situation-sh/situation/releases/download/v0.14.0/schema.json) for the output data.

> [!WARNING]  
> Situation tries to abstract all the IT mess. It currently targets Linux and Windows but keep in mind that it has not been tested on all the machines on Earth. It does not mean that is a dangerous codebase, only that it may fail.

View the Situation [documentation](https://situation-sh.github.io/situation/).
