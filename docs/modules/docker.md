---
linux: true
windows: true
macos: unknown
root: true
title: Docker
summary: "Retrieves information about docker containers."
date: 2026-04-09
filename: docker.go
std_imports:
  - context
  - fmt
  - net
  - net/netip
  - runtime
  - strings
  - time
imports:
  - github.com/asiffer/puzzle
  - github.com/moby/moby/api/types/container
  - github.com/moby/moby/client
  - github.com/sirupsen/logrus
options:
  - name: host
    type: string
    default: defaultDockerHost()

---

{% if windows == true %}{{ windows_ok }}{% endif %}
{% if linux == true %}{{ linux_ok }}{% endif %}
{% if root == true %}{{ root_required }}{% endif %}

DockerModule retrieves information about docker containers.

### Details


It uses the official go client that performs HTTP queries either on port `:2376` (on windows generally) or on UNIX sockets.

We generally need some privileges to reads UNIX sockets, so it may require root privileges (the alternative is to belong to the `docker` group)

{% if options %}
### Options

| Name | Type | Default | Flag |
| ---- | ---- | ------- | ---- |{% for opt in options %}
| {{ opt.name }} | {{ opt.type|backticked }} | {{ opt.default }} | {{ ('--' ~ (title|lower) ~ '-' ~ opt.name)|backticked  }} |{% endfor %}

{% endif %}

### Dependencies

/// tab | Standard library

{% for i in std_imports %}
- [{{ i }}](https://pkg.go.dev/{{ i }})
{% endfor %}

///

/// tab | External

{% for i in imports %}
- [{{ i }}](https://pkg.go.dev/{{ i }})
{% endfor %}

///
