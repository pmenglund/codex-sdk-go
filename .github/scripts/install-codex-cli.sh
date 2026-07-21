#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "usage: install-codex-cli.sh --destination DIR [--target TARGET] [--archive FILE] [--metadata-dir DIR]" >&2
}

script_dir="$(dirname "$(realpath "${BASH_SOURCE[0]}")")"
metadata_dir="$(realpath "${script_dir}/../codex")"
destination=""
target=""
archive=""

require_value() {
  if [[ $# -lt 2 || -z "$2" || "$2" == --* ]]; then
    echo "error: $1 requires a value" >&2
    usage
    exit 2
  fi
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --destination)
      require_value "$@"
      destination="$2"
      shift 2
      ;;
    --target)
      require_value "$@"
      target="$2"
      shift 2
      ;;
    --archive)
      require_value "$@"
      archive="$2"
      shift 2
      ;;
    --metadata-dir)
      require_value "$@"
      metadata_dir="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      usage
      exit 2
      ;;
  esac
done

if [[ -z "$destination" ]]; then
  usage
  exit 2
fi

version_file="${metadata_dir}/version"
checksums_file="${metadata_dir}/checksums.txt"
if [[ ! -s "$version_file" || ! -s "$checksums_file" ]]; then
  echo "error: Codex version/checksum metadata is missing from ${metadata_dir}" >&2
  exit 1
fi

version="$(tr -d '[:space:]' < "$version_file")"
if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "error: invalid Codex version metadata: ${version}" >&2
  exit 1
fi

if [[ -z "$target" ]]; then
  case "$(uname -s)-$(uname -m)" in
    Linux-x86_64) target="linux-x86_64" ;;
    Linux-aarch64|Linux-arm64) target="linux-aarch64" ;;
    Darwin-arm64) target="darwin-aarch64" ;;
    Darwin-x86_64) target="darwin-x86_64" ;;
    *)
      echo "error: unsupported Codex installer platform: $(uname -s)-$(uname -m)" >&2
      exit 1
      ;;
  esac
fi

case "$target" in
  linux-x86_64) asset="codex-x86_64-unknown-linux-musl.tar.gz" ;;
  linux-aarch64) asset="codex-aarch64-unknown-linux-musl.tar.gz" ;;
  darwin-aarch64) asset="codex-aarch64-apple-darwin.tar.gz" ;;
  darwin-x86_64) asset="codex-x86_64-apple-darwin.tar.gz" ;;
  *)
    echo "error: unsupported Codex installer target: ${target}" >&2
    exit 1
    ;;
esac

expected="$(awk -v asset="$asset" '$2 == asset { print $1 }' "$checksums_file")"
if [[ ! "$expected" =~ ^[0-9a-f]{64}$ ]]; then
  echo "error: no valid checksum recorded for ${asset}" >&2
  exit 1
fi

temp_dir="$(mktemp -d)"
trap 'rm -rf "$temp_dir"' EXIT
download="${temp_dir}/${asset}"
if [[ -n "$archive" ]]; then
  cp "$archive" "$download"
else
  curl --proto '=https' --tlsv1.2 --fail --silent --show-error --location \
    --output "$download" \
    "https://github.com/openai/codex/releases/download/rust-v${version}/${asset}"
fi

if command -v sha256sum >/dev/null 2>&1; then
  actual="$(sha256sum "$download" | awk '{ print $1 }')"
else
  actual="$(shasum -a 256 "$download" | awk '{ print $1 }')"
fi
if [[ "$actual" != "$expected" ]]; then
  echo "error: checksum mismatch for ${asset}" >&2
  exit 1
fi

tar -xzf "$download" -C "$temp_dir"
binary_name="${asset%.tar.gz}"
if [[ ! -f "${temp_dir}/${binary_name}" ]]; then
  echo "error: archive does not contain ${binary_name}" >&2
  exit 1
fi

mkdir -p "$destination"
install -m 0755 "${temp_dir}/${binary_name}" "${destination}/codex"
"${destination}/codex" --version
