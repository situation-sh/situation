---
linux: true
windows: true
macos: unknown
root: false
title: VMware
summary: "Tries to connect to esxi/vcenter hosts and list VMs"
date: 2025-09-24
filename: vmware.go
std_imports:
  - context
  - fmt
  - net
  - net/url
  - regexp
  - strings
  - time
imports:
  - github.com/sirupsen/logrus
  - github.com/vmware/govmomi
  - github.com/vmware/govmomi/find
  - github.com/vmware/govmomi/vim25/mo
  - github.com/vmware/govmomi/vim25/types
---

{% if windows == true %}{{ windows_ok }}{% endif %}
{% if linux == true %}{{ linux_ok }}{% endif %}
{% if root == true %}{{ root_required }}{% endif %}

VMwareModule tries to connect to esxi/vcenter hosts and list VMs

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
