---
linux: true
windows: true
macos: unknown
root: false
title: Host Network
summary: "Retrieves basic newtork information about the host: interfaces along with their mac, ip and mask (IPv4 and IPv6)"
date: 2025-07-24
filename: host_network.go
std_imports:
  - fmt
  - net
  - strings
imports:
  - github.com/libp2p/go-netroute
---

{% if windows == true %}{{ windows_ok }}{% endif %}
{% if linux == true %}{{ linux_ok }}{% endif %}
{% if root == true %}{{ root_required }}{% endif %}

HostNetworkModule retrieves basic newtork information about the host: interfaces along with their mac, ip and mask (IPv4 and IPv6)

### Details


It uses the [go](https://pkg.go.dev/net) standard library.

On Linux, it uses the Netlink API. On Windows, it calls `GetAdaptersAddresses`.

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
