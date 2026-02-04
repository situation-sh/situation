---
linux: true
windows: true
macos: unknown
root: false
title: HostCPU
summary: "Retrieves host CPU info: model, vendor and the number of cores."
date: 2026-02-02
filename: host_cpu.go
std_imports:
  - context
  - fmt
  - strconv
imports:
  - github.com/shirou/gopsutil/v4/cpu
---

{% if windows == true %}{{ windows_ok }}{% endif %}
{% if linux == true %}{{ linux_ok }}{% endif %}
{% if root == true %}{{ root_required }}{% endif %}

HostCPUModule retrieves host CPU info: model, vendor and the number of cores.

### Details


It heavily relies on the [gopsutil](https://github.com/shirou/gopsutil/) library.

On Linux, it reads `/proc/cpuinfo`. On Windows it performs the `win32_Processor` WMI request

On windows 11, the local user account must have administrator permissions (it does not mean it must be run as root).

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
