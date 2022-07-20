---
title: Pre-built binaries
---

The agent is currently available for Linux and Windows on x86_64 architectures.

=== "wget"

    ```bash
    wget -qO situation https://{{ variables.go_module }}/releases/download/v{{ variables.version }}/situation-{{ variables.version }}-amd64-linux && chmod +x situation
    ```

=== "curl"

    ```bash
    curl -sLo ./situation https://{{ variables.go_module }}/releases/download/v{{ variables.version }}/situation-{{ variables.version }}-amd64-linux && chmod +x ./situation
    ```

=== "PowerShell"

    ```ps1
    Invoke-RestMethod -OutFile situation.exe -Uri situation https://{{ variables.go_module }}/releases/download/v{{ variables.version }}/situation-{{ variables.version }}-amd64-windows.exe
    ```
