# Providers

`yuiop` is a thin wrapper over the platform's system package manager. Each provider knows
how to detect itself, whether a package is installed, and how to run each operation. All
other concerns (flags, privilege, prompts) are the provider's job.

| Provider | Platform | Installed check | Install | Remove | Upgrade | Search/List |
|----------|----------|-----------------|---------|--------|---------|-------------|
| `brew`   | macOS    | `brew list <pkg>` | `brew install <pkg>` | `brew uninstall <pkg>` | `brew upgrade <pkg>` | `brew search` / `brew list` |
| `apt`    | Debian   | `dpkg -s <pkg>` | `sudo apt-get install -y <pkg>` | `sudo apt-get remove -y <pkg>` | `sudo apt-get install --only-upgrade -y <pkg>` | `apt-cache search` / `dpkg -l` |
| `pacman` | Arch     | `pacman -Q <pkg>` | `sudo pacman -S --noconfirm <pkg>` | `sudo pacman -R --noconfirm <pkg>` | `sudo pacman -Su --noconfirm` | `pacman -Ss` / `pacman -Q` |

## `brew`

- Detected on macOS; also gated on `brew` being on `PATH`.
- **Formulae only** in v1 — no casks. GUI/desktop apps stay with `qwert` custom recipes.
- Never uses `sudo`. Homebrew installs to its own prefix.
- `upgrade` of a specific package: `brew upgrade <pkg>`. `upgrade --all`: `brew upgrade`.

## `apt`

- Detected by `/usr/bin/apt-get`.
- Uses `sudo` when not root; `yuiop` passes tty through so the password prompt works.
- `-y` for non-interactive install/remove/upgrade.
- `upgrade` of a specific package uses `apt-get install --only-upgrade -y <pkg>`.

## `pacman`

- Detected by `/usr/bin/pacman`.
- Uses `sudo` when not root.
- `--noconfirm` for non-interactive operations.
- `upgrade --all`: `pacman -Su --noconfirm`.

## Provider interface

New providers implement a single interface (see the spec): `Name`, `Installed(pkg)`,
`Install(pkg)`, `Remove(pkg)`, `Upgrade(pkg)`, `UpgradeAll()`, `List()`, `Search(term)`.
The registry picks the provider from the resolved platform.