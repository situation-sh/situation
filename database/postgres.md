---

title: PostgreSQL

summary: Database schema for PostgreSQL storage

---

+mynaui:link-one+ Primary Key&emsp;+mynaui:key+ Foreign Key&emsp;+mynaui:one-diamond-solid+ +mynaui:two-diamond-solid+ Unique

## subnetworks


| Name | Type |  |
|------|------|-------------|
| `id` | `BIGINT` | +mynaui:link-one+ |
| `created_at` | `TIMESTAMPTZ` |  |
| `updated_at` | `TIMESTAMPTZ` |  |
| `network_cidr` | `VARCHAR` | +mynaui:one-diamond-solid+ |
| `network_addr` | `VARCHAR` |  |
| `mask_size` | `BIGINT` |  |
| `ip_version` | `BIGINT` |  |
| `gateway` | `VARCHAR` |  |
| `vlan_id` | `BIGINT` |  |
| `tag` | `VARCHAR` | +mynaui:one-diamond-solid+ |


## machines


| Name | Type |  |
|------|------|-------------|
| `id` | `BIGINT` | +mynaui:link-one+ |
| `created_at` | `TIMESTAMPTZ` |  |
| `updated_at` | `TIMESTAMPTZ` |  |
| `hostname` | `VARCHAR` |  |
| `host_id` | `VARCHAR` | +mynaui:one-diamond-solid+ |
| `arch` | `VARCHAR` |  |
| `platform` | `VARCHAR` |  |
| `distribution` | `VARCHAR` |  |
| `distribution_version` | `VARCHAR` |  |
| `distribution_family` | `VARCHAR` |  |
| `uptime` | `BIGINT` |  |
| `agent` | `VARCHAR` | +mynaui:one-diamond-solid+ |
| `cpe` | `VARCHAR` |  |
| `chassis` | `VARCHAR` |  |
| `parent_machine_id` | `BIGINT` | [+mynaui:key+](#machines) |


## cpus


| Name | Type |  |
|------|------|-------------|
| `id` | `BIGINT` | +mynaui:link-one+ |
| `created_at` | `TIMESTAMPTZ` |  |
| `updated_at` | `TIMESTAMPTZ` |  |
| `model_name` | `VARCHAR` |  |
| `vendor` | `VARCHAR` |  |
| `cores` | `BIGINT` |  |
| `machine_id` | `BIGINT` | +mynaui:one-diamond-solid+ [+mynaui:key+](#machines) |


## gpus


| Name | Type |  |
|------|------|-------------|
| `id` | `BIGINT` | +mynaui:link-one+ |
| `created_at` | `TIMESTAMPTZ` |  |
| `updated_at` | `TIMESTAMPTZ` |  |
| `index` | `BIGINT` | +mynaui:one-diamond-solid+ |
| `product` | `VARCHAR` |  |
| `vendor` | `VARCHAR` |  |
| `driver` | `VARCHAR` |  |
| `machine_id` | `BIGINT` | +mynaui:one-diamond-solid+ [+mynaui:key+](#machines) |


## disks


| Name | Type |  |
|------|------|-------------|
| `id` | `BIGINT` | +mynaui:link-one+ |
| `created_at` | `TIMESTAMPTZ` |  |
| `updated_at` | `TIMESTAMPTZ` |  |
| `name` | `VARCHAR` | +mynaui:one-diamond-solid+ |
| `model` | `VARCHAR` |  |
| `size` | `BIGINT` |  |
| `type` | `VARCHAR` |  |
| `controller` | `VARCHAR` |  |
| `partitions` | `JSON` |  |
| `machine_id` | `BIGINT` | +mynaui:one-diamond-solid+ [+mynaui:key+](#machines) |


## network_interfaces


| Name | Type |  |
|------|------|-------------|
| `id` | `BIGINT` | +mynaui:link-one+ |
| `created_at` | `TIMESTAMPTZ` |  |
| `updated_at` | `TIMESTAMPTZ` |  |
| `name` | `VARCHAR` | +mynaui:one-diamond-solid+ |
| `mac` | `VARCHAR` | +mynaui:two-diamond-solid+ |
| `mac_vendor` | `VARCHAR` |  |
| `ip` | `VARCHAR[]` |  |
| `gateway` | `VARCHAR` |  |
| `flags` | `JSON` |  |
| `tag` | `VARCHAR` | +mynaui:two-diamond-solid+ |
| `machine_id` | `BIGINT` | +mynaui:two-diamond-solid+ [+mynaui:key+](#machines) |


## network_interface_subnets


| Name | Type |  |
|------|------|-------------|
| `network_interface_id` | `BIGINT` | +mynaui:link-one+ [+mynaui:key+](#network_interfaces) |
| `subnetwork_id` | `BIGINT` | +mynaui:link-one+ [+mynaui:key+](#subnetworks) |
| `ip` | `VARCHAR` | +mynaui:link-one+ |
| `mac_subnet` | `VARCHAR` | +mynaui:one-diamond-solid+ |


## packages


| Name | Type |  |
|------|------|-------------|
| `id` | `BIGINT` | +mynaui:link-one+ |
| `created_at` | `TIMESTAMPTZ` |  |
| `updated_at` | `TIMESTAMPTZ` |  |
| `name` | `VARCHAR` | +mynaui:one-diamond-solid+ |
| `version` | `VARCHAR` | +mynaui:one-diamond-solid+ |
| `vendor` | `VARCHAR` |  |
| `manager` | `VARCHAR` |  |
| `install_time_unix` | `BIGINT` |  |
| `files` | `JSONB` |  |
| `machine_id` | `BIGINT` | +mynaui:one-diamond-solid+ [+mynaui:key+](#machines) |


## applications


| Name | Type |  |
|------|------|-------------|
| `id` | `BIGINT` | +mynaui:link-one+ |
| `created_at` | `TIMESTAMPTZ` |  |
| `updated_at` | `TIMESTAMPTZ` |  |
| `name` | `VARCHAR` | +mynaui:one-diamond-solid+ |
| `args` | `JSONB` |  |
| `pid` | `BIGINT` | +mynaui:one-diamond-solid+ |
| `version` | `VARCHAR` |  |
| `protocol` | `VARCHAR` |  |
| `config` | `JSON` |  |
| `cpe` | `VARCHAR` |  |
| `machine_id` | `BIGINT` | +mynaui:one-diamond-solid+ [+mynaui:key+](#machines) |
| `package_id` | `BIGINT` | [+mynaui:key+](#packages) |


## application_endpoints


| Name | Type |  |
|------|------|-------------|
| `id` | `BIGINT` | +mynaui:link-one+ |
| `created_at` | `TIMESTAMPTZ` |  |
| `updated_at` | `TIMESTAMPTZ` |  |
| `port` | `INTEGER` | +mynaui:one-diamond-solid+ |
| `protocol` | `VARCHAR` | +mynaui:one-diamond-solid+ |
| `addr` | `VARCHAR` | +mynaui:one-diamond-solid+ |
| `tls` | `JSON` |  |
| `fingerprints` | `JSON` |  |
| `application_protocols` | `JSONB` |  |
| `saas` | `VARCHAR` |  |
| `application_id` | `BIGINT` | [+mynaui:key+](#applications) |
| `network_interface_id` | `BIGINT` | +mynaui:one-diamond-solid+ [+mynaui:key+](#network_interfaces) |


## users


| Name | Type |  |
|------|------|-------------|
| `id` | `BIGINT` | +mynaui:link-one+ |
| `created_at` | `TIMESTAMPTZ` |  |
| `updated_at` | `TIMESTAMPTZ` |  |
| `uid` | `VARCHAR` | +mynaui:one-diamond-solid+ |
| `gid` | `VARCHAR` |  |
| `name` | `VARCHAR` |  |
| `username` | `VARCHAR` |  |
| `domain` | `VARCHAR` |  |
| `machine_id` | `BIGINT` | +mynaui:one-diamond-solid+ [+mynaui:key+](#machines) |


## user_applications


| Name | Type |  |
|------|------|-------------|
| `id` | `BIGINT` | +mynaui:link-one+ |
| `created_at` | `TIMESTAMPTZ` |  |
| `updated_at` | `TIMESTAMPTZ` |  |
| `user_id` | `BIGINT` | +mynaui:one-diamond-solid+ [+mynaui:key+](#users) |
| `application_id` | `BIGINT` | +mynaui:one-diamond-solid+ [+mynaui:key+](#applications) |
| `linux` | `VARCHAR` |  |


## flows


| Name | Type |  |
|------|------|-------------|
| `id` | `BIGINT` | +mynaui:link-one+ |
| `created_at` | `TIMESTAMPTZ` |  |
| `updated_at` | `TIMESTAMPTZ` |  |
| `src_application_id` | `BIGINT` | +mynaui:one-diamond-solid+ [+mynaui:key+](#applications) |
| `src_network_interface_id` | `BIGINT` | [+mynaui:key+](#network_interfaces) |
| `src_addr` | `VARCHAR` | +mynaui:one-diamond-solid+ |
| `dst_endpoint_id` | `BIGINT` | +mynaui:one-diamond-solid+ [+mynaui:key+](#application_endpoints) |


## endpoint_policies


| Name | Type |  |
|------|------|-------------|
| `id` | `BIGINT` | +mynaui:link-one+ |
| `created_at` | `TIMESTAMPTZ` |  |
| `updated_at` | `TIMESTAMPTZ` |  |
| `endpoint_id` | `BIGINT` | +mynaui:one-diamond-solid+ [+mynaui:key+](#application_endpoints) |
| `action` | `VARCHAR` | +mynaui:one-diamond-solid+ |
| `src_endpoint_id` | `BIGINT` | +mynaui:one-diamond-solid+ [+mynaui:key+](#application_endpoints) |
| `src_addr` | `VARCHAR` | +mynaui:one-diamond-solid+ |
| `priority` | `BIGINT` |  |
| `source` | `VARCHAR` |  |
