| Name | Summary | Dependencies | Status |
|------|---------|--------------|--------|
| [host-cpu](host_cpu.md)   | HostCPUModule retrieves host CPU info: model, vendor and the number of cores.      | [host-basic](host_basic.md)           | {{ linux_ok }} {{ windows_ok }}     |
| [host-disk](host_disk.md)   | HostDiskModule retrieves basic information about disk: name, model, size, type, controller and partitions.      | [host-basic](host_basic.md)           | {{ linux_ok }} {{ windows_ok }}     |
| [host-gpu](host_gpu.md)   | HostGPUModule retrieves basic information about GPU: index, vendor and product name.      | [host-basic](host_basic.md)           | {{ linux_ok }} {{ windows_ok }}     |
| [host-network](host_network.md)   | HostNetworkModule retrieves basic newtork information about the host: interfaces along with their mac, ip and mask (IPv4 and IPv6)      | [host-basic](host_basic.md)           | {{ linux_ok }} {{ windows_ok }}     |
| [netstat](netstat.md)   | NetstatModule aims to retrieve infos like the netstat command does It must be run as root to retrieve PID/process information.      | [host-basic](host_basic.md)           | {{ linux_ok }} {{ windows_ok }} {{ root_required }}     |
| [ping](ping.md)   | PingModule pings local networks to discover new hosts.      | [host-network](host_network.md)           | {{ linux_ok }} {{ windows_ok }}     |
| [docker](docker.md)   | DockerModule retrieves information about docker containers.      | [host-network](host_network.md), [tcp-scan](tcp_scan.md)           | {{ linux_ok }} {{ windows_ok }} {{ root_required }}     |
| [host-basic](host_basic.md)   | HostBasicModule retrieves basic information about the host: hostid, architecture, platform, distribution, version and uptime      |            | {{ linux_ok }} {{ windows_ok }}     |
| [rpm](rpm.md)   | RPMModule reads package information from the rpm package manager.      | [host-basic](host_basic.md), [netstat](netstat.md)           | {{ linux_ok }}     |
| [tcp-scan](tcp_scan.md)   | TCPScanModule tries to connect to neighbor TCP ports.      | [arp](arp.md)           | {{ linux_ok }} {{ windows_ok }}     |
| [arp](arp.md)   | ARPModule reads internal ARP table to find network neighbors.      | [ping](ping.md)           | {{ linux_ok }} {{ windows_ok }}     |
| [dpkg](dpkg.md)   | DPKGModule reads package information from the dpkg package manager.      | [host-basic](host_basic.md), [netstat](netstat.md)           | {{ linux_ok }}     |