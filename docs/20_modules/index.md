---
title: Modules reference
summary: List of all collectors
sidebar_title: Reference
---


<div style="display: flex; flex-direction: row; gap: 1.5rem; font-family: monospace; align-items: center;">
	<div style="display: flex; flex-direction: row; gap: 0.5rem; align-items: center;">
		<img src="{{ linux_icon_src }}" alt="linux" />
		<span>Linux</span>
	</div>
	<div style="display: flex; flex-direction: row; gap: 0.5rem; align-items: center;">
		<img src="{{ windows_icon_src }}" alt="windows" />
		<span>Windows</span>
	</div>
	<div style="display: flex; flex-direction: row; gap: 0.5rem; align-items: center;">
		<img src="{{ root_required_icon_src }}" alt="root-required" />
		<span>Root required</span>
	</div>
</div>
| Name | Summary | Dependencies | Status |
|------|---------|--------------|--------|
| [appuser](appuser.md)   | AppUserModule fills user information from the PID of an application      | [netstat](netstat.md)           | {{ linux_ok }} {{ windows_ok }}     |
| [arp](arp.md)   | ARPModule reads internal ARP table to find network neighbors.      | [ping](ping.md)           | {{ linux_ok }} {{ windows_ok }}     |
| [chassis](chassis.md)   | ChassisModule fills host chassis information      | [host-basic](host_basic.md)           | {{ linux_ok }}     |
| [docker](docker.md)   | DockerModule retrieves information about docker containers.      | [host-network](host_network.md), [tcp-scan](tcp_scan.md)           | {{ linux_ok }} {{ windows_ok }} {{ root_required }}     |
| [dpkg](dpkg.md)   | DPKGModule reads package information from the dpkg package manager.      | [host-basic](host_basic.md), [netstat](netstat.md)           | {{ linux_ok }}     |
| [host-basic](host_basic.md)   | HostBasicModule retrieves basic information about the host: hostid, architecture, platform, distribution, version and uptime      |            | {{ linux_ok }} {{ windows_ok }}     |
| [host-cpu](host_cpu.md)   | HostCPUModule retrieves host CPU info: model, vendor and the number of cores.      | [host-basic](host_basic.md)           | {{ linux_ok }} {{ windows_ok }}     |
| [host-disk](host_disk.md)   | HostDiskModule retrieves basic information about disk: name, model, size, type, controller and partitions.      | [host-basic](host_basic.md)           | {{ linux_ok }} {{ windows_ok }}     |
| [host-gpu](host_gpu.md)   | HostGPUModule retrieves basic information about GPU: index, vendor and product name.      | [host-basic](host_basic.md)           | {{ linux_ok }} {{ windows_ok }}     |
| [host-network](host_network.md)   | HostNetworkModule retrieves basic newtork information about the host: interfaces along with their mac, ip and mask (IPv4 and IPv6)      | [host-basic](host_basic.md)           | {{ linux_ok }} {{ windows_ok }}     |
| [ja4](ja4.md)   | JA4Module attempts JA4, JA4S and JA4X fingerprinting      | [tls](tls.md)           | {{ linux_ok }} {{ windows_ok }}     |
| [msi](msi.md)   | MSIModule creates models.Packages instance from the windows registry      | [host-basic](host_basic.md)           | {{ windows_ok }} {{ root_required }}     |
| [netstat](netstat.md)   | NetstatModule aims to retrieve infos like the netstat command does It must be run as root to retrieve PID/process information.      | [host-basic](host_basic.md), [host-network](host_network.md)           | {{ linux_ok }} {{ windows_ok }} {{ root_required }}     |
| [ping](ping.md)   | PingModule pings local networks to discover new hosts.      | [host-network](host_network.md)           | {{ linux_ok }} {{ windows_ok }}     |
| [reverse-lookup](reverse_lookup.md)   | ReverseLookupModule tries to get a hostname attached to a local IP address      | [netstat](netstat.md)           | {{ linux_ok }} {{ windows_ok }}     |
| [rpm](rpm.md)   | RPMModule reads package information from the rpm package manager.      | [host-basic](host_basic.md), [netstat](netstat.md)           | {{ linux_ok }}     |
| [snmp](snmp.md)   | SNMPModule Module to collect data through SNMP protocol.      | [arp](arp.md)           | {{ linux_ok }} {{ windows_ok }}     |
| [ssh](ssh.md)   | SSHModule aims to retrieve info from remote ssh services.      | [tcp-scan](tcp_scan.md)           | {{ linux_ok }} {{ root_required }}     |
| [tcp-scan](tcp_scan.md)   | TCPScanModule tries to connect to neighbor TCP ports.      | [arp](arp.md)           | {{ linux_ok }} {{ windows_ok }}     |
| [tls](tls.md)   | TLSModule enrich endpoints with TLS information.      | [tcp-scan](tcp_scan.md)           | {{ linux_ok }} {{ windows_ok }}     |
| [vmware](vmware.md)   | VMwareModule try to connect to esxi/vcenter hosts and list VMs      | [tls](tls.md)           | {{ linux_ok }} {{ windows_ok }}     |
| [zypper](zypper.md)   | ZypperModule reads package information from the zypper package manager.      | [host-basic](host_basic.md), [netstat](netstat.md)           | {{ linux_ok }}     |