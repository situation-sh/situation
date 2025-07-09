---
linux: true
windows: unknown
macos: 
root: true
title: SSH
summary: "Aims to retrieve info from remote ssh services."
date: 2025-05-09
filename: ssh.go
std_imports:
  - encoding/json
  - fmt
  - net
  - regexp
  - strings
  - time
imports:
  - github.com/praetorian-inc/fingerprintx/pkg/plugins
  - github.com/praetorian-inc/fingerprintx/pkg/plugins/services/ssh
---

{% if windows == true %}{{ windows_ok }}{% endif %}
{% if linux == true %}{{ linux_ok }}{% endif %}
{% if root == true %}{{ root_required }}{% endif %}

SSHModule aims to retrieve info from remote ssh services.

### Details


It mainly tries to connect to open tcp/22 ports, gathering everything it can like the `host_key` and the algorithms available. In the OpenSSH case it also tries to parse the banner to get product and OS infos (versions notably)

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
