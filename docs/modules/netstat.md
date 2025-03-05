---
linux: true
windows: true
macos: unknown
root: true
title: Netstat
summary: "NetstatModule aims to retrieve infos like the netstat command does It must be run as root to retrieve PID/process information."
date: 2025-03-05
filename: netstat.go
std_imports:
  - os/user
  - runtime
imports:
  - github.com/cakturk/go-netstat/netstat
---

{% if windows == true %}{{ windows_ok }}{% endif %}
{% if linux == true %}{{ linux_ok }}{% endif %}
{% if root == true %}{{ root_required }}{% endif %}

NetstatModule aims to retrieve infos like the netstat command does It must be run as root to retrieve PID/process information.

### Details
 Without these data, it is rather hard to build reliable links between open ports and programs.

This module is then able to create flows between applications according to the tuple (src, srcport, dst, dstport).

On windows, the privileges are not checked (because we need to parse the SID or another thing maybe). So the module is always run.

### Dependencies

=== "Standard library"

	{% for i in std_imports %}
	 - [{{ i }}](https://pkg.go.dev/{{ i }})
	{% endfor %}

=== "External"

	{% for i in imports %}
	 - [{{ i }}](https://pkg.go.dev/{{ i }})
	{% endfor %}
