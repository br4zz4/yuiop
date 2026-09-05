#!/usr/bin/env bash
# yuiop installer — the last package manager you need to know.
#
# Installs the yuiop binary into ~/.local/bin (no sudo). Prefers a prebuilt
# binary from GitHub Releases; falls back to `go install` for users with Go.
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
  go install "github.com/${REPO}@${gover}"
}

mkdir -p "${INSTALL_DIR}"
if ! install_from_release; then
  install_from_go
fi

msg "installed to ${INSTALL_DIR}/yuiop"
"${INSTALL_DIR}/yuiop" version