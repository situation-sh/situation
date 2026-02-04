---
linux: true
windows: true
macos: unknown
root: true
title: Netstat
summary: "Retrieves active connections."
date: 2026-02-02
filename: netstat.go
std_imports:
  - context
  - fmt
  - os
  - os/user
  - runtime
imports:
  - github.com/cakturk/go-netstat/netstat
---

{% if windows == true %}{{ windows_ok }}{% endif %}
{% if linux == true %}{{ linux_ok }}{% endif %}
{% if root == true %}{{ root_required }}{% endif %}

NetstatModule retrieves active connections.

### Details


It enumerates TCP, UDP, TCP6 and UDP6 sockets to discover listening endpoints, running applications (with PID and command line), and network flows between them. It must be run as root on Linux to retrieve PID/process information; without these data it is hard to build reliable links between open ports and programs.

This module is then able to create flows between applications according to the tuple (src, srcport, dst, dstport).

On Windows, the privileges are not checked. So the module is always run.

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
