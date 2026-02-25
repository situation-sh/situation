---
sidebar_title: Home 
title: Situation 
summary: IT data collection infrastructure
---

Situation provides the core infrastructure to automatically collect and consolidate IT data (machines, device, apps, network, flows...), on its own. 
Providing then an up-to-date and reliable view of the current state of your infra (or your home LAN), namely the *graph*.

Now you are ready to build a context-rich IT tool above Situation.

~{architecture}(architecture.json)

**Yet another scanning tool?**

Situation is different from common tools like [nmap](https://nmap.org/), [telegraf](https://www.influxdata.com/time-series-platform/telegraf/) or [osquery](https://osquery.io/):

1. It aims to run without prior knowledge
2. Agents collaborate natively
3. It builds the whole infra, namely the graph (not only the nodes)




