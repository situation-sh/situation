---
linux: true
windows: true
macos: unknown
root: false
title: Ping
summary: "PingModule pings local networks to discover new hosts."
date: 2024-06-13
filename: ping.go
std_imports:
  - fmt
  - net
  - os/user
  - runtime
  - strings
  - sync
  - time
imports:
  - github.com/prometheus-community/pro-bing
  - github.com/sirupsen/logrus
---

{% if windows == true %}{{ windows_ok }}{% endif %}
{% if linux == true %}{{ linux_ok }}{% endif %}
{% if root == true %}{{ root_required }}{% endif %}

PingModule pings local networks to discover new hosts.

### Details


The module relies on [pro-bing](https://github.com/prometheus-community/pro-bing) library.

A single ping attempt is made on every host of the local networks (the host may belong to several networks). Only IPv4 networks with prefix length >=20 are treated. The ping timeout is hardset to 300ms.

### Dependencies

=== "Standard library"

	{% for i in std_imports %}
	 - [{{ i }}](https://pkg.go.dev/{{ i }})
	{% endfor %}

=== "External"

	{% for i in imports %}
	 - [{{ i }}](https://pkg.go.dev/{{ i }})
	{% endfor %}
