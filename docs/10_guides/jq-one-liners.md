---
title: jq one-liners
summary: Quick and dirty stdout examples
order: 10
---

While `situation` aims to send collected data to a "remote" place for further analysis,
its output can be quickly worked by basic cli tools like [jq](https://stedolan.github.io/jq/).

## Network discovery

/// tab | Linux

```bash
situation --stdout | jq -r '.machines[] | .nics[] | .mac + "\t" + .ip'
```
///

/// tab | Windows

```ps1
situation.exe --stdout | jq -r '.machines[] | .nics[] | .mac + \"\t\" + .ip'
```
///

```shell
aa:f4:b5:eb:ba:71       192.168.1.11
0e:de:c8:62:b5:1c       192.168.1.54
18:19:ba:91:b7:c7       192.168.1.13
c1:d3:d2:ab:41:cb       192.168.1.31
47:20:7b:a3:fb:2b       192.168.1.57
```

You can even put the results in a csv file:

/// tab | Linux

```bash
situation --stdout | jq -r '.machines[] | .nics[] | [.mac,.ip] | @csv' > output.csv
```

///

/// tab | Windows

```ps1
situation.exe --stdout | jq -r '.machines[] | .nics[] | [.mac,.ip] | @csv' > output.csv
```

///

## Open ports

/// tab | Linux

```bash
situation --stdout | jq -r '(.machines[] | .packages[] | .applications[] | .endpoints[] | [.addr,(.port|tostring)+"/"+.protocol])|@tsv'
```
///

/// tab | Windows

```ps1
situation.exe --stdout | jq -r '(.machines[] | .packages[] | .applications[] | .endpoints[] | [.addr,(.port|tostring)+\"/\"+.protocol])|@tsv'
```
///

```shell
192.168.1.1     53/tcp
192.168.1.1     80/tcp
192.168.1.1     1287/tcp
192.168.1.11    139/tcp
192.168.1.11    445/tcp
192.168.1.11    22/tcp
192.168.1.54    80/tcp
192.168.1.13    53/tcp
192.168.1.13    80/tcp
192.168.1.13    22/tcp
192.168.1.13    443/tcp
```

## List services

/// tab | Linux

```bash
situation | jq -r '(["Service","Address","Port"]|(., map(length*"-"))), (.machines[]|select(.hosted_agent)|.packages[]|.applications[]|.name as $n|.endpoints[]|[$n,.addr,(.port|tostring)+"/"+.protocol])|@tsv' | column -ts $'\t'
```

///

/// tab | Windows

```ps1
situation.exe | jq -r '([\"Service\",\"Address\",\"Port\"]), (.machines[]|select(.hosted_agent)|.packages[]|.applications[]|.name as $n|.endpoints[]|[$n,.addr,(.port|tostring)+\"/\"+.protocol])|@csv' | ConvertFrom-Csv
```

///

```shell
Service          Address        Port
-------          -------        ----
systemd-resolve  0.0.0.0        5355/tcp
systemd-resolve  ::             5355/tcp6
rpcbind          0.0.0.0        111/tcp
rpcbind          ::             111/tcp6
dnsmasq          192.168.122.1  53/tcp
rpc.statd        0.0.0.0        35645/tcp
rpc.statd        ::             50443/tcp6
systemd          ::             6556/tcp6
kdeconnectd      ::             1716/tcp6
```
