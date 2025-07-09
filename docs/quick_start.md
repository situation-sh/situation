# Quick start

## Run

You guess it. To run the agent, you don't need to provide any extra configuration.

=== "Linux"

    ```bash
    situation
    ```

=== "Windows"

    ```ps1
    situation.exe
    ```

By default, the json output is printed to stdout (you can find the json schema in the [latest release](https://github.com/situation-sh/situation/releases/latest)).

=== "Linux"

    ```json { .scrollable }
    --8<-- "docs/data/linux.json"
    ```


=== "Windows"

    ```json { .scrollable }
    --8<-- "docs/data/windows.json"
    ```

So what you should do next, is to **pipe that json to another tool** (like `jq` see in the [guide](./10_guides/jq-one-liners.md)).

## Other commands

You can get the output json schema with the `schema` subcommand.

=== "Linux"

    ```bash
    situation schema
    ```

=== "Windows"

    ```ps1
    situation.exe schema
    ```

Every agent as an internal UUID (`cafecafe-cafe-cafe-cafe-cafecafecafe`) by default.
This can be printed with `id` subcommand,

=== "Linux"

    ```bash
    situation id
    ```

=== "Windows"

    ```ps1
    situation.exe id
    ```

and refreshed with `refresh-id` subcommand.

=== "Linux"

    ```bash
    situation refresh-id
    ```

=== "Windows"

    ```ps1
    situation.exe refresh-id
    ```

## One-liners

While `situation` aims to send collected data to a "remote" place for further analysis,
its output can be quickly worked by basic cli tools like [jq](https://stedolan.github.io/jq/).

### Network discovery

=== "Linux"

    ```bash
    situation | jq -r '.machines[] | .nics[] | .mac + "\t" + .ip'
    ```

=== "Windows"

    ```ps1
    situation.exe | jq -r '.machines[] | .nics[] | .mac + \"\t\" + .ip'
    ```

```console
aa:f4:b5:eb:ba:71       192.168.1.11
0e:de:c8:62:b5:1c       192.168.1.54
18:19:ba:91:b7:c7       192.168.1.13
c1:d3:d2:ab:41:cb       192.168.1.31
47:20:7b:a3:fb:2b       192.168.1.57
```

You can even put the results in a csv file:

=== "Linux"

    ```bash
    situation | jq -r '.machines[] | .nics[] | [.mac,.ip] | @csv' > output.csv
    ```

=== "Windows"

    ```ps1
    situation.exe | jq -r '.machines[] | .nics[] | [.mac,.ip] | @csv' > output.csv
    ```

### Open ports

=== "Linux"

    ```bash
    situation | jq -r '(.machines[] | .packages[] | .applications[] | .endpoints[] | [.addr,(.port|tostring)+"/"+.protocol])|@tsv'
    ```

=== "Windows"

    ```ps1
    situation.exe | jq -r '(.machines[] | .packages[] | .applications[] | .endpoints[] | [.addr,(.port|tostring)+\"/\"+.protocol])|@tsv'
    ```

```console
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

### List services

=== "Linux"

    ```bash
    situation | jq -r '(["Service","Address","Port"]|(., map(length*"-"))), (.machines[]|select(.hosted_agent)|.packages[]|.applications[]|.name as $n|.endpoints[]|[$n,.addr,(.port|tostring)+"/"+.protocol])|@tsv' | column -ts $'\t'
    ```

=== "Windows"

    ```ps1
    situation.exe | jq -r '([\"Service\",\"Address\",\"Port\"]), (.machines[]|select(.hosted_agent)|.packages[]|.applications[]|.name as $n|.endpoints[]|[$n,.addr,(.port|tostring)+\"/\"+.protocol])|@csv' | ConvertFrom-Csv
    ```

```console
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
