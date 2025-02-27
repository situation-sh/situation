---
linux: true
windows: true
macos: unknown
root: false
title: ARP
summary: "ARPModule reads internal ARP table to find network neighbors."
date: 2025-02-27
filename: arp.go
std_imports:
  - encoding/binary
  - fmt
  - net
  - syscall
  - time
  - unsafe
imports:
  - github.com/vishvananda/netlink
  - golang.org/x/sys/windows
---

{% if windows == true %}{{ windows_ok }}{% endif %}
{% if linux == true %}{{ linux_ok }}{% endif %}
{% if root == true %}{{ root_required }}{% endif %}

ARPModule reads internal ARP table to find network neighbors.

### Details
 It **does not send ARP requests** but leverage the [Ping] module that is likely to update the local table.

On Linux, it uses the Netlink API with the [netlink](https://github.com/vishvananda/netlink1) library. On Windows, it calls `GetIpNetTable2`.

[Ping]: ping.md

### Dependencies

=== "Standard library"

	{% for i in std_imports %}
	 - [{{ i }}](https://pkg.go.dev/{{ i }})
	{% endfor %}

=== "External"

	{% for i in imports %}
	 - [{{ i }}](https://pkg.go.dev/{{ i }})
	{% endfor %}
