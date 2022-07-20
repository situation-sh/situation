# Situation

The autonomous data collector.

Situation is a project that aims to **discover** everything on information systems, on its own. In a way, it lies between [nmap](https://nmap.org/), [telegraf](https://www.influxdata.com/time-series-platform/telegraf/) and [osquery](https://osquery.io/). However it mainly differs from them on the following aspect: **user do not declare what to collect or where**.

When we run tools like `nmap` or `telegraf`, we know the targets (ex: a subnetwork, a specific service...) and we must configure the tool in this way. `situation` aims to run without prior knowledge and this philosophy has two advantages:

- frictionless deployment (single binary, just download and run)
- no blind spots (who knows exactly what runs on his/her system?)

Situation is bound to collect data, nothing more. To go further, `situation` provides a [json schema](https://{{ variables.go_module }}/releases/download/v{{ variables.version }}/schema.json) for the output data.


!!! tip "Status"
    Situation is a early stage project. It currently targets Linux and Windows but keep in mind that it has not been extensively tested. It does not mean that is a dangerous codebase, only that it may fail.
