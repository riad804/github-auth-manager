Here is a clean, professional **README.md** file for your `git-auth-manager` project.  
It explains the purpose, installation, usage, features, and important notes in a way that's suitable for GitHub or any open-source repo.

```markdown
# git-auth-manager

**Manage multiple Git accounts (GitHub, GitLab, Bitbucket) per repository — without SSH keys and without conflicts.**

A secure, cross-platform CLI tool that makes Git authentication **repository-scoped** using HTTPS + OAuth tokens.

Each repository can use its own account automatically via a custom Git credential helper — no more manual `git config user.name/email` switching or global token conflicts.

## Features

- Repository-scoped authentication (each repo uses its own account)
- Secure OAuth login (browser-based, PKCE, 127.0.0.1 callback)
- Tokens stored in OS keychain / credential manager (never plaintext)
- Supports **GitHub**, **GitLab**, **Bitbucket**
- Custom Git credential helper (`!git-auth-manager credential`)
- Commands: `login`, `logout`, `link`, `unlink`, `accounts`, `status`, `doctor`
- Automatic token refresh (non-blocking when possible)
- Cross-platform: macOS, Linux, Windows
- Works with existing HTTPS remotes — no SSH required
- Safe integration: chains with existing credential helpers

## Installation

### Option 1: Build from source (recommended for now)

```bash
# Clone the repo
git clone https://github.com/yourusername/git-auth-manager.git
cd git-auth-manager

# Build for your current platform
go build -o git-auth-manager ./cmd/git-auth-manager

# Or universal macOS binary (Intel + Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o git-auth-manager-arm64 ./cmd/git-auth-manager
GOOS=darwin GOARCH=amd64 go build -o git-auth-manager-amd64 ./cmd/git-auth-manager
lipo -create -output git-auth-manager git-auth-manager-arm64 git-auth-manager-amd64

# Move to PATH
chmod +x git-auth-manager
mv git-auth-manager ~/bin/   # or /usr/local/bin/
```

### Option 2: Using Goreleaser (for release builds)

```bash
goreleaser build --snapshot --clean
```

Look in `dist/` for the binary.

### Setup Git credential helper (one-time)

Run this once after installation:

```bash
git-auth-manager doctor
```

Or manually:

```bash
git config --global credential.helper 'cache --timeout=900'
git config --global --add credential.helper '!git-auth-manager credential'
```

> Recommended: add `cache` helper first for speed.

## Quick Start

1. **Login to an account**

```bash
git-auth-manager login --provider github --id work
# or
git-auth-manager login --provider gitlab --id personal
```

→ Opens browser → login → token securely stored

2. **List your accounts**

```bash
git-auth-manager accounts
```

3. **Link a repository to an account**

```bash
cd /path/to/your/repo
git-auth-manager link
```

→ Selects from your accounts (or use `--id work`)

4. **Use Git normally**

```bash
git pull
git push
git fetch
```

→ Automatically uses the correct token for this repo

5. **Check status**

```bash
git-auth-manager status
```

6. **Unlink repo**

```bash
git-auth-manager unlink
```

7. **Logout**

```bash
git-auth-manager logout --id work
```

## All Commands

| Command               | Description                                      |
|-----------------------|--------------------------------------------------|
| `login`               | OAuth login to GitHub/GitLab/Bitbucket           |
| `logout --id <id>`    | Remove account and token                         |
| `accounts`            | List all logged-in accounts                      |
| `link [--id <id>]`    | Link current repo to an account                  |
| `unlink`              | Remove account link from current repo            |
| `status`              | Show which account is linked to current repo     |
| `doctor`              | Check setup & configure credential helper        |
| `credential`          | (Internal) Git credential helper                 |

## How It Works (High-Level)

1. You login once per account → token stored in OS secure storage (Keychain / Credential Manager / libsecret)
2. You link a repo → `.git/git-auth-manager.json` stores `account_id`
3. Git calls the credential helper → helper reads repo mapping → fetches correct token → provides username/password to Git
4. Token refresh happens automatically (async when near expiry)

Supported username/password formats:

- **GitHub**: `username = <token>`, `password = ""`
- **GitLab**: `username = oauth2`, `password = <token>`
- **Bitbucket**: `username = x-token-auth`, `password = <token>`

## Important Notes

- **You must register your own OAuth app** for each provider and put real `client_id` values in `internal/auth/providers.go`
  - GitHub: https://github.com/settings/apps
  - GitLab: https://gitlab.com/-/profile/applications
  - Bitbucket: https://bitbucket.org/account/settings/apps/
  - Redirect URI: `http://127.0.0.1:1236/callback`

- Tokens are **never** stored in plaintext — only in OS-native secure storage

- The tool **does not** override existing global credential helpers (uses `--add`)

- Performance: For best speed, use the recommended `cache` + custom helper combo

## Contributing

Contributions welcome!

1. Fork & branch
2. Make changes
3. `go test ./...`
4. Submit PR

## License

MIT

```

Feel free to copy-paste this directly into your `README.md`.

### Suggestions for next steps

- Replace `yourusername` with your actual GitHub username
- Add real screenshots (login flow, status command, etc.)
- Add a "Supported Providers" section with icons
- Include a small "Troubleshooting" section later

Let me know if you want to add sections like:

- Roadmap
- Security considerations
- Comparison with other tools (gh auth, git-credential-manager, etc.)
- Homebrew tap instructions

Happy to refine it further!