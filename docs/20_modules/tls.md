---
linux: true
windows: true
macos: unknown
root: false
title: TLS
summary: "Enrich endpoints with TLS information."
date: 2025-07-24
filename: tls.go
std_imports:
  - crypto/sha1
  - crypto/sha256
  - crypto/tls
  - encoding/hex
  - fmt
  - net
imports: []
---

{% if windows == true %}{{ windows_ok }}{% endif %}
{% if linux == true %}{{ linux_ok }}{% endif %}
{% if root == true %}{{ root_required }}{% endif %}

TLSModule enrich endpoints with TLS information.

### Details


The module only uses the Go standard library. Currently it only supports TLS over TCP.

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
