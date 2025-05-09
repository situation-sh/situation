---
linux: false
windows: true
macos: unknown
root: true
title: MSI
summary: "Creates models.Packages instance from the windows registry"
date: 2025-05-09
filename: msi.go
std_imports:
  - fmt
  - io/fs
  - os
  - path/filepath
  - strings
  - sync
  - time
imports:
  - github.com/sirupsen/logrus
  - golang.org/x/sys/windows/registry
---

{% if windows == true %}{{ windows_ok }}{% endif %}
{% if linux == true %}{{ linux_ok }}{% endif %}
{% if root == true %}{{ root_required }}{% endif %}

MSIModule creates models.Packages instance from the windows registry

### Details


For system-wide apps, it looks at `HKLM/SOFTWARE/Microsoft/Windows/CurrentVersion/Uninstall/*` and `HKLM/WOW6432Node/SOFTWARE/Microsoft/Windows/CurrentVersion/Uninstall/*` for 32bits apps. For user-specific apps: `HKCU/SOFTWARE/Microsoft/Windows/CurrentVersion/Uninstall/*`.

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
