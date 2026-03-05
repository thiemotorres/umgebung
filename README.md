# umgebung

A command-line tool for managing encrypted environment variable sets. Store secrets in an encrypted SQLite database and activate them in sub-shells on demand.

## Install

```
go install github.com/feto/umgebung@latest
```

## Commands

| Command | Description |
|---------|-------------|
| `umgebung init` | Initialize the encrypted database (`~/.config/umgebung/umgebung.db`) |
| `umgebung new <name>` | Create a new named env set |
| `umgebung edit <name>` | Edit an env set in your `$EDITOR` |
| `umgebung up [name]` | Activate an env set in a sub-shell (reads `.umgebung` if no name given) |
| `umgebung down` | Exit the active sub-shell |
| `umgebung place <name>` | Write a `.umgebung` file in the current directory |
| `umgebung import <name> <file>` | Import variables from a `.env` file into an env set |
| `umgebung export <name> <file>` | Export the env set to a `.env` file |

Run `umgebung` without arguments to open the interactive TUI browser.

## Example Workflow

```bash
# One-time setup
umgebung init

# Create an env set for a project
umgebung new myproject
# Opens $EDITOR — add KEY=VALUE lines, save and quit

# Pin the env set to a project directory
cd ~/projects/myapp
umgebung place myproject   # writes .umgebung

# Activate (reads .umgebung automatically)
umgebung up
# (myproject) $ echo $MY_SECRET   # variables are live in this shell
# (myproject) $ exit               # deactivates

# Or activate by name from anywhere
umgebung up myproject

# Import from an existing .env file
umgebung import myproject .env.production

# Export to a file
umgebung export myproject backup.env
```

## Security

- All variable values are encrypted with AES-256-GCM using a key derived from your master password via Argon2id.
- The database is stored at `~/.config/umgebung/umgebung.db` (or `$XDG_CONFIG_HOME/umgebung/umgebung.db`).
- Your password is cached for the current session by a background agent daemon (`umgebung agent`), so you only enter it once per login session.
- `.umgebung` files contain only the env set name (no secrets) and are safe to commit if desired. The binary and `*.db` files are excluded by the provided `.gitignore`.
- Secrets never touch disk in plaintext.
