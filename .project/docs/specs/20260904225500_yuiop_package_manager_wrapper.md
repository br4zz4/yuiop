---
title: yuiop
status: decided
created: 2026-09-04
updated: 2026-09-04
owner: "@oporpino"
certainty: high
decisions:
  - q1: universal-first
  - q2: go
  - q4: config-yml-per-machine
  - q5: json-text-from-v1
  - q6: install-sh-and-go-install
  - q7: system-packages-only
  - q9: hand-maintained-embedded-table
  - q13: config-plain-platform-only
  - q14: no-cask-v1
---

# yuiop — the last package manager you need to know

> **TLDR**: `yuiop` é um wrapper fino e universal sobre os package managers de sistema
> (`brew`, `apt`, `pacman`). Ele **não é** um package manager: não tem estado de pacotes,
> não tem repositório, não resolve dependências — ele resolve o **nome**, delega a operação
> ao PM da plataforma e repassa o resto. Nasce do `qwert` como camada de PM, mas tem
> ambição de servir qualquer usuário: `qwert` + `yuiop` = `QWERTYUIOP`.

## Status

**Decided** — decisões de design tomadas via sessão de grill. Os *rationales* estão
na seção [Decisões](#decisões). Implementação em Go pendente (scaffold inicial).

## Contexto

O qwert hoje tem uma camada interna de resolução de package manager
(`src/adapters/yuiop.rs` e adapters `Brew`/`Apt`/`Pacman` em `br4zz4/qwert`) que mapeia
recipe → PM da plataforma. Queremos evoluir isso para um **projeto standalone**:

- abstrai brew/apt/pacman sob uma interface comum;
- expõe operações `install`, `remove`, `list`, `search`, `upgrade`, `info`, `status`;
- resolve nomes de pacote por PM quando o mesmo pacote tem nomes diferentes
  (ex: `delta` no brew é `git-delta`);
- é usado pelo qwert como **camada padrão** de instalação de pacotes de sistema.

Posicionamento: **universal-first**. O yuiop **nasce do qwert** (herda o catálogo curado,
o contrato de saída, as regras de detecção), mas o produto é pensado para qualquer pessoa
que queira um único comando para gerenciar pacotes de sistema — o qwert é o consumidor
flagship, não o único.

## Princípios

1. **É um wrapper, não um package manager.** Zero estado de pacotes próprio. Zero
   repositório/registry. Quem executa é o PM da plataforma.
2. **Resolve o nome, delega o resto.** O yuiop sabe responder *"canonical `delta` →
   brew `git-delta`, apt `git-delta`, pacman `git-delta`"*. Depois da resolução, ele
   "extrapola" o comando do PM (flags, privilégio) e passa o resto adiante.
3. **System packages apenas.** Linguagens/runtimes ficam com o `asdf`. Setups de
   shell/GUI/ferramentas customizadas ficam com o qwert (recipes `qwert`/`custom`).
4. **Por máquina.** A configuração do yuiop descreve a *máquina*, não o projeto.
5. **Contrato estável desde a v1.** Saída parseável (texto ou JSON), exit codes
   documentados. O qwert chama `yuiop` via subprocesso e parseia.
6. **Catálogo embutido + override zero.** A resolução usa a tabela embutida no binário.
   Não há aliases "de package" em config: se o nome não está na tabela, o yuiop diz
   claramente que não conhece aquele package.

## Objetivos

- Repo `br4zz4/yuiop` com binário `yuiop` em **Go** (stdlib-first, sem framework pesado).
- CLI: `install`, `remove`, `list`, `search`, `upgrade`, `info`, `status`, `platform`, `version`.
- Resolução por plataforma: macOS → brew, Debian → apt, Arch → pacman; extensível por
  provedores.
- Tabela de nomes de pacote por PM embutida (`data/packages.yml`, `go:embed`).
- Config por máquina **opcional** (`~/.config/yuiop/config.yml`), só `platform`.
- O qwert delega install/uninstall/upgrade de pacotes de sistema para o `yuiop`.

## Fora de escopo

- **Casks (brew) na v1** — yuiop só trata formulae; apps GUI seguem via recipes custom do qwert.
- **Provedores extra** (cargo/npm/pip/asdf) — linguagens e runtimes são domínio do asdf.
- **Windows.** Plataformas suportadas: macOS, Debian, Arch.
- **Registry público pesquisável** e **daemon/serviço** — é um binário CLI.
- **Self-upgrade** — `yuiop upgrade` atualiza *pacotes* via PM da plataforma, nunca a si
  mesmo; o upgrade do próprio binário pertence ao canal que o instalou (install.sh/go install).

## Decisões

| # | Decisão | Escolha |
|---|---------|---------|
| Q1 | Posicionamento | **(b) Universal-first** — nasce do qwert, produto para qualquer usuário |
| Q2 | Linguagem | **(b) Go** — binário estático por target, instalação trivial |
| Q4 | Configuração | **(a) XDG por máquina**, arquivo opcional |
| Q5 | Contrato de saída | **(a) JSON + texto humano desde a v1** |
| Q6 | Instalação | **(b) install.sh (GitHub content) + `go install`** — nunca via brew |
| Q7 | Escopo de provedores | **(a) brew/apt/pacman**, interface genérica aberta, system packages só |
| Q9 | Fonte da tabela | **(b) mantida à mão no repo** (`data/packages.yml`); desconhecido → erro claro |
| Q13 | Schema do config | Só `platform`. Sem aliases, sem seção `providers` |
| Q14 | Cask | **(c) sem cask na v1** |

### Rationale

- **Go sobre Rust**: o qwert consome o yuiop por **subprocesso** (não há código
  compartilhado), então "mesmo stack" não vale o custo. Go dá binário estático
  verdadeiro (`CGO_ENABLED=0`) para darwin/linux × amd64/arm64 sem toolchain cruzado,
  e `encoding/json` é stdlib — o contraste com a fase-2-JSON do spec original.
- **Config por máquina = XDG + só platform**: a config descreve a máquina. Aliases não
  fazem sentido porque a resolução é responsabilidade da **tabela embutida**; permitir
  override de nomes na config criaria uma segunda fonte de verdade. `platform` é a única
  coisa que a máquina pode querer fixar.
- **Sem cask na v1**: casks são apps GUI; no modelo "system packages", fórmulas CLI são o
  núcleo. Apps GUI continuam via recipes custom do qwert (`brew install --cask`),
  o que mantém o yuiop simples e honesto.
- **install.sh + go install, não brew**: usar brew para instalar o próprio wrapper de brew
  é cíclico e estranho. `install.sh` baixa o binário de GitHub content; `go install` serve
  quem já usa Go. O `self install` do qwert passa a instalar o yuiop junto.
- **Tabela embutida, não gerada**: uma geração automática de `qwert-recipes` na CI
  introduz acoplamento de build e corrida de catálogo. A tabela é pequena, curada por
  humanos e o mecanismo de override no código é simples. O custo de manutenção é
  aceitável para os ~dezenas de casos relevantes.

## Mudanças

### 1. `br4zz4/yuiop` — binário CLI em Go

Estrutura de referência:

```
cmd/yuiop/main.go              ← entrypoint fino
embed.go                       ← go:embed da tabela de pacotes
internal/cli/                  ← subcomandos, help, exit codes
internal/platform/             ← detecção + precedência do override
internal/config/               ← config.yml por máquina (só platform)
internal/resolve/              ← tabela de nomes (go:embed data/packages.yml)
internal/provider/             ← interface Provider + registro
internal/provider/brew.go      ← provedor brew
internal/provider/apt.go       ← provedor apt
internal/provider/pacman.go    ← provedor pacman
data/packages.yml              ← tabela curada, embutida no build
```

Verbos (v1):

```
yuiop install <pkg>            # instala via PM da plataforma
yuiop remove <pkg>
yuiop list
yuiop search <term>
yuiop upgrade <pkg> | --all
yuiop info <pkg>
yuiop status <pkg>             # instalado? (contrato para o qwert)
yuiop platform [<platform>]    # mostra/fixa o override
yuiop version
```

Flags globais: `--json`, `--platform <brew|apt|pacman>`, `--config <path>`.

Precedência da plataforma: `--platform` > `YUIOP_PLATFORM` > `config.yml:platform` >
autodetect (macOS → brew; `/usr/bin/apt-get` → apt; `/usr/bin/pacman` → pacman).

### 2. Interface `Provider`

```go
type Provider interface {
    Name() string
    // Installed reports whether pkg is installed.
    Installed(pkg string) (bool, error)
    // Run executes an operation, attaching stdin/tty so sudo prompts work.
    Install(pkg string) error
    Remove(pkg string) error
    Upgrade(pkg string) error
    UpgradeAll() error
    List() ([]string, error)
    Search(term string) ([]string, error)
}
```

Provedores: `Brew` (formulae; sem sudo, sem cask), `Apt` (`sudo apt-get ... -y`), `Pacman`
(`sudo pacman ... --noconfirm`). O binário **repassa stdin/tty** para prompts de senha
funcionarem. `install` é **idempotente**: checa `status` antes, e "já instalado" é sucesso
(exit 0).

### 3. Resolução de nomes

`data/packages.yml` (tabela curada, embutida via `go:embed`):

```yaml
packages:
  delta: { brew: git-delta, apt: git-delta, pacman: git-delta }
  opencode: { brew: anomalyco/tap/opencode }
```

Resolução: canonical → nome do PM. Se o canonical não está na tabela, o yuiop responde de
forma clara:

```
yuiop: no knowledge of package 'foo'
```

Sugestão de extensão de tabela: `add`/PR do próprio repo — a curadoria é comunitária e por
máquina se resolve com o `config.yml` só para plataforma.

### 4. Configuração por máquina

`~/.config/yuiop/config.yml` (ou `$YUIOP_CONFIG`), **opcional** — ausente/vazio =>
autodetect + PM padrão:

```yaml
platform: macos          # brew | apt | pacman — fixa a plataforma por máquina
```

Sem aliases (resolução é da tabela embutida), sem seção `providers`. `yuiop platform macos`
grava neste arquivo.

### 5. Contrato de saída

Texto humano por padrão; `--json` para máquinas (o qwert sempre usa `--json`).

Exit codes (documentados em `docs/CONTRACT.md`):

| Código | Significado |
|--------|-------------|
| 0 | sucesso (inclui "já instalado") |
| 1 | falha do provedor / erro genérico |
| 2 | uso inválido |
| 3 | não encontrado (canonical desconhecido ou search sem hits) |

### 6. Instalação

- `install.sh` — baixa o binário estático (GitHub content) para `~/.local/bin/yuiop`, sem sudo.
- `go install github.com/br4zz4/yuiop@latest` — para quem já usa Go.
- **Não** via brew (evita o ciclo brew-instala-brew).
- O `self install` do qwert passa a instalar o yuiop no mesmo canal.

### 7. Integração no qwert

- `src/adapters/` do qwert: os recipes *package-kind* passam a delegar para o binário
  `yuiop <verb> <canonical> --json` via `Command`.
- qwert não resolve mais nomes: chama o canonical e o yuiop resolve. Recipes `qwert/custom`
  (incluindo casks e setups GUI) continuam no qwert.
- `CLAUDE.md` → `AGENTS.md` no repo qwert (documentação para agentes), com nota da camada yuiop.

## Como verificar

1. `go build ./... && go vet ./...` em `yuiop`; `yuiop version` e `yuiop platform` funcionam.
2. `yuiop install tmux` instala via PM da plataforma; `yuiop list` lista; `yuiop search delta` acha.
3. `yuiop install nonexistent-package` → saída clara + exit 3.
4. `yuiop --platform arch install fzf` respeita override; `yuiop platform` persiste no config.yml.
5. `yuiop install <pkg>` de novo → exit 0 (idempotente).
6. qwert: `make t` continua verde; `qwert install tmux` num Linux chama `yuiop` e instala via apt/pacman.

## Documentação

- `README.md` — identidade "wrapper, não package manager"; quickstart; fronteira de domínio (yuiop/asdf/qwert).
- `docs/CONFIG.md` — config.yml, precedência, paths.
- `docs/CONTRACT.md` — saída texto/JSON, exit codes (contrato com o qwert).
- `docs/PROVIDERS.md` — brew/apt/pacman: comandos, flags, privilégio.
- `qwert/CLAUDE.md`(`AGENTS.md`) e `README.md` atualizados para citar o yuiop como camada de PM.
- Este repo guarda as decisões (ADR via `decisions` no frontmatter).