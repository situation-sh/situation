---
linux: true
windows: true
macos: unknown
root: unknown
title: App User
summary: "Fills user information from the PID of an application"
date: 2025-09-24
filename: appuser.go
std_imports:
  - bufio
  - errors
  - fmt
  - os
  - os/user
  - strconv
  - strings
  - syscall
  - unsafe
imports: []
---

{% if windows == true %}{{ windows_ok }}{% endif %}
{% if linux == true %}{{ linux_ok }}{% endif %}
{% if root == true %}{{ root_required }}{% endif %}

AppUserModule fills user information from the PID of an application

### Details


On Linux, it uses the /proc/\<PID>/status entrypoint. On Windows, it calls `OpenProcessToken`, `GetTokenInformation` and `LookupAccountSidW`.

On windows, even if the agent is run as administrator, it may not have the required privileges to scan some processes like wininit.exe, services.exe.

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
