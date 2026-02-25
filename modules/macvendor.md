---
linux: 
windows: 
macos: 
root: 
title: MAC Vendor
summary: "Resolves manufacturer from MAC addresses."
date: 2026-02-25
filename: macvendor.go
std_imports:
  - context
  - fmt
imports: []
---

{% if windows == true %}{{ windows_ok }}{% endif %}
{% if linux == true %}{{ linux_ok }}{% endif %}
{% if root == true %}{{ root_required }}{% endif %}

MACVendorModule resolves manufacturer from MAC addresses.

### Details


It uses a built-in lookup table of IEEE OUI assignments (generated from the official IEEE OUI registry) to match the first 3 octets of each MAC address to a vendor name.

The module queries all network interfaces that have a MAC address but no vendor yet, and updates them in bulk.

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
