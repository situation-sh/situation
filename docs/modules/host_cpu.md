---
linux: true
windows: true
macos: unknown
root: false
title: Host CPU
summary: "HostCPUModule retrieves host CPU info: model, vendor and the number of cores."
date: 2023-07-20
filename: host_cpu.go
std_imports:
  - fmt
  - strconv
imports:
  - github.com/shirou/gopsutil/v3/cpu
---

{% if windows == true %}{{ windows_ok }}{% endif %}
{% if linux == true %}{{ linux_ok }}{% endif %}
{% if root == true %}{{ root_required }}{% endif %}

HostCPUModule retrieves host CPU info: model, vendor and the number of cores.

### Details


It heavily relies on the [gopsutil](https://github.com/shirou/gopsutil/) library.

On Linux, it reads `/proc/cpuinfo`. On Windows it performs the `win32_Processor` WMI request

### Dependencies

=== "Standard library"

	{% for i in std_imports %}
	 - [{{ i }}](https://pkg.go.dev/{{ i }})
	{% endfor %}

=== "External"

	{% for i in imports %}
	 - [{{ i }}](https://pkg.go.dev/{{ i }})
	{% endfor %}
