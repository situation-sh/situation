# Store

The **Store** is a relational database layer where modules persist collected data. It is implemented by the `BunStorage` type in `pkg/store/`, built on top of the [Bun ORM](https://bun.uptrace.dev/) and supports both **SQLite** (embedded, default) and **PostgreSQL** (remote).

!!! warning
    The store and the models may evolve. Changes to models (add/modify/remove attributes) can have a wide impact. Always check existing methods before writing raw queries.

## Usage

Inside a module's `Run` function, the storage can be extracted from the context.

```go
func (m *MyModule) Run(ctx context.Context) error {
    storage := getStorage(ctx)
    // ...
}
```

Then, you can either use storage helpers to read/write the database or code your own query. 

```go
// some helpers are diretly available
host := storage.GetOrCreateHost(ctx)

// Otherwise you can build your own queries 
// using the underlying Bun DB
err := storage.DB().
    NewUpdate().
    Model((*models.Machine)(nil)).
    Where("id = ?", host.ID).
    Set("hostname = ?", hostname).
    Returning("*").
    Scan(ctx, host)
```

!!! info "Info"
    The modules are likely to build their own queries since they collect different things.

