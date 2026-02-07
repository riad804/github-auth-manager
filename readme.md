# git-auth-manager (gham)

A small Go CLI to manage multiple Git accounts per repository using HTTPS tokens and OS secure storage.

Key points:
- CLI binary name: `gham` (see `cmd/auth-manager`)
- Module: `github.com/riad804/auth-manager`
- Built with `cobra` and uses OS keyring for storing tokens

Features
- Repository-scoped account linking (link a repo to a saved account)
- OAuth-based login for providers (tokens stored in OS keyring)
- Commands: `login`, `logout`, `accounts`, `link`, `unlink`, `status`, `doctor`, `credential`

Quickstart
1. Build:

```bash
go build -o gham ./cmd
```

2. Run the CLI (see subcommands):

```bash
./gham --help
./gham login --help
```

3. Typical flow:
- `gham login --provider github --id work`
- `cd /path/to/repo && gham link` (link repo to an account)
- Use `git pull` / `git push` as normal — the credential helper supplies the correct token

Project layout
- `cmd/` — CLI entry and subcommands (`cmd/auth-manager`)
- `auth-manager/` — command implementations (login, link, logout, doctor, etc.)
- `internals/` — auth utilities, providers, git helpers
- `models/` — data models for accounts/tokens
- `storage/` — config and keyring initialization

Build & Test
- Build: `go build ./...` or `go build -o gham ./cmd`
- Run tests (if present): `go test ./...`

Development notes
- Config directory: `~/.config/git-auth-manager` (created automatically)
- The CLI root command is defined with `Use: "gham"` in `cmd/auth-manager/root.go`

Creating OAuth apps
- Register OAuth apps for providers and set appropriate `client_id` values in `internals/auth/providers.go`.

.gitignore and Safety
- Keep secrets and local build artifacts out of git. Add `.env`, local key files, and editor directories to `.gitignore`.

Contributing
- Fork, create a branch, run `go test ./...`, and open a PR.

Usage Examples
- Login to a provider (opens browser):

```bash
gham login --provider github --id work
```

- List saved accounts:

```bash
gham accounts
```

- Link current repository to an account (interactive):

```bash
cd /path/to/repo
gham link
```

- Link non-interactively by id:

```bash
gham link --id work
```

- Show current repo status (which account is linked):

```bash
gham status
```

- Unlink repository:

```bash
gham unlink
```

- Logout and remove an account by id:

```bash
gham logout --id work
```

- Run setup checks / configure credential helper:

```bash
gham doctor
```
---
Updated README to reflect actual package and CLI names present in this repository.
