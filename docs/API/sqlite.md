---
title: SQLite
summary: Database schema for SQLite storage
---

+mynaui:link-one+ Primary Key&emsp;+mynaui:key+ Foreign Key&emsp;+mynaui:one-diamond-solid+ +mynaui:two-diamond-solid+ Unique

## subnetworks


| Name | Type |  |
|------|------|-------------|
| `id` | `INTEGER` | +mynaui:link-one+ |
| `created_at` | `TIMESTAMP` |  |
| `updated_at` | `TIMESTAMP` |  |
| `network_cidr` | `VARCHAR` | +mynaui:one-diamond-solid+ |
| `network_addr` | `VARCHAR` |  |
| `mask_size` | `INTEGER` |  |
| `ip_version` | `INTEGER` |  |
| `gateway` | `VARCHAR` |  |
| `vlan_id` | `INTEGER` |  |
| `tag` | `VARCHAR` | +mynaui:one-diamond-solid+ |


## machines


| Name | Type |  |
|------|------|-------------|
| `id` | `INTEGER` | +mynaui:link-one+ |
| `created_at` | `TIMESTAMP` |  |
| `updated_at` | `TIMESTAMP` |  |
| `hostname` | `VARCHAR` |  |
| `host_id` | `VARCHAR` | +mynaui:one-diamond-solid+ |
| `arch` | `VARCHAR` |  |
| `platform` | `VARCHAR` |  |
| `distribution` | `VARCHAR` |  |
| `distribution_version` | `VARCHAR` |  |
| `distribution_family` | `VARCHAR` |  |
| `uptime` | `INTEGER` |  |
| `agent` | `VARCHAR` | +mynaui:one-diamond-solid+ |
| `cpe` | `VARCHAR` |  |
| `chassis` | `VARCHAR` |  |
| `parent_machine_id` | `INTEGER` | [+mynaui:key+](#machines) |


## cpus


| Name | Type |  |
|------|------|-------------|
| `id` | `INTEGER` | +mynaui:link-one+ |
| `created_at` | `TIMESTAMP` |  |
| `updated_at` | `TIMESTAMP` |  |
| `model_name` | `VARCHAR` |  |
| `vendor` | `VARCHAR` |  |
| `cores` | `INTEGER` |  |
| `machine_id` | `INTEGER` | +mynaui:one-diamond-solid+ [+mynaui:key+](#machines) |


## gpus


| Name | Type |  |
|------|------|-------------|
| `id` | `INTEGER` | +mynaui:link-one+ |
| `created_at` | `TIMESTAMP` |  |
| `updated_at` | `TIMESTAMP` |  |
| `index` | `INTEGER` | +mynaui:one-diamond-solid+ |
| `product` | `VARCHAR` |  |
| `vendor` | `VARCHAR` |  |
| `driver` | `VARCHAR` |  |
| `machine_id` | `INTEGER` | +mynaui:one-diamond-solid+ [+mynaui:key+](#machines) |


## disks


| Name | Type |  |
|------|------|-------------|
| `id` | `INTEGER` | +mynaui:link-one+ |
| `created_at` | `TIMESTAMP` |  |
| `updated_at` | `TIMESTAMP` |  |
| `name` | `VARCHAR` | +mynaui:one-diamond-solid+ |
| `model` | `VARCHAR` |  |
| `size` | `INTEGER` |  |
| `type` | `VARCHAR` |  |
| `controller` | `VARCHAR` |  |
| `partitions` | `VARCHAR` |  |
| `machine_id` | `INTEGER` | +mynaui:one-diamond-solid+ [+mynaui:key+](#machines) |


## network_interfaces


| Name | Type |  |
|------|------|-------------|
| `id` | `INTEGER` | +mynaui:link-one+ |
| `created_at` | `TIMESTAMP` |  |
| `updated_at` | `TIMESTAMP` |  |
| `name` | `VARCHAR` | +mynaui:two-diamond-solid+ |
| `mac` | `VARCHAR` | +mynaui:one-diamond-solid+ |
| `mac_vendor` | `VARCHAR` |  |
| `ip` | `VARCHAR` |  |
| `gateway` | `VARCHAR` |  |
| `flags` | `VARCHAR` |  |
| `tag` | `VARCHAR` | +mynaui:one-diamond-solid+ |
| `machine_id` | `INTEGER` | +mynaui:two-diamond-solid+ [+mynaui:key+](#machines) |


## network_interface_subnets


| Name | Type |  |
|------|------|-------------|
| `network_interface_id` | `INTEGER` | +mynaui:link-one+ [+mynaui:key+](#network_interfaces) |
| `subnetwork_id` | `INTEGER` | +mynaui:link-one+ [+mynaui:key+](#subnetworks) |
| `ip` | `VARCHAR` | +mynaui:link-one+ |
| `mac_subnet` | `VARCHAR` | +mynaui:one-diamond-solid+ |


## packages


| Name | Type |  |
|------|------|-------------|
| `id` | `INTEGER` | +mynaui:link-one+ |
| `created_at` | `TIMESTAMP` |  |
| `updated_at` | `TIMESTAMP` |  |
| `name` | `VARCHAR` | +mynaui:one-diamond-solid+ |
| `version` | `VARCHAR` | +mynaui:one-diamond-solid+ |
| `vendor` | `VARCHAR` |  |
| `manager` | `VARCHAR` |  |
| `install_time_unix` | `INTEGER` |  |
| `files` | `VARCHAR` |  |
| `machine_id` | `INTEGER` | +mynaui:one-diamond-solid+ [+mynaui:key+](#machines) |


## applications


| Name | Type |  |
|------|------|-------------|
| `id` | `INTEGER` | +mynaui:link-one+ |
| `created_at` | `TIMESTAMP` |  |
| `updated_at` | `TIMESTAMP` |  |
| `name` | `VARCHAR` | +mynaui:one-diamond-solid+ |
| `args` | `VARCHAR` |  |
| `pid` | `INTEGER` | +mynaui:one-diamond-solid+ |
| `version` | `VARCHAR` |  |
| `protocol` | `VARCHAR` |  |
| `config` | `VARCHAR` |  |
| `cpe` | `VARCHAR` |  |
| `machine_id` | `INTEGER` | +mynaui:one-diamond-solid+ [+mynaui:key+](#machines) |
| `package_id` | `INTEGER` | [+mynaui:key+](#packages) |


## application_endpoints


| Name | Type |  |
|------|------|-------------|
| `id` | `INTEGER` | +mynaui:link-one+ |
| `created_at` | `TIMESTAMP` |  |
| `updated_at` | `TIMESTAMP` |  |
| `port` | `INTEGER` | +mynaui:one-diamond-solid+ |
| `protocol` | `VARCHAR` | +mynaui:one-diamond-solid+ |
| `addr` | `VARCHAR` | +mynaui:one-diamond-solid+ |
| `tls` | `VARCHAR` |  |
| `fingerprints` | `VARCHAR` |  |
| `application_protocols` | `VARCHAR` |  |
| `saas` | `VARCHAR` |  |
| `application_id` | `INTEGER` | [+mynaui:key+](#applications) |
| `network_interface_id` | `INTEGER` | +mynaui:one-diamond-solid+ [+mynaui:key+](#network_interfaces) |


## users


| Name | Type |  |
|------|------|-------------|
| `id` | `INTEGER` | +mynaui:link-one+ |
| `created_at` | `TIMESTAMP` |  |
| `updated_at` | `TIMESTAMP` |  |
| `uid` | `VARCHAR` | +mynaui:one-diamond-solid+ |
| `gid` | `VARCHAR` |  |
| `name` | `VARCHAR` |  |
| `username` | `VARCHAR` |  |
| `domain` | `VARCHAR` |  |
| `machine_id` | `INTEGER` | +mynaui:one-diamond-solid+ [+mynaui:key+](#machines) |


## user_applications


| Name | Type |  |
|------|------|-------------|
| `id` | `INTEGER` | +mynaui:link-one+ |
| `created_at` | `TIMESTAMP` |  |
| `updated_at` | `TIMESTAMP` |  |
| `user_id` | `INTEGER` | +mynaui:one-diamond-solid+ [+mynaui:key+](#users) |
| `application_id` | `INTEGER` | +mynaui:one-diamond-solid+ [+mynaui:key+](#applications) |
| `linux` | `VARCHAR` |  |


## flows


| Name | Type |  |
|------|------|-------------|
| `id` | `INTEGER` | +mynaui:link-one+ |
| `created_at` | `TIMESTAMP` |  |
| `updated_at` | `TIMESTAMP` |  |
| `src_application_id` | `INTEGER` | +mynaui:one-diamond-solid+ [+mynaui:key+](#applications) |
| `src_network_interface_id` | `INTEGER` | [+mynaui:key+](#network_interfaces) |
| `src_addr` | `VARCHAR` | +mynaui:one-diamond-solid+ |
| `dst_endpoint_id` | `INTEGER` | +mynaui:one-diamond-solid+ [+mynaui:key+](#application_endpoints) |


## endpoint_policies


| Name | Type |  |
|------|------|-------------|
| `id` | `INTEGER` | +mynaui:link-one+ |
| `created_at` | `TIMESTAMP` |  |
| `updated_at` | `TIMESTAMP` |  |
| `endpoint_id` | `INTEGER` | +mynaui:one-diamond-solid+ [+mynaui:key+](#application_endpoints) |
| `action` | `VARCHAR` | +mynaui:one-diamond-solid+ |
| `src_endpoint_id` | `INTEGER` | +mynaui:one-diamond-solid+ [+mynaui:key+](#application_endpoints) |
| `src_addr` | `VARCHAR` | +mynaui:one-diamond-solid+ |
| `priority` | `INTEGER` |  |
| `source` | `VARCHAR` |  |
