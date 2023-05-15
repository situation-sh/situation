---
title: Pre-built binaries
---

The agent is currently available for Linux and Windows on x86_64 architectures.

=== "wget"

    ```bash
    wget -qO situation {{ github_repo }}/releases/download/{{ latest_tag() }}/{{ latest_binary("linux") }}
    chmod +x ./situation
    ```

=== "curl"

    ```bash
    curl -sLo ./situation {{ github_repo }}/releases/download/{{ latest_tag() }}/{{ latest_binary("linux") }}
    chmod +x ./situation
    ```

=== "PowerShell"

    ```ps1
    Invoke-RestMethod -OutFile situation.exe -Uri {{ github_repo }}/releases/download/{{ latest_tag() }}/{{ latest_binary("windows") }}
    ```
