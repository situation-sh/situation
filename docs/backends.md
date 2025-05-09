---
title: Backends
summary: Dispatch the collected data
order: 40
---

The collected data can be sent to different endpoints, a.k.a. backends.

## Stdout

The simplest way to show what situation ahs collected is to print the final payload to stdout.

/// tab |  Linux
```bash
situation --stdout
```
/// 

/// tab |  Windows
```ps1
situation.exe --stdout
```
///


## File

The payload can also be stored in a file.

/// tab | Linux
```bash
situation --file --file-path=/tmp/situation.json
```
/// 

/// tab | Windows
```ps1
situation.exe --file --file-path="C:\Users\situation.json"
```
/// 

## HTTP

Finally, the http backend is very convenient to send the payload (json) directly to a remote server. 

/// tab | Linux
```bash
situation --http --http-url=http://localhost:8000/situation/ --http-extra-header="X-API-Key=d50deba3-6183-425a-b35c-ef0e030c284e" 
```
/// 

/// tab | Windows
```ps1
situation.exe --http --http-url=http://localhost:8000/situation/ --http-extra-header="X-API-Key=d50deba3-6183-425a-b35c-ef0e030c284e" 
```
/// 

By default it uses the `POST` method, but you can also use `PUT` by using the `--http-method` option. 

Also, it embeds a default authorization header, filled with the agent id:  `Authorization: <agent-id>` (can be modified with the `--http-authorization` flag)



