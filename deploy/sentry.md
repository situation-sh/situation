---
title: Sentry
summary: Agent monitoring
---

The agent runs can be monitored through [Sentry](https://sentry.io) once you provide your [sentry DSN](https://docs.sentry.io/concepts/key-terms/dsn-explainer/) to the `--sentry` flag.

/// tab | Linux

```bash
situation run --sentry="https://6eb600b24dd5c42fb149cf84d2240bef@o458801043321e512.ingest.us.sentry.io/4509589955543040"
```

///

/// tab | Windows

```ps1
situation.exe run -sentry="https://6eb600b24dd5c42fb149cf84d2240bef@o458801043321e512.ingest.us.sentry.io/4509589955543040"
```

///

The internal sentry client forward logs (not debug logs) in addition to fine-grained monitoring data (module level). It attaches the following metadata (not configurable).

| Metadata                                                                                | Value                        | Example                                    |
| --------------------------------------------------------------------------------------- | ---------------------------- | ------------------------------------------ |
| [`ServerName`](https://docs.sentry.io/platforms/java/configuration/options/#serverName) | Agent ID (hex string format) | `fc097e65503cb3ad9eb8e10f5a617611`         |
| [`Release`](https://docs.sentry.io/platforms/java/configuration/options/#release)       | Agent version                | `0.20.0`                                   |
| [`Dist`](https://docs.sentry.io/platforms/java/configuration/options/#dist)             | Agent commit                 | `d7fa41d254d75496200d6fe0c71b1b4bf13892b1` |

