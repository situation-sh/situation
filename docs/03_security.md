---
title: Security
summary: Recommendations for PostgreSQL
external_links:
  "PostgreSQL reference": https://www.postgresql.org/docs/18/admin.html
---

This page covers security hardening for a PostgreSQL-backed setup where multiple agents share the same database. It only applies to the **PostgreSQL backend** — SQLite deployments are out of scope.

The following topics are covered:

- **TLS** — encrypting connections and verifying server/client identity
- **Roles** — limiting agent privileges to read/write only

Feel free to select and adapt what applies to your environment. In all cases, refer to the official [PostgreSQL administration documentation](https://www.postgresql.org/docs/18/admin.html).

!!! info "Assumptions"
    - A superuser already exists (e.g. `user`/`password`). In Docker, this is set via `POSTGRES_USER` and `POSTGRES_PASSWORD`.
    - Data is stored in the `situation` schema, in a database named `situation` (Docker: `POSTGRES_DB`).
    - The instance is reachable at `db.example.org:5432`.

## TLS

TLS can provide connection encryption as well as server and client authentication.

### Enable SSL mode

Enabling SSL is the first step to use TLS features.
According to the postgres documentation:

> To start in SSL mode, files containing the server certificate and private key must exist.
> By default, these files are expected to be named server.crt and server.key, respectively, in the server's data directory [^postgres-ssl-tcp-basic]

On `postgresql.conf` you can then enable SSL as follows, assuming that you already provide `server.crt` and `server.key` at the right location (`$PGDATA` folder).

```ini
ssl = on
# ssl_cert_file is default to '${PGDATA}/server.crt'
# ssl_key_file is default to '${PGDATA}/server.key'
```

### Reject unencrypted connections

Through the `pg_hba.conf` file, the server can reject any remote connections that do not use SSL. 

```ini
# Trust local (local connection can be made unconditionally)
# conntype  db      user        auth-method 
local       all     all         trust
# Allow ssl connections
# conntype  db      user        address     auth-method 
hostssl     all     all         0.0.0.0/0   scram-sha-256
hostssl     all     all         ::/0        scram-sha-256
# Explicitly reject non-SSL TCP connections
# conntype  db      user        address     auth-method 
hostnossl   all     all         0.0.0.0/0   reject
hostnossl   all     all         ::/0        reject
```

On the client side, at minimum `sslmode=require` must be passed to ensure that nothing will be sent unencrypted. Using `sslmode=verify-full` is recommended to also verify the server identity (see [Client verification](#client-verification) below).

### Client verification

> By default, PostgreSQL will not perform any verification of the server certificate. This means that it is possible to spoof the server identity without the client knowing. [^postgres-ssl-support]

To let agents (or other third-party clients) verifying server's identity we must provide to them the certificate authority that signed the server certificate.

Currently the CA certificate cannot be inlined in the DSN, so you must put it somewhere on the system where the client runs and append `sslmode=verify-full&sslrootcert=/path/to/ca.crt` to the DSN.

Here is an example of the db migration made by the agent.

/// tab | Linux

```bash
situation migrate --db="postgres://user:password@db.example.org:5432/situation?sslmode=verify-full&sslrootcert=ca.crt"
```

///

/// tab | Windows

```ps1
situation.exe migrate --db="postgres://user:password@db.example.org:5432/situation?sslmode=verify-full&sslrootcert=ca.crt"
```

///

!!! warning "Note"
    Setting `sslmode=verify-ca` will check the server certificate (i.e. it is well signed by the CA) but it won't verify that the server host name matches the name stored in the server certificate.

### Client certificate

The PostgreSQL server can also authenticate client 
through SSL certificates (in SSL connections only of course).

You can modify `pg_hba.conf` as follows to activate this method.

```ini
# Allow ssl connections with client cert auth
# conntype  db      user        address     auth-method 
hostssl     all     all         0.0.0.0/0   cert
hostssl     all     all         ::/0        cert
```
According to the docs:

> The CN (Common Name) attribute of the certificate will be compared to the requested database user name [^postgres-cert-authentication]

So the server will check whether the client cert is well signed by the trusted CA and if its CN matches the user in the DSN. On the `postgresql.conf` you should append:

```ini
# trusted certificate authorities
ssl_ca_file = /path/to/ca.crt
```

On the client side, client certificate and private key must be passed  through the `sslcert` and `sslkey` options (see the example below). The password can then be omitted.

/// tab | Linux

```bash
situation migrate --db="postgres://user:@db.example.org:5432/situation?sslmode=verify-full&sslrootcert=ca.crt&sslcert=client.crt&sslkey=client.key"
```

///

/// tab | Windows

```ps1
situation.exe migrate --db="postgres://user:@db.example.org:5432/situation?sslmode=verify-full&sslrootcert=ca.crt&sslcert=client.crt&sslkey=client.key"
```

///

!!! warning "Warning"
    Some options like `sslcert` or `sslkey` are not supported by postgres clients.

## Roles

We can also add roles to limit privileges. 
For instance, we can keep the superuser role to run migrations and then create a basic `agent` role that will only perform read/write and won't be able to modify tables.

In the postgres instance:
```sql
-- Agents: read/write data only, cannot touch schema
CREATE ROLE agent LOGIN PASSWORD 'secure-password';
GRANT USAGE ON SCHEMA situation TO agent;
```

Assuming migrations have been run with the superuser (or another user with enough privileges), you can run the agent with `--no-migration` so that it does not try to run the migrations.

/// tab | Linux

```bash
situation run --no-migration --db="postgres://agent:secure-password@db.example.org:5432/situation?..."
```

///

/// tab | Windows

```ps1
situation.exe run --no-migration --db="postgres://agent:secure-password@db.example.org:5432/situation?..."
```

///

[^postgres-ssl-tcp-basic]: [https://www.postgresql.org/docs/current/ssl-tcp.html#SSL-SETUP](https://www.postgresql.org/docs/current/ssl-tcp.html#SSL-SETUP)

[^postgres-ssl-support]: [https://www.postgresql.org/docs/current/libpq-ssl.html#LIBPQ-SSL](https://www.postgresql.org/docs/current/libpq-ssl.html#LIBPQ-SSL)

[^postgres-cert-authentication]: [https://www.postgresql.org/docs/current/auth-cert.html#AUTH-CERT](https://www.postgresql.org/docs/current/auth-cert.html#AUTH-CERT)