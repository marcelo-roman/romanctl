#!/bin/sh
# Instala o romanctl: curl -fsSL <url>/install.sh | sh
# Variáveis: ROMANCTL_VERSION (padrão: a última), ROMANCTL_INSTALL_DIR, ROMANCTL_REPO, ROMANCTL_BASE_URL.
set -eu

REPO="${ROMANCTL_REPO:-marcelo-roman/romanctl}"
VERSION="${ROMANCTL_VERSION:-latest}"

say() { printf '%s\n' "$*"; }
fail() { printf 'romanctl: %s\n' "$*" >&2; exit 1; }

case "$(uname -s)" in
  Linux) os=linux ;;
  Darwin) os=darwin ;;
  *) fail "sistema não suportado: $(uname -s). O romanctl roda em Linux e macOS." ;;
esac
case "$(uname -m)" in
  x86_64 | amd64) arch=amd64 ;;
  aarch64 | arm64) arch=arm64 ;;
  *) fail "arquitetura não suportada: $(uname -m). O romanctl roda em amd64 e arm64." ;;
esac

if [ -n "${ROMANCTL_BASE_URL:-}" ]; then
  base="$ROMANCTL_BASE_URL"
elif [ "$VERSION" = latest ]; then
  base="https://github.com/$REPO/releases/latest/download"
else
  base="https://github.com/$REPO/releases/download/$VERSION"
fi

if [ -n "${ROMANCTL_INSTALL_DIR:-}" ]; then
  dir="$ROMANCTL_INSTALL_DIR"
elif [ "$(id -u)" = 0 ]; then
  dir=/usr/local/bin
else
  dir="$HOME/.local/bin"
fi

fetch() {
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$1" -o "$2"
  elif command -v wget >/dev/null 2>&1; then
    wget -q "$1" -O "$2"
  else
    fail "precisa de curl ou wget."
  fi
}

sha256() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | cut -d' ' -f1
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$1" | cut -d' ' -f1
  else
    fail "precisa de sha256sum ou shasum para conferir o download."
  fi
}

asset="romanctl-$os-$arch"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT INT TERM

say "Baixando $asset ($VERSION)..."
fetch "$base/$asset" "$tmp/$asset" || fail "não consegui baixar $base/$asset"
fetch "$base/SHA256SUMS" "$tmp/SHA256SUMS" || fail "não consegui baixar a lista de checksums"

expected="$(awk -v f="$asset" '$2 == f || $2 == "*" f { print $1 }' "$tmp/SHA256SUMS")"
[ -n "$expected" ] || fail "$asset não aparece em SHA256SUMS"
actual="$(sha256 "$tmp/$asset")"
[ "$expected" = "$actual" ] || fail "checksum não confere (esperado $expected, obtido $actual). Nada foi instalado."
say "Checksum SHA-256 conferido."

mkdir -p "$dir"
chmod 755 "$tmp/$asset"
mv "$tmp/$asset" "$dir/romanctl"
say "Instalado em $dir/romanctl"
"$dir/romanctl" version
"$dir/romanctl" templates update >/dev/null 2>&1 || say "Os templates ficam para depois: romanctl templates update"

case ":$PATH:" in
  *":$dir:"*) ;;
  *) say ""
     say "Adicione $dir ao PATH do seu shell, por exemplo:"
     say "  echo 'export PATH=\"$dir:\$PATH\"' >> ~/.bashrc" ;;
esac
say "Para começar: romanctl login"
