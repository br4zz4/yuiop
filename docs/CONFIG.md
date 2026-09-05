# Configuration

`yuiop` reads a single, **optional** config file that describes the *machine* — not the
project. With no config file present, everything is auto-detected.

## Location

| Source | Path |
|--------|------|
| Env var | `$YUIOP_CONFIG` |
| XDG | `$XDG_CONFIG_HOME/yuiop/config.yml` → default `~/.config/yuiop/config.yml` |

If the file is absent or empty, `yuiop` auto-detects the platform and selects the default
package manager.

## Schema (v1)

```yaml
platform: macos   # brew | apt | pacman
```

`platform` is the only key in v1. It pins which package manager this machine uses,
overriding auto-detection. It is set by `yuiop platform <brew|apt|pacman>`.

There are deliberately **no aliases and no `providers` section**. Name resolution
(canonical → package-per-manager) is the job of the embedded table — see
[`docs/CONTRACT.md`](CONTRACT.md#name-resolution) — so a config override would create a
second source of truth.

## Platform precedence

The effective platform is chosen in this order (first that yields a value wins):

1. `--platform <brew|apt|pacman>` (CLI flag)
2. `YUIOP_PLATFORM` (env var)
3. `platform:` in `config.yml`
4. Auto-detect: macOS → `brew`; `/usr/bin/apt-get` → `apt`; `/usr/bin/pacman` → `pacman`

```sh
yuiop --platform arch install fzf   # one-shot override
YUIOP_PLATFORM=apt yuiop install fzf
yuiop platform macos                # persist to config.yml
yuiop platform                      # print the effective platform
```

## Unknown platform

If `yuiop` cannot determine a supported platform (e.g. Windows, or an unknown Linux), it
reports a clear error with the supported platforms and exits non-zero. There is no
silent fallback.