---
title: yuiop
status: proposed
created: 2026-09-04
updated: 2026-09-04
owner: "@oporpino"
certainty: high
---

# yuiop — the last package manager you need to know

> **TLDR**: `yuiop` é um wrapper universal sobre os demais package managers. Onde o `qwert` usa `brew`/`apt`/`pacman` na plataforma, o `yuiop` abstrai isso num único comando — e o qwert passa a usá-lo como padrão. `yuiop` + `qwert` = `QWERTYUIOP`.

## Contexto

O qwert hoje tem uma camada interna de resolução de package manager (`src/adapters/yuiop.rs` em `br4zz4/qwert`) que mapeia recipe → PM da plataforma. Queremos evoluir isso para um **projeto standalone** que:
- abstrai brew/apt/pacman (e futuramente cargo/npm/asdf etc.) sob uma interface comum;
- expõe operações `install`, `remove`, `list`, `search`, `upgrade`, `info`, `version`;
- resolve nomes de pacote por PM quando o mesmo pacote tem nomes diferentes (ex: `opencode` no Arch vs tap no brew);
- é usado pelo qwert como **camada padrão** de instalação.

## Objetivos

- Repo `br4zz4/yuiop` com binário `yuiop`(Rust, mesmo stack do qwert: clap + serde).
- Interface CLI: `yuiop install <pkg>`, `yuiop remove <pkg>`, `yuiop list`, `yuiop search <term>`, `yuiop upgrade <pkg|--all>`, `yuiop info <pkg>`, `yuiop version`.
- Resolução por plataforma: macOS → brew, Debian → apt, Arch → pacman; extensível por provedores.
- Suporte a nome de pacote por PM (`packages = { brew = "x", pacman = "y" }`) e a um override explícito de plataforma.
- Compatível com o schema de recipes do qwert (lê `install.toml`/manifestos equivalentes).
- O qwert passa a delegar instal/uninstall/upgrade para o `yuiop`.

## Fora de escopo (fase 1)

- Provedores extras além de brew/apt/pacman (cargo/npm/asdf, etc.) — fica aberto pelo trait `Provedor`.
- Registry público pesquisável.
- `yuiop` como daemon/serviço — é um binário CLI.

## Mudanças

### 1. `br4zz4/yuiop` — binário CLI

```
Usage: yuiop <COMMAND>

Commands:
  install   Instalar um pacote
  remove    Remover um pacote
  list      Listar pacotes instalados
  search    Buscar pacotes
  upgrade   Atualizar pacotes
  info      Detalhes de um pacote
  version   Versão do yuiop e do PM subjacente
```

- Auto-detecta a plataforma (mesma lógica do qwert: `cfg!(target_os)`, checar `/usr/bin/apt-get`, `/usr/bin/pacman`).
- Override explícito: `--platform macos|debian|arch` (e env `YUIOP_PLATFORM`).
- Saída sempre **texto estável/parsável** (uma linha por item) — o qwert chama `yuiop` via `Command` e parseia a saída. Padrão v1: texto simples; JSON fica para fase 2 para não adicionar serde_json já.

### 2. Trait `PackageManager`

```
trait PackageManager {
    fn install(&self, pkg: &str) -> Result<()>;
    fn remove(&self, pkg: &str) -> Result<()>;
    fn list(&self) -> Vec<String>;
    fn search(&self, term: &str) -> Vec<String>;
    fn upgrade(&self, pkg: Option<&str>) -> Result<()>;
    fn info(&self, pkg: &str) -> String;
    fn name_for(&self, canonical: &str) -> String;  // resolve alias por PM
}
```

Provedores: `Brew`, `Apt`, `Pacman`. Mesa de nome de pacote opcional por provedor:

```toml
[yuiop]
name = "tmux"
packages = { brew = "tmux", apt = "tmux", pacman = "tmux" }
```

### 3. Integração no qwert

`src/adapters/yuiop.rs` do qwert:
- Vira um front impl (ou é removido) que chama o binário `yuiop` externo (via `Command`) em vez de embutir os adapters.
- `default_adapter()`, `for_kind()` e o `install/uninstall/upgrade` de recipe delegam para o `yuiop`.
- O install.sh do qwert passa a também instalar o `yuiop` (binário em `~/.local/bin` ou `/opt/qwert/bin`).

## Como verificar

1. `cargo build --release` em `yuiop` gera `yuiop` funcional em macOS (brew), Debian (apt), Arch (pacman).
2. `yuiop install tmux` instala via PM da plataforma; `yuiop list` lista; `yuiop search realterm` acha; etc.
3. `yuiop --platform arch install fzf` respeita override.
4. qwert: `make t` continua verde; `qwert install tmux` num Linux chama `yuiop` e instala via apt/pacman.

## Documentação

- `README.md` do repo yuiop com uso e exemplo de provedores.
- `qwert/CLAUDE.md` e `README.md` atualizados para citar o yuiop como camada de PM.
- `.project/docs/specs/` deste repo guarda decisões futuras (provedores extras, saída JSON, registros).