Test-only migrations

This directory is used by the server test harness to run migrations and seed
minimal card data derived from `test/decklists/*.csv`. It is not used by
production or `make migrate-local`.

The harness connects to the database named by `DATABASE_URL` only to create a
uniquely named `edhgo_test_*` scratch database. All migrations and test writes
run in that scratch database, which is dropped after the suite. The configured
role therefore needs `CREATE DATABASE` permission, but the administration
database itself is never migrated or reset.

After migrations, `server/main_test.go` supplements the SQL seed with card
names from the committed Archidekt response fixture plus explicitly named test
cards. This keeps provider and deck-building tests deterministic without the
623 MB `All Printings.json` development dataset.

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
