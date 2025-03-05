---
linux: true
windows: false
macos: false
root: false
title: Zypper
summary: "ZypperModule reads package information from the zypper package manager."
date: 2025-03-05
filename: zypper.go
std_imports:
  - fmt
  - time
imports:
  - github.com/knqyf263/go-rpmdb/pkg
---

{% if windows == true %}{{ windows_ok }}{% endif %}
{% if linux == true %}{{ linux_ok }}{% endif %}
{% if root == true %}{{ root_required }}{% endif %}

ZypperModule reads package information from the zypper package manager.

### Details


This module is relevant for distros that use zypper, like suse and their derivatives. It uses [go-rpmdb](https://github.com/knqyf263/go-rpmdb/).

It reads `/var/lib/rpm/Packages.db`.

### Dependencies

=== "Standard library"

	{% for i in std_imports %}
	 - [{{ i }}](https://pkg.go.dev/{{ i }})
	{% endfor %}

=== "External"

	{% for i in imports %}
	 - [{{ i }}](https://pkg.go.dev/{{ i }})
	{% endfor %}
