---
linux: true
windows: true
macos: unknown
root: false
title: TLS
summary: "Enriches TCP endpoints with TLS certificate information."
date: 2026-02-05
filename: tls.go
std_imports:
  - context
  - crypto/sha1
  - crypto/sha256
  - crypto/tls
  - encoding/hex
  - fmt
  - net
imports:
  - github.com/sirupsen/logrus
  - github.com/uptrace/bun
---

{% if windows == true %}{{ windows_ok }}{% endif %}
{% if linux == true %}{{ linux_ok }}{% endif %}
{% if root == true %}{{ root_required }}{% endif %}

TLSModule enriches TCP endpoints with TLS certificate information.

### Details


It connects to endpoints on well-known TLS ports (HTTPS, IMAPS, LDAPS, etc.) and performs a TLS handshake to extract the leaf certificate. For each certificate it collects: subject, issuer, validity period, serial number, signature and public key algorithms, SHA-1/SHA-256 fingerprints, and DNS names.

The module only uses the Go standard library. Currently it only supports TLS over TCP.

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
