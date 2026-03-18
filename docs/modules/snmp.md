---
linux: true
windows: true
macos: unknown
root: false
title: SNMP
summary: "Collects network interface data from neighbors via SNMP."
date: 2026-03-17
filename: snmp.go
std_imports:
  - context
  - errors
  - fmt
  - net
  - strconv
  - strings
  - sync
  - time
imports:
  - github.com/asiffer/puzzle
  - github.com/gosnmp/gosnmp
  - github.com/sirupsen/logrus
options:
  - name: version
    type: uint8
    default: uint8(gosnmp.Version2c)
  - name: community
    type: string
    default: "public"
  - name: timeout
    type: time.Duration
    default: 3 * time.Second
  - name: transport
    type: string
    default: "udp"
  - name: port
    type: uint16
    default: 161

---

{% if windows == true %}{{ windows_ok }}{% endif %}
{% if linux == true %}{{ linux_ok }}{% endif %}
{% if root == true %}{{ root_required }}{% endif %}

SNMPModule collects network interface data from neighbors via SNMP.

### Details


This module requires access to the OID tree `.1.3.6.1.2.1`. In case of snmpd, the configuration (snmpd.conf) should include:

``` view systemonly included .1.3.6.1.2.1 ```

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
