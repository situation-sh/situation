---
linux: true
windows: true
macos: unknown
root: false
title: JA4
summary: "Attempts JA4, JA4S and JA4X fingerprinting"
date: 2025-07-28
filename: ja4.go
std_imports:
  - crypto/sha256
  - crypto/tls
  - crypto/x509
  - crypto/x509/pkix
  - encoding/asn1
  - encoding/binary
  - encoding/hex
  - fmt
  - net
  - slices
  - strings
  - time
imports: []
---

{% if windows == true %}{{ windows_ok }}{% endif %}
{% if linux == true %}{{ linux_ok }}{% endif %}
{% if root == true %}{{ root_required }}{% endif %}

JA4Module attempts JA4, JA4S and JA4X fingerprinting

### Details


For technical details you look at [https://github.com/FoxIO-LLC/ja4/blob/main/technical_details/README.md](https://github.com/FoxIO-LLC/ja4/blob/main/technical_details/README.md) It first look at TLS endpoints (given by the [TLS module](./tls.md)) and then tries to connect to them, collecting then JA4, JA4S and JA4X fingerprints.

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
