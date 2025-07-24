---
linux: true
windows: true
macos: unknown
root: false
title: Host Disk
summary: "Retrieves basic information about disk: name, model, size, type, controller and partitions."
date: 2025-07-24
filename: host_disk.go
std_imports:
  - fmt
imports:
  - github.com/jaypipes/ghw
---

{% if windows == true %}{{ windows_ok }}{% endif %}
{% if linux == true %}{{ linux_ok }}{% endif %}
{% if root == true %}{{ root_required }}{% endif %}

HostDiskModule retrieves basic information about disk: name, model, size, type, controller and partitions.

### Details


It heavily relies on the [ghw](https://github.com/jaypipes/ghw/) library.

On Windows, it uses WMI requests:

  ```ps1
  SELECT Caption, CreationClassName, Description, DeviceID, FileSystem, FreeSpace, Name, Size, SystemName FROM Win32_LogicalDisk
  ```

  ```ps1
  SELECT DeviceId, MediaType FROM MSFT_PhysicalDisk
  ```

  ```ps1
  SELECT Access, BlockSize, Caption, CreationClassName, Description, DeviceID, DiskIndex, Index, Name, Size, SystemName, Type FROM Win32_DiskPartition
  ```

  ```ps1
  SELECT Antecedent, Dependent FROM Win32_LogicalDiskToPartition
  ```

On Linux, it reads `/sys/block/$DEVICE/**` files. On windows 11, the local user account must have administrator permissions (it does not mean it must be run as root).

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
