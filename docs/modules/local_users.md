---
linux: true
windows: true
macos: false
root: false
title: LocalUsers
summary: "Lists all local user accounts on the system."
date: 2026-02-13
filename: local_users.go
std_imports:
  - bufio
  - context
  - fmt
  - os
  - os/user
  - strings
  - syscall
  - unsafe
imports:
  - golang.org/x/sys/windows
---

{% if windows == true %}{{ windows_ok }}{% endif %}
{% if linux == true %}{{ linux_ok }}{% endif %}
{% if root == true %}{{ root_required }}{% endif %}

LocalUsersModule lists all local user accounts on the system.

### Details


On **Linux**, the module reads `/etc/passwd` to enumerate user entries. Each UID is then resolved through the standard `os/user` library to retrieve the full user details.

On **Windows**, the module calls the Win32 `NetUserEnum` API (from `netapi32.dll`) to enumerate local accounts filtered to normal user accounts. Each username is then resolved with `os/user.Lookup`, and the user's domain is determined by converting the SID via `LookupAccountSid`.

The collected users are stored in the database with an upsert strategy based on `(machine_id, uid)`.

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
