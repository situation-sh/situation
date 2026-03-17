---
linux: true
windows: true
macos: unknown
root: false
title: TCP Scan
summary: "Tries to connect to neighbor TCP ports."
date: 2026-03-17
filename: tcp_scan.go
std_imports:
  - context
  - fmt
  - net
  - strings
  - time
imports:
  - github.com/asiffer/puzzle
options:
  - name: timeout
    type: time.Duration
    default: 200 * time.Millisecond

---

{% if windows == true %}{{ windows_ok }}{% endif %}
{% if linux == true %}{{ linux_ok }}{% endif %}
{% if root == true %}{{ root_required }}{% endif %}

TCPScanModule tries to connect to neighbor TCP ports.

### Details


The module only uses the Go standard library.

A TCP connect is performed on the [NMAP top 1000 ports](https://nullsec.us/top-1-000-tcp-and-udp-ports-nmap-default/). These connection attempts are made concurrently against the hosts previously found. The connections have a 500ms timeout.

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
