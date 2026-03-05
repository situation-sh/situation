---
title: MCP
summary: Connect AI to your IT infrastructure
new: true
---

Ready to **prompt your infra**? An MCP server has been integrated into the agent, the latter acting as a proxy between your favourite LLM and the database.

!!! info "Transport"
    Currently, only the `stdio` transport is supported, mainly for security and agent size concerns.

!!! info "Security"
    The database is open in read-only mode (read-only transactions for postgres), preventing it from destructive operations.

## Command

You can start the server as follows, but generally this command is wrapped by the tool that will use it.

/// tab | Linux

```bash
situation mcp --db="file:db.sqlite"
```

///

/// tab | Windows

```ps1
situation.exe mcp --db="file:db.sqlite"
```

///

!!! warning "Warning"
    The database DSN can require special characters (like `&`) that won't be
    well escaped by the tools using this MCP. We advise passing the DSN
    through the `SITUATION_DB` environment variable.

## Integration

Several LLM (and tools above) support the `mcp.json` format (filename and locations may change, read their... manual). 
You can then easily provide the situation MCP.

```json
{
    "mcpServers": {
        "situation": {
            "type":"stdio",
            "command":"/path/to/situation",
            "args":["mcp"],
            "env": {
                "SITUATION_DB":"postgresql://user:password@127.0.0.1:5432/situation?ssl=disable&connect_timeout=10"
            }
        }
    }
}
```

## Inspection

You can test the MCP server through the [MCP Inspector](https://modelcontextprotocol.io/docs/tools/inspector). Note the quote escape to ensure that the `db` parameter is protected.

/// tab | `npx`

```bash
npx @modelcontextprotocol/inspector /path/to/situation mcp --db="\"file:db.sqlite\""
```

///

/// tab | `bunx`

```bash
bunx @modelcontextprotocol/inspector /path/to/situation mcp --db="\"file:db.sqlite\""
```

///