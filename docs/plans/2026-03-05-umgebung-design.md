# umgebung - Design Document

Date: 2026-03-05

## Overview

`umgebung` is a CLI/TUI tool for managing environment variable sets (EnvSets) per project. Instead of storing secrets in plaintext files (like direnv), it stores them encrypted in a local SQLite database. A password is required once per terminal session, cached by a background agent daemon.

## Architecture

Three components:

1. **CLI** (`umgebung`) - main binary, all commands via cobra
2. **Agent Daemon** (`umgebung-agent`) - background process that holds the derived encryption key in RAM; communicates via Unix Socket at `/tmp/umgebung-<uid>/agent.sock`; auto-forked by the CLI on first use; times out after 15 minutes of inactivity
3. **SQLite DB** (`~/.config/umgebung/umgebung.db`) - stores EnvSets with AES-256-GCM encrypted values

## Data Model

```sql
env_sets (id INTEGER PRIMARY KEY, name TEXT UNIQUE, created_at DATETIME, updated_at DATETIME)
env_vars (id INTEGER PRIMARY KEY, env_set_id INTEGER, key TEXT, value BLOB, FOREIGN KEY(env_set_id) REFERENCES env_sets(id))
```

Key names are stored in plaintext (useful for TUI preview). Values are encrypted blobs. A random salt is stored unencrypted in the DB metadata table for Argon2id key derivation.

## Encryption Flow

```
User Password + Salt → Argon2id → 32-byte Key → AES-256-GCM (per value)
```

- Salt: random 16 bytes, generated at `umgebung init`, stored in DB
- Nonce: random 12 bytes, prepended to each encrypted value blob
- The agent daemon holds the derived 32-byte key in memory after first unlock

## Agent Daemon (IPC)

- CLI checks for `/tmp/umgebung-<uid>/agent.sock`
- If socket missing: auto-fork daemon, wait for ready signal
- If socket present: send unlock request or key retrieval request
- Daemon resets inactivity timer on each request
- Daemon exits after 15 min inactivity (key gone from RAM)

Protocol: simple JSON over Unix socket
```json
// Request
{"action": "unlock", "password": "..."}
{"action": "get_key"}

// Response
{"ok": true}
{"ok": false, "error": "wrong password"}
{"key": "<base64>"}
```

## TUI Layout

Triggered by `umgebung` with no arguments (after init).

```
+- EnvSets ----------+- Preview: PRODUCTION -------------------+
| > PRODUCTION       |  DATABASE_URL   = ******************   |
|   STAGING          |  API_KEY        = ******************   |
|   LOCAL            |  DEBUG          = ******************   |
|                    |                                        |
| [n]ew  [d]elete    |  [e]dit  [u]p  [enter] activate        |
+--------------------+----------------------------------------+
```

- Left pane: list of EnvSets, arrow key navigation
- Right pane: preview of keys (values masked with `*`)
- Values are only decrypted for display on explicit reveal (`r` key) or activation
- All CRUD operations accessible via keyboard shortcuts

## Sub-Shell Activation

`umgebung up XXXX`:
1. Retrieves key from agent (prompts password if daemon not unlocked)
2. Decrypts all values for EnvSet XXXX
3. Spawns `$SHELL` with env vars injected
4. Sets `UMGEBUNG_ACTIVE=XXXX` in the sub-shell env
5. Modifies `PS1` to prepend `(XXXX)` to the prompt

`umgebung down` prints a hint that the user should `exit` the current shell (since a child process cannot exit a parent shell). Alternatively it can be aliased.

`umgebung up` (no args): reads `.umgebung` file in CWD, uses its content as the EnvSet name.

## Commands

| Command | Action |
|---|---|
| `umgebung` | Open TUI if initialized, otherwise redirect to init |
| `umgebung init` | Create DB, set master password |
| `umgebung new XXXX` | Create new EnvSet, open $EDITOR for key=value pairs |
| `umgebung edit XXXX` | Edit existing EnvSet in $EDITOR |
| `umgebung up XXXX` | Activate EnvSet XXXX in a sub-shell |
| `umgebung up` | Activate EnvSet from `.umgebung` file in CWD |
| `umgebung down` | Print instructions to exit active sub-shell |
| `umgebung place XXXX` | Write `.umgebung` file with XXXX into CWD |
| `umgebung import XXXX file.env` | Parse .env file and create EnvSet XXXX |
| `umgebung export XXXX file.env` | Write EnvSet XXXX to plaintext .env file |

## Tech Stack

- **Go** - language
- **cobra** - CLI framework
- **bubbletea** + **lipgloss** - TUI framework and styling
- **modernc.org/sqlite** - pure-Go SQLite driver (no CGo)
- **golang.org/x/crypto** - Argon2id + AES-256-GCM
- Unix Domain Sockets - agent IPC

## File Layout

```
~/.config/umgebung/
  umgebung.db       # encrypted SQLite database

/tmp/umgebung-<uid>/
  agent.sock        # Unix socket for daemon IPC

.umgebung           # project file (plaintext EnvSet name, committed or gitignored per preference)
```

## Security Notes

- Master password is never stored anywhere
- Derived key lives only in agent daemon RAM
- `.umgebung` files contain only the EnvSet name (no secrets)
- `export` command produces plaintext - user is warned before writing
- DB file should be chmod 600
