---
linux: true
windows: true
macos: unknown
root: false
title: SNMP
summary: "SNMPModule This module need to access the following OID TREE: .1.3.6.1.2.1 In case of snmpd, the conf (snmpd.conf) should include something like this: view systemonly included .1.3.6.1.2.1"
date: 2024-01-25
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

SNMPModule This module need to access the following OID TREE: .1.3.6.1.2.1 In case of snmpd, the conf (snmpd.conf) should include something like this: view systemonly included .1.3.6.1.2.1

### Details


### Dependencies

=== "Standard library"

	{% for i in std_imports %}
	 - [{{ i }}](https://pkg.go.dev/{{ i }})
	{% endfor %}

=== "External"

	{% for i in imports %}
	 - [{{ i }}](https://pkg.go.dev/{{ i }})
	{% endfor %}
