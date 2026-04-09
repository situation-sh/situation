---
linux: true
windows: true
macos: unknown
root: false
title: StandardProtocol
summary: "Fills standard protocol information for endpoints."
date: 2026-04-09
filename: standard_protocol.go
std_imports:
  - context
  - fmt
  - strings
imports:
  - github.com/uptrace/bun
  - github.com/uptrace/bun/dialect

---

{% if windows == true %}{{ windows_ok }}{% endif %}
{% if linux == true %}{{ linux_ok }}{% endif %}
{% if root == true %}{{ root_required }}{% endif %}

StandardProtocolModule fills standard protocol information for endpoints.

### Details


As examples, it fills "http" for TCP port 80, "dns" for UDP port 53, etc.

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
