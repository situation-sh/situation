---
title: Scheduled task 
summary: Run situation from time to time
---

Once downloaded, the binary can install itself as a cronjob (Linux) or as a scheduled task (Windows). Why not a service? Situation should be run regularly but _occasionally_. It is not meant to run indefinitely in the background.

## Installing

Here is an example that installs the scheduled task that triggers every day at midnight.

/// tab | Linux

```bash
situation cron --task-start 00:00:00
```

///

/// tab | Windows

```ps1
situation.exe task --task-start 00:00:00
```

///

!!! warning "Important"
    You need admin privileges to install the scheduled task. On Windows, it creates a `SYSTEM` task and on Linux, it writes the job to `/etc/cron.d/situation`.

Any **run** parameters passed to the command line will be appended to the task. It means that if you run the command below,

/// tab | Linux

```bash
situation cron --task-start 00:00:00 --db=db.sqlite
```

///

/// tab | Windows

```ps1
situation.exe task --task-start 00:00:00 --db=db.sqlite
```

///

the binary will be run with the flags `--db=db.sqlite`.

## Uninstalling

The task can be removed with the `--uninstall` flag (privileges still required).

/// tab | Linux

```bash
situation cron --uninstall
```

///

/// tab | Windows

```ps1
situation.exe task --uninstall
```

///
