---
linux: true
windows: false
macos: false
root: false
title: DPKG
summary: "Reads package information from the dpkg package manager."
date: 2026-02-13
filename: dpkg.go
std_imports:
  - bufio
  - context
  - errors
  - fmt
  - os
  - path/filepath
  - strings
  - time
imports: []
---

{% if windows == true %}{{ windows_ok }}{% endif %}
{% if linux == true %}{{ linux_ok }}{% endif %}
{% if root == true %}{{ root_required }}{% endif %}

DPKGModule reads package information from the dpkg package manager.

### Details


This module is relevant for distros that use dpkg, like debian, ubuntu and their derivatives. It only uses the standard library.

It reads `/var/log/dpkg.log` and also files from `/var/lib/dpkg/info/`.

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
