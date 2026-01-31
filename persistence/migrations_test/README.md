Test-only migrations

This directory is used by the server test harness to run migrations and seed
minimal card data derived from `test/decklists/*.csv`. It is not used by
production or `make migrate-local`.

If your local test DB gets marked dirty (e.g. after a failed test migration),
you can reset it with:

```bash
migrate \
  -database "postgres://edhgo:edhgo@localhost:5432/edhgo?sslmode=disable" \
  -source "file://./persistence/migrations_test" \
  force 20231018205455
```

Then rerun:

```bash
go test ./server -count=1
```
