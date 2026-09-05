# yuiop

> **yuiop — the last package manager you need to know.** `qwert` + `yuiop` = `QWERTYUIOP`.

`yuiop` is a thin, universal wrapper over the **system** package managers (`brew`, `apt`,
`pacman`). One interface to install, remove, list, search, upgrade and inspect packages —
regardless of which manager your machine runs.

> **yuiop is not a package manager.** It has no package state, no repository, and
> resolves no dependencies. It resolves the *name* (canonical → package-per-manager),
> delegates the operation to the platform's package manager, and gets out of the way.

`yuiop` was born from [`qwert`](https://github.com/br4zz4/qwert), the dev-environment
manager, which uses it as its default system package-management layer. But it is designed
for anyone — qwert is the flagship consumer, not the only one.

## The domain boundary

| Domain | Tool |
|--------|------|
| System packages (`tmux`, `delta`, `opencode`) | **yuiop** |
| Languages & runtimes (`node`, `python`, `ruby`) | [`asdf`](https://asdf-vm.com) |
| Shell config, GUI apps, custom setups | `qwert` (custom recipes) |

## Install

```sh
# Binary into ~/.local/bin (no sudo)
curl -fsSL https://raw.githubusercontent.com/br4zz4/yuiop/main/install.sh | bash

# ...or, if you already use Go
go install github.com/br4zz4/yuiop@latest
```

> `yuiop upgrade` upgrades *packages*, never itself. Updating the `yuiop` binary belongs
> to the channel that installed it.

## Usage

```sh
yuiop install tmux            # brew install tmux / sudo apt install tmux / sudo pacman -S tmux
yuiop remove tmux
yuiop list
yuiop search delta
yuiop upgrade --all
yuiop info delta
yuiop status tmux             # installed? (exit 0/1) — the machine contract
yuiop platform                # → brew | apt | pacman
yuiop version
```

`yuiop` resolves canonical names through an embedded, curated table — `delta` becomes
`brew install git-delta` on macOS. If it doesn't know a package, it says so:

```sh
$ yuiop install foo
yuiop: no knowledge of package 'foo'   # exit 3
```

## Configuration

Configuration is **per machine** and **optional**. With no config present, `yuiop`
auto-detects the platform and selects its default package manager.

```yaml
# ~/.config/yuiop/config.yml
platform: macos   # brew | apt | pacman — pin the platform for this machine
```

See [`docs/CONFIG.md`](docs/CONFIG.md) for precedence and paths,
[`docs/CONTRACT.md`](docs/CONTRACT.md) for output format and exit codes, and
[`docs/PROVIDERS.md`](docs/PROVIDERS.md) for how each manager is invoked.

## Status

Initial design decided — see the [spec](.project/docs/specs/20260904225500_yuiop_package_manager_wrapper.md).
Go implementation in progress.

## License

TBD.