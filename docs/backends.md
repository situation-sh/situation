---
title: Backends
---

# Backends

## Stdout (default)

The default behavior of Situation is to print the final payload to stdout.

=== "CLI"

    ```bash
    situation --backends.stdout.enabled=false
    ```

=== "YAML"

    ```yaml
    backends:
        stdout:
            enabled: false
    ```


!!!warning
    Due to a bug in a third party library, the `enabled` attribute cannot be changed in the configuration file (see [this issue](https://github.com/urfave/cli/issues/1395))

## File

The payload can also be stored in a file.

=== "CLI"

    ```bash
    situation --backends.file.enabled=true --backends.file.format=json --backends.file.path=/tmp/situation.json
    ```

=== "YAML"

    ```yaml
    backends:
        file:
            enabled: true
            format: json
            path: /tmp/situation.json
    ```

## HTTP

Finally, the http backend is very convenient to send the payload (json) directly to a remote server.

=== "CLI"

    ```bash
    situation --backends.http.enabled=true --backends.http.url=http://localhost:8000/situation/ --backends.http.method=POST --backends.http.header.content-type=application/json --backends.http.header.authorization="Bearer <APIKEY>" 
    ```

=== "YAML"

    ```yaml
    backends:
        http:
            enabled: true
            url: http://localhost:8000/situation/
            method: POST
            header:
                content-type: application/json
                authorization: "Bearer <APIKEY>" 
    ```