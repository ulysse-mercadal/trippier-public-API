# Contributing

## Branching

- `main` is always deployable.
- Branch off `main` for every change: `feat/…`, `fix/…`, `chore/…`.
- Open a PR against `main`. Squash-merge once CI is green.
- Do not push directly to `main`.

## Running tests

No local Go or Python install required — everything runs in Docker.

```bash
make test           # all services
make test-go-auth   # auth-api only
make test-go-poi    # poi-api only
make test-python    # itinerary-api only
```

For a faster inner loop when working on a single service, you can run Go tests directly if you have Go installed:

```bash
cd auth-api && go test ./...
cd poi-api  && go test ./...
```

## Linting

```bash
make lint           # golangci-lint (Go) + ruff + mypy (Python)
```

CI runs lint and tests on every push and PR (see `.github/workflows/ci.yml`). A PR with a failing lint check will not be merged.

## Adding a database migration

1. Create `auth-api/internal/db/migrations/NNN_description.sql` (next sequential number).
2. Write idempotent SQL — use `IF NOT EXISTS` / `IF EXISTS` so re-running is safe.
3. Add a `//go:embed migrations/NNN_description.sql` directive and append the variable to the migration slice in `auth-api/internal/db/postgres.go`.
4. The migration runs automatically at service startup.

## Adding a provider to poi-api

1. Implement `providers.Provider` in `poi-api/internal/providers/<name>/`.
2. Wire it up in `poi-api/cmd/server/main.go` → `buildProviders()`.
3. Add the provider constant to `types.AllProviders` in `poi-api/pkg/types/`.
4. Write at least one unit test using the `mockProvider` pattern in `search/service_test.go`.

## Opening a PR

- Keep PRs focused — one logical change per PR.
- Include a short description of what changed and why.
- If you're fixing a bug, reference the issue or describe how to reproduce it.
- Make sure `make test` and `make lint` pass locally before pushing.
