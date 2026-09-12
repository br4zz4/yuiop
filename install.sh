#!/usr/bin/env bash
# yuiop installer — the last package manager you need to know.
#
# Installs the yuiop binary into ~/.local/bin (no sudo). Prefers a prebuilt
# binary from GitHub Releases; falls back to `go install` for users with Go.
# Verifies the install dir is on PATH and prints instructions otherwise.
#
#   curl -fsSL https://raw.githubusercontent.com/br4zz4/yuiop/main/install.sh | bash
#
# Overrides:
#   YUIOP_VERSION      release tag to fetch (default: latest)
#   YUIOP_INSTALL_DIR  install directory (default: ~/.local/bin)

set -euo pipefail

VERSION="${YUIOP_VERSION:-latest}"
INSTALL_DIR="${YUIOP_INSTALL_DIR:-$HOME/.local/bin}"
REPO="br4zz4/yuiop"
API="https://api.github.com/repos/${REPO}"

msg() { printf '\033[1;34myuiop\033[0m: %s\n' "$*"; }
warn() { printf '\033[1;33myuiop\033[0m: %s\n' "$*"; }

resolve_latest() {
  # `latest` is not a tag whose asset name we know — resolve it to the real tag
  # so the download URL carries the versioned asset name (yuiop_<ver>_<os>_<arch>).
  curl -fsSL "${API}/releases/latest" | grep '"tag_name"' | sed 's/.*"tag_name": *"\(.*\)".*/\1/'
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64) echo amd64 ;;
    arm64|aarch64) echo arm64 ;;
    *) echo "" ;;
  esac
}

detect_os() {
  case "$(uname -s)" in
    Darwin) echo darwin ;;
    Linux) echo linux ;;
    *) echo "" ;;
  esac
}

install_from_release() {
  local os arch url
  os="$(detect_os)"
  arch="$(detect_arch)"
  if [[ -z "$os" || -z "$arch" ]]; then
    return 1
  fi
  if [[ "${VERSION}" == "latest" ]]; then
    VERSION="$(resolve_latest)" || return 1
  fi
  url="https://github.com/${REPO}/releases/download/${VERSION}/yuiop_${VERSION#v}_${os}_${arch}"
  msg "downloading ${url}"
  curl -fsSL "$url" -o "${INSTALL_DIR}/yuiop"
  chmod +x "${INSTALL_DIR}/yuiop"
}

install_from_go() {
  msg "no release binary available; using go install"
  command -v go >/dev/null 2>&1 || {
    printf 'yuiop: no release binary for this OS/arch and go is not installed.\n' >&2
    exit 1
  }
  local gover="${VERSION}"
  [[ "${gover}" == "latest" ]] && gover="latest"
  # The main package lives in cmd/yuiop — installing the repo root fails with
  # "package ... is not a main package".
  go install "github.com/${REPO}/cmd/yuiop@${gover}"
  # `go install` drops the binary in its own GOBIN (often not on PATH). Copy it
  # into INSTALL_DIR so `yuiop` is reachable regardless of the Go toolchain.
  local gobin gobin_yuiop
  gobin="$(go env GOBIN 2>/dev/null || echo "$HOME/go/bin")"
  gobin_yuiop="${gobin}/yuiop"
  if [[ -x "${gobin_yuiop}" && "${gobin_yuiop}" != "${INSTALL_DIR}/yuiop" ]]; then
    cp "${gobin_yuiop}" "${INSTALL_DIR}/yuiop"
    chmod +x "${INSTALL_DIR}/yuiop"
  fi
}

# shell_rc returns the user's rc file for the detected shell (or .profile).
shell_rc() {
  local sh
  sh="$(basename "${SHELL:-$0}")"
  case "${sh}" in
    zsh) echo "$HOME/.zshrc" ;;
    bash) echo "$HOME/.bashrc" ;;
    *) echo "$HOME/.profile" ;;
  esac
}

# ensure_on_path prints instructions when INSTALL_DIR is missing from PATH.
ensure_on_path() {
  local dir="${1}"
  case ":${PATH}:" in
    *":${dir}:"*) return 0 ;;
  esac
  local rc
  rc="$(shell_rc)"
  warn "${dir} is not on your PATH"
  warn "add it to ${rc}:"
  printf '  echo '\''export PATH="%s:$PATH"'\'' >> %s\n' "${dir}" "${rc}"
  printf '  source %s\n' "${rc}"
}

mkdir -p "${INSTALL_DIR}"
if ! install_from_release; then
  install_from_go
fi

msg "installed to ${INSTALL_DIR}/yuiop"
# Version check is informative only — don't abort the install if it fails
# (e.g. running with a stripped PATH).
"${INSTALL_DIR}/yuiop" version || true
ensure_on_path "${INSTALL_DIR}"