# AGENTS.md

## Cursor Cloud specific instructions

The startup update script already installs all dependencies (Go modules, `auth-ui` pnpm deps, root npm deps). The notes below cover non-obvious caveats for developing/running the two services. Standard commands live in `CONTRIBUTING.md` and the `makefile`.

### Services

| Service | Location | Dev command (run from) | Port | Notes |
|---------|----------|------------------------|------|-------|
| GoAM backend | `cmd/main.go` | `go run ./cmd/main.go` (repo root) | 8081 | Core IAM server (auth-flow engine, OAuth2/OIDC, admin API, auth-ui reverse proxy). |
| Auth UI | `auth-ui/` | `pnpm run dev` (in `auth-ui`) | 4000 | Next.js login/registration UI. Reverse-proxied by the backend. |

- Run the backend from the repo root: it reads `goam.yaml` from the current directory.
- The DB is embedded SQLite (`goiam.db`, auto-created and auto-migrated via `run_db_migrations: true`); no external database is needed for local dev. A `goiam.db` file appears in the repo root after first start (gitignored / do not commit).
- Admin auth is disabled in `goam.yaml` (`unsafe_disable_admin_auth: true`) and a seed `admin`/`admin` user is created — demo config only.
- For full end-to-end auth UX, both services must run. Browse flows via the backend proxy: `http://localhost:8081/{tenant}/{realm}/authui/{flow}` (example realm: `acme/customers`).

### pnpm version caveat (important)

`auth-ui` must be installed with **pnpm v10**. Corepack defaults to pnpm v11, which no longer reads the `pnpm.overrides` field in `auth-ui/package.json`, so `pnpm install --frozen-lockfile` fails with `ERR_PNPM_LOCKFILE_CONFIG_MISMATCH`. The update script pins pnpm 10 via `corepack prepare pnpm@10.33.3 --activate`; keep using pnpm 10 for `auth-ui`.

### Auth flow routing / debugging

- Each realm exposes flows by id at `/{tenant}/{realm}/authui/{flowId}`. Not every YAML file in `config/tenants/.../flows/` is wired to a same-named route. The `acme/customers` realm has `login` configured; navigating to an unconfigured flow id shows "Flow not found".
- Append `?debug` to open the Flow Debugger, which lets you pick any flow (e.g. select `username-password-register` to exercise registration). This is the easiest way to test individual flows.

### Lint / test / build

- Backend: `make test` (go test `-short`), `go vet ./...`, `make build` (binary to `bin/goiam`).
- Root JS: `npm test` (Jest unit tests for shared `auth-ui/lib` utils).
- Note: `auth-ui`'s `pnpm run lint` is currently broken because `eslint` is not declared in `auth-ui/package.json` dependencies (pre-existing); `eslint: not found`.
