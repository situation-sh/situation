---
linux: true
windows: true
macos: unknown
root: false
title: Host Basic
summary: "HostBasicModule retrieves basic information about the host: hostid, architecture, platform, distribution, version and uptime"
date: 2025-03-05
filename: host_basic.go
std_imports:
  - os
  - time
imports:
  - github.com/google/uuid
  - github.com/shirou/gopsutil/v4/host
---

{% if windows == true %}{{ windows_ok }}{% endif %}
{% if linux == true %}{{ linux_ok }}{% endif %}
{% if root == true %}{{ root_required }}{% endif %}

HostBasicModule retrieves basic information about the host: hostid, architecture, platform, distribution, version and uptime

### Details


It heavily relies on the [gopsutil](https://github.com/shirou/gopsutil/) library.

 | Data                 | Linux                           | Windows                    |
 |----------------------|---------------------------------|----------------------------|
 | hostname             | `uname` syscall                 | `GetComputerNameExW` call  |
 | arch                 | `uname` syscall                 | `GetNativeSystemInfo` call |
 | platform             | `runtime.GOOS` variable         | `runtime.GOOS` variable    |
 | distribution         | scanning `/etc/*-release` files | `HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Windows NT\CurrentVersion*` register keys |
 | distribution version | scanning `/etc/*-release` files | `HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Windows NT\CurrentVersion*` register keys |
 | hostid               | reading `/sys/class/dmi/id/product_uuid`, `/etc/machine-id` or `/proc/sys/kernel/random/boot_id` | `HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Cryptography\MachineGuid` register key |
 | uptime               | `sysinfo` syscall               | `GetTickCount64` call      |

### Dependencies

=== "Standard library"

	{% for i in std_imports %}
	 - [{{ i }}](https://pkg.go.dev/{{ i }})
	{% endfor %}

=== "External"

	{% for i in imports %}
	 - [{{ i }}](https://pkg.go.dev/{{ i }})
	{% endfor %}
