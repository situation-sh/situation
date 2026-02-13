---
linux: true
windows: true
macos: unknown
root: false
title: HostNetwork
summary: "Retrieves basic network information about the host."
date: 2026-02-13
filename: host_network.go
std_imports:
  - context
  - fmt
  - net
  - strings
imports:
  - github.com/libp2p/go-netroute
---

{% if windows == true %}{{ windows_ok }}{% endif %}
{% if linux == true %}{{ linux_ok }}{% endif %}
{% if root == true %}{{ root_required }}{% endif %}

HostNetworkModule retrieves basic network information about the host.

### Details


It uses the [net](https://pkg.go.dev/net) standard library to grab interfaces along with their name, MAC address, IP addresses (IPv4 and IPv6), subnet masks and [go-netroute](https://github.com/libp2p/go-netroute) for gateway detection.

On Linux, it uses the Netlink API. On Windows, it calls `GetAdaptersAddresses`.

Virtual interfaces (veth, qemu) are filtered out. The module also creates subnetwork records and links each network interface to its subnets.

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
