---
linux: true
windows: true
macos: unknown
root: unknown
title: ReverseLookup
summary: "Tries to get a hostname attached to a local IP address"
date: 2026-02-18
filename: reverse_lookup.go
std_imports:
  - context
  - fmt
  - net
  - strings
imports: []
---

{% if windows == true %}{{ windows_ok }}{% endif %}
{% if linux == true %}{{ linux_ok }}{% endif %}
{% if root == true %}{{ root_required }}{% endif %}

ReverseLookupModule tries to get a hostname attached to a local IP address

### Details


It basically calls [net.LookupAddr](https://pkg.go.dev/net#LookupAddr) that uses the host resolver to perform a reverse lookup for the given addresses.

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
