---
linux: 
windows: 
macos: 
root: 
title: SaaS
summary: "Identifies SaaS applications from discovered endpoints."
date: 2026-02-18
filename: saas.go
std_imports:
  - context
  - errors
  - fmt
  - net
  - strings
imports:
  - github.com/asiffer/puzzle
---

{% if windows == true %}{{ windows_ok }}{% endif %}
{% if linux == true %}{{ linux_ok }}{% endif %}
{% if root == true %}{{ root_required }}{% endif %}

SaaSModule identifies SaaS applications from discovered endpoints.

### Details


For each endpoint that has not been classified yet, the module runs a set of pluggable detectors. Each detector matches endpoints by TLS DNS name suffixes or IP address ranges to identify known SaaS providers (e.g. GitHub, Outlook, Teams, SharePoint, Datadog, Sentry, Elastic, Anthropic).

The first matching detector wins and the SaaS name is stored on the endpoint.

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
