# Contract

This is the stable contract between `yuiop` and any consumer — most importantly the
`qwert` dev-environment manager, which invokes `yuiop` as a subprocess and parses its
output. It is stable from v1: consumers must not depend on undocumented behavior.

## Output

By default, `yuiop` prints human-friendly output. Pass `--json` (a global flag) for
machine-readable output; `qwert` always uses `--json`.

- **stdout** — the result (list of packages, resolved info, status).
- **stderr** — diagnostics, errors, and the underlying package manager's noise.
- Successful operation: exit `0`; no result is written to stdout on success for verbs
  that are pure actions (`install`, `remove`, `upgrade`).

## JSON

Stable JSON shapes, one document per invocation:

**`status`**
```json
{ "platform": "brew", "package": "tmux", "installed": true }
```

**`list`**
```json
{ "platform": "brew", "packages": ["delta", "tmux"] }
```

**`search`**
```json
{ "platform": "brew", "term": "delta", "matches": ["delta", "git-delta"] }
```

**`info`**
```json
{ "platform": "brew", "package": "delta", "resolved": "git-delta", "installed": true }
```

**`platform`**
```json
{ "platform": "brew" }
```

Unknown keys are additive (consumers must tolerate new fields); removed/renamed fields
are breaking.

## Exit codes

| Code | Meaning |
|------|---------|
| `0` | success — including "already installed" (`install` is idempotent) |
| `1` | provider failure / generic error (the underlying manager failed) |
| `2` | invalid usage (unknown command, bad flag, missing argument) |
| `3` | not found — unknown canonical name, or a search with no hits |

## Name resolution

`yuiop` maps a canonical name to the package-per-manager using the embedded table
(`data/packages.yml`, curated in-repo). Examples:

```yaml
packages:
  delta:    { brew: git-delta,  apt: git-delta,  pacman: git-delta }
  opencode: { brew: anomalyco/tap/opencode }
```

If a canonical name is not in the table, `yuiop` responds:

```
yuiop: no knowledge of package 'foo'
```

…and exits `3`. It never guesses, never falls back to `install <name>` silently.

## Idempotency & privilege

- **Idempotent**: `install` checks `status` first; an already-installed package is
  success (exit `0`). Same for `remove` (already removed is success).
- **Privilege**: `apt` and `pacman` are invoked via `sudo` when not root; `brew` never
  uses sudo. `yuiop` **passes stdin/tty through** so password prompts work.
- **Non-interactive**: `install`/`remove`/`upgrade` pass `-y` / `--noconfirm`; `search`
  and `list` are textual.

## Invocation from qwert

`qwert` calls the binary for system-package recipes:

```sh
yuiop status <canonical> --json   # installed? → {"installed":true|false}
yuiop install <canonical> --json
yuiop remove  <canonical> --json
yuiop upgrade <canonical> --json
```

`qwert` passes the **canonical** name and lets `yuiop` resolve the per-manager package.
`qwert` no longer resolves names itself.