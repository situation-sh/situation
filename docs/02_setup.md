---
title: Setup
summary: Deploy Situation on your infra
---

One of the feature of Situation is that **agents can collaborate** once they share the same database (currently, only postgres is supported). 

## Database 

The first step is to provision a [PostgreSQL](https://www.postgresql.org/) database. 

You can choose to host an instance on-premise (with `docker` for instance) or let a remote service do it for you (like `supabase` or any cloud provider).

Here is an example with docker:
```bash
docker run \
    --rm \
    --detach \
    --name situation-pg \
    -e "POSTGRES_PASSWORD=password" \
    -e "POSTGRES_USER=user" \
    -e "POSTGRES_DB=situation" \
    -p "5432:5432" \
    --health-cmd="pg_isready -U user" \
    --health-interval=2s \
    --health-timeout=2s \
    --health-retries=2 \
    postgres:17.6
```

The paramount thing is to get the DSN of the instance so as to pass it to the agents. In the example above it is something like `postgres://user:password@[ENDPOINT]/situation?sslmode=disable`.

## Agents

Once your db is ready, you can deploy agents anywhere on your infra. We advise to do it incrementally since **IT mapping generally means IT exploring**. 

!!! info "Rule of thumbs"
    - deploy one agent per subnet since every agent is able to grab data from neighboring hosts
    - deploy agent on hosts where exhaustive data collection is mandatory

On each host where agent are deployed, follow the next steps.

### Identifier

Every agent has an internal ID that must be unique across your swarm. 
Stock agent has a default identifier that must be changed.

/// tab | Linux

```bash
situation refresh-id
```

///

/// tab | Windows

```ps1
situation.exe refresh-id
```

///

!!! warning "Important:"
    Once the agent has its unique ID, it must keep it and stay on the same host.


You can check the new id by calling the `id` subcommand.

### Execution

/// tab | Linux

```bash
situation run --db="postgres://user:password@[ENDPOINT]/situation?sslmode=disable"
```

///

/// tab | Windows

```ps1
situation.exe run --db="postgres://user:password@[ENDPOINT]/situation?sslmode=disable"
```

///