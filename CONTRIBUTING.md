# Contributing

Contributions are welcome, and they are greatly appreciated!

You can contribute in many ways:

## Types of Contributions

### Report Bugs

Report bugs at [https://github.com/Identityplane/GoAM/issues](https://github.com/Identityplane/GoAM/issues).

If you are reporting a bug, please include:

* Your operating system name and version.
* Any details about your local setup that might be helpful in troubleshooting.
* Detailed steps to reproduce the bug.

### Submit Feedback

The best way to send feedback is to contact the core team via email info@identityplane.com. If you have ideas on how to make identity better, we'd love to hear from you!

**Current Status**: The core team is currently focused on making GoAM ready for production. During this phase, we appreciate your input about:

* Architecture suggestions and improvements
* Feature requests and use cases
* General feedback and ideas
* Bug reports and issues

While we may not be able to implement all suggestions immediately, your feedback helps us build a better product. We're committed to creating a robust, production-ready identity management solution, and your insights are invaluable to that process.

If you are proposing a feature:

* Explain in detail how it would work.
* Keep the scope as narrow as possible, to make it easier to implement.
* Remember that this is a volunteer-driven project, and that contributions
  are welcome :)

## Setup dev environment

### Go backend (API server)

Prerequisites: Go (see `go.mod` for the required version).

```
go mod tidy
go run cmd/main.go
Visit: http://localhost:8081/readyz
```

### VS Code debugging

Install [Delve](https://github.com/go-delve/delve) (`dlv`) on your machine; the Go extension uses it to debug. For example: `go install github.com/go-delve/delve/cmd/dlv@latest` (ensure `$(go env GOPATH)/bin` is on your `PATH`).

[`.vscode/launch.json`](.vscode/launch.json) defines two launch configurations:

- **Goiam SQLite** — runs the server with the default SQLite setup (suitable for everyday development).
- **Goiam Postgres** — runs against `postgres://goiam:secret123@localhost:5432/goiamdb`. PostgreSQL must be available on `localhost:5432` with those credentials; run it locally (often via Kubernetes—see below—and port-forward or expose `5432`).

### Auth UI (web app)

The Auth UI is a separate web application. In development, GoAM reverse-proxies it under the realm route prefix:
- `/{tenant}/{realm}/authui/…` (e.g. `/acme/customers/authui/login?debug`)

Prerequisites: Node.js + `pnpm`.

```
cd auth-ui
pnpm install
pnpm run dev
Visit: http://localhost:8081/acme/customers/authui/login?debug
```

### Kubernetes (optional)

Prefer OrbStack; Minikube also works (you may need to adapt commands in the `Makefile`).

Prerequisites:
- kubectl with a working k8s cluster
- helm
- (optional) k9s

```
make docker-all   # build Docker images (GoAM + admin UI)
make helm-deploy  # deploy locally via Helm
```
