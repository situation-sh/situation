---
linux: true
windows: true
macos: unknown
root: false
title: StandardProtocol
summary: "Fills standard protocol information for endpoints."
date: 2026-02-13
filename: standard_protocol.go
std_imports:
  - context
  - fmt
  - strings
imports:
  - github.com/uptrace/bun
---

{% if windows == true %}{{ windows_ok }}{% endif %}
{% if linux == true %}{{ linux_ok }}{% endif %}
{% if root == true %}{{ root_required }}{% endif %}

StandardProtocolModule fills standard protocol information for endpoints.

### Details


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
