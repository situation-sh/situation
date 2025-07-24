---
linux: true
windows: true
macos: unknown
root: false
title: SNMP
summary: "Module to collect data through SNMP protocol."
date: 2025-07-24
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
  - github.com/gosnmp/gosnmp
  - github.com/sirupsen/logrus
---

{% if windows == true %}{{ windows_ok }}{% endif %}
{% if linux == true %}{{ linux_ok }}{% endif %}
{% if root == true %}{{ root_required }}{% endif %}

SNMPModule Module to collect data through SNMP protocol.

### Details


This module need to access the following OID TREE: `.1.3.6.1.2.1` In case of snmpd, the configuration (snmpd.conf) should then include something like this:

 ```conf
 view systemonly included .1.3.6.1.2.1
 ```

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
