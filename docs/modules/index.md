# Modules

<div id="modules" markdown>
|Name|Dependencies|Architectures|Linux|Windows|System requirements|Required Go modules|
|---|---|---|---|---|---|---|
|[arp](arp.md)|[`Ping`](ping.md)|++question++|:white_check_mark:|:white_check_mark:||<ul><li>[`golang.org/x/sys/windows`]({{ variables.godoc_base_url }}/golang.org/x/sys/windows)</li><li>[`github.com/vishvananda/netlink`]({{ variables.godoc_base_url }}/github.com/vishvananda/netlink/)</li></ul>|
|[docker](docker.md)|[`Host network`](host_network.md) [`TCP Scan`](tcp_scan.md)|++question++|:white_check_mark:{ title="OK" }|:military_helmet:{ title="Beta" }|Must belong to `docker` group (Linux + unix socket)|<ul><li>[`github.com/docker/docker/client`]({{ variables.godoc_base_url }}/github.com/docker/docker/client)</li><li> [`github.com/docker/docker/api/types`]({{ variables.godoc_base_url }}/github.com/docker/docker/api/types)</li><li>[`github.com/docker/docker/api/types/filters`]({{ variables.godoc_base_url }}/github.com/docker/docker/api/types/filters)</li><li>[`github.com/docker/docker/api/types/network`]({{ variables.godoc_base_url }}/github.com/docker/docker/api/types/network)</li></ul>|
|[host-basic](host_basic.md)||++question++|:white_check_mark:{ title="OK" }|:white_check_mark:{ title="OK" }||<ul><li>[`github.com/shirou/gopsutil/v3/host`]({{ variables.godoc_base_url }}/github.com/shirou/gopsutil/v3/host)</li></ul>|
|[host-cpu](host_cpu.md)|[`Host basic`](host_basic.md)|++question++|:white_check_mark:{ title="OK" }|:white_check_mark:{ title="OK" }||<ul><li>[`github.com/shirou/gopsutil/v3/cpu`]({{ variables.godoc_base_url }}/github.com/shirou/gopsutil/v3/cpu)</li></ul>|
|[host-network](host_network.md)|[`Host basic`](host_basic.md)|++question++|:white_check_mark:{ title="OK" }|:white_check_mark:{ title="OK" }|||
|[netstat](netstat.md)|[`Host basic`](host_basic.md)|++question++|:white_check_mark:|:white_check_mark:|Need root privileges|<ul><li>[`github.com/cakturk/gonetstat/netstat`]({{ variables.godoc_base_url }}/github.com/cakturk/gonetstat/netstat)</li></ul>|
|[ping](ping.md)|[`Host network`](host_network.md)|++question++|:white_check_mark:{ title="OK" }|:white_check_mark:{ title="OK" }||<ul><li>[`github.com/goping/ping`]({{ variables.godoc_base_url }}/github.com/goping/ping)</li></ul>|
|[tcp-scan](tcp_scan.md)|[`ARP`](arp.md)|++question++|:white_check_mark:{ title="OK" }|:white_check_mark:{ title="OK" }|||
</div>