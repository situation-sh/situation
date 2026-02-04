---
linux: true
windows: true
macos: unknown
root: false
title: Ping
summary: "Pings local networks to discover new hosts."
date: 2026-02-02
filename: ping.go
std_imports:
  - context
  - encoding/binary
  - fmt
  - net
  - os
  - sync
  - syscall
  - time
  - unsafe
imports:
  - github.com/asiffer/puzzle
  - github.com/sirupsen/logrus
  - golang.org/x/net/icmp
  - golang.org/x/net/ipv4
  - golang.org/x/sys/windows
---

{% if windows == true %}{{ windows_ok }}{% endif %}
{% if linux == true %}{{ linux_ok }}{% endif %}
{% if root == true %}{{ root_required }}{% endif %}

PingModule pings local networks to discover new hosts.

### Details


The module relies on [pro-bing](https://github.com/prometheus-community/pro-bing) library.

A single ping attempt is made on every host of the local networks (the host may belong to several networks). Only IPv4 networks with prefix length >=20 are treated. The ping timeout is hardset to 300ms.

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
