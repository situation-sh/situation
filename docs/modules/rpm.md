---
linux: true
windows: false
macos: false
root: false
title: RPM
summary: "RPMModule reads package information from the rpm package manager."
date: 2025-03-05
filename: rpm.go
std_imports:
  - bytes
  - database/sql
  - encoding/binary
  - fmt
  - io/fs
  - path
  - path/filepath
  - time
  - unicode/utf8
imports:
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

=== "Standard library"

	{% for i in std_imports %}
	 - [{{ i }}](https://pkg.go.dev/{{ i }})
	{% endfor %}

=== "External"

	{% for i in imports %}
	 - [{{ i }}](https://pkg.go.dev/{{ i }})
	{% endfor %}
