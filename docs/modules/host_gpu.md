---
linux: true
windows: true
macos: unknown
root: false
title: Host GPU
summary: "HostGPUModule retrieves basic information about GPU: index, vendor and product name."
date: 2024-01-25
filename: host_gpu.go
std_imports:
  - fmt
imports:
  - github.com/jaypipes/ghw
---

{% if windows == true %}{{ windows_ok }}{% endif %}
{% if linux == true %}{{ linux_ok }}{% endif %}
{% if root == true %}{{ root_required }}{% endif %}

HostGPUModule retrieves basic information about GPU: index, vendor and product name.

### Details


It heavily relies on [ghw](https://github.com/jaypipes/ghw). On Linux it reads `/sys/class/drm/` folder. On Windows, it performs the following WMI query:

 ```ps1
 SELECT Caption, CreationClassName, Description, DeviceID, Manufacturer, Name, PNPClass, PNPDeviceID FROM Win32_PnPEntity
 ```

### Dependencies

=== "Standard library"

	{% for i in std_imports %}
	 - [{{ i }}](https://pkg.go.dev/{{ i }})
	{% endfor %}

=== "External"

	{% for i in imports %}
	 - [{{ i }}](https://pkg.go.dev/{{ i }})
	{% endfor %}
