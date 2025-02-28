---
linux: true
windows: true
macos: unknown
root: unknown
title: Reverse Lookup
summary: "ReverseLookupModule tries to get a hostname attached to a local IP address"
date: 2025-02-28
filename: reverse_lookup.go
std_imports:
  - net
  - strings
imports: []
---

{% if windows == true %}{{ windows_ok }}{% endif %}
{% if linux == true %}{{ linux_ok }}{% endif %}
{% if root == true %}{{ root_required }}{% endif %}

ReverseLookupModule tries to get a hostname attached to a local IP address

### Details


### Dependencies

=== "Standard library"

	{% for i in std_imports %}
	 - [{{ i }}](https://pkg.go.dev/{{ i }})
	{% endfor %}

=== "External"

	{% for i in imports %}
	 - [{{ i }}](https://pkg.go.dev/{{ i }})
	{% endfor %}
