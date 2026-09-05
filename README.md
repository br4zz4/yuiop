# yuiop

> **yuiop — the last package manager you need to know.** `qwert` + `yuiop` = `QWERTYUIOP`.

`yuiop` is a universal wrapper over the platform package managers. It gives you one interface
to install, remove, list, search, upgrade and inspect packages — regardless of whether the
underlying manager is `brew`, `apt` or `pacman`.

The qwert dev-environment manager uses `yuiop` as its default package-management layer.

## Status

Spec being drafted. See `.project/docs/specs/20260904225500_yuiop_package_manager_wrapper.md`.

## Usage (planned)

```sh
yuiop install tmux
yuiop remove tmux
yuiop list
yuiop search delta
yuiop upgrade --all
yuiop platform            # → brew | apt | pacman
yuiop platform arch       # explicit override
```

## Design

- Rust, same stack as qwert: `clap` + `serde`.
- `PackageManager` trait with `Brew`, `Apt`, `Pacman` providers.
- Package aliases per manager (`packages = { brew = "...", pacman = "..." }`).
- Deterministic, parseable text output (the qwert invokes `yuiop` as a subprocess).

## License

TBD.