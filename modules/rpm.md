---
linux: true
windows: false
macos: false
root: false
title: RPM
summary: "Reads package information from the rpm package manager."
date: 2026-02-25
filename: rpm.go
std_imports:
  - bytes
  - context
  - database/sql
  - encoding/binary
  - fmt
  - io/fs
  - path
  - path/filepath
  - unicode/utf8
imports:
  - github.com/hashicorp/go-version
  - github.com/sirupsen/logrus
  - modernc.org/sqlite
---

{% if windows == true %}{{ windows_ok }}{% endif %}
{% if linux == true %}{{ linux_ok }}{% endif %}
{% if root == true %}{{ root_required }}{% endif %}

RPMModule reads package information from the rpm package manager.

### Details


This module is relevant for distros that use rpm, like fedora, redhat and their derivatives. It uses an sqlite client because of the way rpm works.

It tries to read the rpm database: `/var/lib/rpm/rpmdb.sqlite`. Otherwise, it will try to find the `rpmdb.sqlite` file inside `/usr/lib`.

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
