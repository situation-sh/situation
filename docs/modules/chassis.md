---
linux: true
windows: false
macos: unknown
root: unknown
title: Chassis
summary: "ChassisModule fills host chassis information"
date: 2025-03-05
filename: chassis.go
std_imports:
  - os
imports:
  - github.com/godbus/dbus/v5
---

{% if windows == true %}{{ windows_ok }}{% endif %}
{% if linux == true %}{{ linux_ok }}{% endif %}
{% if root == true %}{{ root_required }}{% endif %}

ChassisModule fills host chassis information

### Details


Currently it only works under linux. It uses DBUS and the "org.freedesktop.hostname1" service to get the type of the chassis (like laptop, vm, desktop etc.) In the future it may rather rely on [ghw](https://github.com/jaypipes/ghw/) but at that time it does not fully get the info on windows.

### Dependencies

=== "Standard library"

	{% for i in std_imports %}
	 - [{{ i }}](https://pkg.go.dev/{{ i }})
	{% endfor %}

=== "External"

	{% for i in imports %}
	 - [{{ i }}](https://pkg.go.dev/{{ i }})
	{% endfor %}
