<div align="center">

# 💡 explain

**Demystify Linux terminal commands in seconds.**  
*A lightweight, zero-dependency, 100% offline CLI tool written in Go that breaks down complex terminal commands, explains their flags, and highlights safety risks before you run them.*

[![Go Report Card](https://goreportcard.com/badge/github.com/elsawy/explain)](https://goreportcard.com/report/github.com/elsawy/explain)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS-lightgrey.svg)]()
[![Release](https://img.shields.io/github/v/release/elsawy/explain?color=green)](https://github.com/elsawy/explain/releases)

<br/>

<img src="assets/demo.svg" alt="explain demo preview" width="850"/>

</div>

---

## 🎯 Why `explain`?

When beginners copy commands from ChatGPT, StackOverflow, or tutorials, they often don't know what cryptic flags (`-xzvf`, `-laht`, `-p 8080:80`, `| sh`) actually do.

- Reading full `man` pages is overwhelming.
- Asking online AI models requires internet access, API keys, and context switching.
- Running commands blindly can lead to accidental data loss (`rm -rf`, `> file`, `dd`).

With `explain`, you just prepend `explain` before any command:

```bash
explain tar -xzf backup.tar.gz
```

---

## ✨ Features

- **⚡ Zero Dependencies & 100% Offline**: Single static binary (<5MB). Instant execution (<5ms) with no network calls or API keys required.
- **🧩 Smart Flag Decomposition**: Unpacks clustered short flags (e.g., `-xzvf` $\to$ `-x`, `-z`, `-v`, `-f`) and maps them to their values.
- **💬 Shell Builtins & Navigation**: Understands `cd`, `pwd`, `echo`, `export`, `alias`, `clear`, and path transitions (`..`, `~`, `-`).
- **🔗 Pipeline & Redirect Aware**: Understands multi-stage pipelines (`ps aux | grep nginx | awk ...`) and I/O redirections (`>`, `>>`, `2>&1`).
- **🛡️ Safety & Danger Meter**: Warns against destructive operations (`rm -rf /`, `dd of=/dev/sdX`, `chmod 777`, `curl ... | bash`).
- **📖 Dynamic Man & Help Fallback**: Real-time extraction from local `man` pages and `--help` for any unlisted command installed on your system.
- **🎨 Modern & Compact Terminal UI**: Beautiful ANSI colors and aligned columns designed to give you everything in 5–8 clean lines.
- **🚀 Interactive Safe Runner**: Use `explain -i` to paste complex pipelines without quotes, or `explain -r "<command>"` to execute with confirmation.

---

## 🛡️ Safety & Hazard Detection

<div align="center">
  <img src="assets/danger.svg" alt="explain safety hazard demo" width="850"/>
</div>

---

## 🚀 Installation

### Option 1: One-Line Installer (Recommended)
```bash
curl -fsSL https://raw.githubusercontent.com/elsawy/explain/main/scripts/install.sh | bash
```

### Option 2: Go Install
```bash
go install github.com/elsawy/explain/cmd/explain@latest
```

### Option 3: Build from Source
```bash
git clone https://github.com/elsawy/explain.git
cd explain
make
make install
```

---

## 📖 CLI Usage

```text
USAGE:
  explain <command with arguments>
  explain "<piped | or compound command>"
  explain -i (interactive mode - no quotes needed)

OPTIONS:
  -i, --interactive Launch interactive mode (paste complex pipelines without quotes)
  -r, --run         Ask to run the command after explaining it
  --json            Output structured analysis in JSON format
  --no-color        Disable colored output
  -v, --version     Show explain version
  -h, --help        Show help message
```

---

## 💡 Shell Integration (Bonus)

Add this alias to your `~/.bashrc` or `~/.zshrc` to quickly explain your last executed terminal command:

```bash
alias explain-last='explain "!!" '
```

---

## 🤝 Contributing

Contributions are welcome! To add support for more commands or flags, feel free to open a Pull Request.

```bash
# Run tests
make test

# Build local binary
make build
```

---

## 📄 License

Distributed under the **MIT License**. See [LICENSE](LICENSE) for details.
