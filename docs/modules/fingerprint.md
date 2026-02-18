---
linux: true
windows: true
macos: unknown
root: false
title: Fingerprint
summary: "Attempts to match the local host against machines already discovered in the shared database."
date: 2026-02-18
filename: fingerprint.go
std_imports:
  - context
  - fmt
  - net
  - strings
imports:
  - github.com/shirou/gopsutil/v4/host
---

{% if windows == true %}{{ windows_ok }}{% endif %}
{% if linux == true %}{{ linux_ok }}{% endif %}
{% if root == true %}{{ root_required }}{% endif %}

FingerprintModule attempts to match the local host against machines already discovered in the shared database.

### Details


This module is critical for multi-agent deployments where Agent A may have discovered Host B remotely (via ARP, ping, TCP scan), and later Agent B starts on Host B. Without fingerprinting, Agent B would create a duplicate machine entry instead of recognizing itself.

Matching strategy:

 1. Agent UUID match → definitive (reconnection case)
 2. HostID (system UUID) match → definitive
 3. Fuzzy matching on MAC/IP/hostname with weighted scores

The module runs before any other module (no dependencies) to ensure the host machine is correctly identified before other modules populate it.

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
