# 🔐 GitHub Authentication Manager (GHAM)

**GHAM** is a lightweight, secure CLI utility that enables seamless management of multiple GitHub authentication contexts. It simplifies switching between personal, professional, or client-based GitHub accounts — without needing to reconfigure credentials manually.

---

## ✨ Features

- **Multi-Context Management**  
  Define and store multiple GitHub account contexts (name, username, email, and PAT).

- **Secure PAT Storage**  
  Uses OS-specific keychains via [`go-keyring`](https://github.com/zalando/go-keyring):  
  - macOS: Keychain  
  - Windows: Credential Manager  
  - Linux: Secret Service (`libsecret`)

- **Repository-Specific Context Binding**  
  Assign a context to a Git repository, so GHAM auto-applies the correct Git identity.

- **Git Wrapper Integration**  
  Use `gham git ...` to automatically inject credentials, username, and email.

---

## 📦 Installation

### 1. Prerequisites

- Go (version **1.20+** recommended)
- A C compiler (`gcc`/`clang`) for CGO dependencies
- A working keyring service:
  - macOS: Keychain Access
  - Linux: `libsecret`
  - Windows: Credential Manager

### 2. Build from Source

```bash
git clone https://github.com/<your-github-username>/gham.git
cd gham

# Build with version info (recommended)
go build -ldflags="-X main.AppVersion=v1.0.0" -o gham .

# Install to PATH (choose one):
sudo mv gham /usr/local/bin/

# OR user-local:
mkdir -p ~/bin && mv gham ~/bin/
export PATH="$HOME/bin:$PATH" # Add to your shell config if needed
```


### 🚀 Usage

```bash
# Get CLI help
gham --help
gham context --help
gham repo --help
gham git --help

# 1. Add a personal GitHub context
gham context add personal --email "me@example.com" --username "myusername"

# 2. Add a work GitHub context
gham context add work --token "ghp_xxx" --email "me@work.com" --username "workusername"

# 3. List configured contexts
gham context list

# 4. Navigate to your Git repository
cd ~/projects/my-repo

# 5. Assign a context to this repo
gham repo assign work

# 6. Check assigned context
gham repo current

# 7. Use GHAM for Git commands with injected credentials
gham git status
gham git pull
gham git commit -m "Updated with GHAM"
gham git push

# 8. Remove a context
gham context remove personal

# 9. Show version
gham version
```



### 💡 Example Workflow

```bash
cd ~/projects/client-x
gham repo assign client-x

# From now on, this repo uses the client-x GitHub identity:
gham git push
```

### 🤝 Contributing
We welcome contributions from the open-source community! Here's how you can help:

⭐ Star this repo to show support
🐛 Report issues and bugs
🛠 Submit pull requests for fixes or new features
📄 Help improve the documentation

## Get Started:
```bash
# Fork the repo
# Clone your fork
git clone https://github.com/<your-username>/gham.git
cd gham

# Create a branch and start hacking!
git checkout -b feature/something-cool
```
Please follow conventional commit guidelines and write tests for new features where appropriate.


### 🙌 Acknowledgements
- go-keyring – Native keyring integration
- Inspired by the need for seamless GitHub identity switching


Built with ❤️ by developers for developers.
```bash
Let me know if you'd like to include badges (build status, license, Go report card), add a `CONTRIBUTING.md` file, or turn this into a template for your GitHub repo automatically.
```