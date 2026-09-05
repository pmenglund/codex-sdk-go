#!/usr/bin/env bash
set -euo pipefail

if [[ $# -gt 1 ]]; then
  echo "usage: $0 [REPOSITORY]" >&2
  exit 2
fi
sdk_root="${1:-$(git rev-parse --show-toplevel)}"
version_file="${sdk_root}/.github/codex/version"
checksums_file="${sdk_root}/.github/codex/checksums.txt"
metadata_file="${sdk_root}/protocol/metadata_gen.go"

cli_version="$(tr -d '[:space:]' < "$version_file")"
generated_version="$(sed -n 's/^const GeneratedCodexVersion = "\([0-9.]*\)"/\1/p' "$metadata_file")"
if [[ ! "$cli_version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Pinned Codex CLI version is not semantic: ${cli_version}" >&2
  exit 1
fi
if [[ ! "$generated_version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Generated Codex protocol version is missing or invalid: ${generated_version}" >&2
  exit 1
fi
if [[ "${generated_version%.*}" != "${cli_version%.*}" ]]; then
  echo "Pinned Codex CLI ${cli_version} does not match generated protocol ${generated_version} at major/minor." >&2
  exit 1
fi

darwin_arm64=0
darwin_x86_64=0
linux_arm64=0
linux_x86_64=0
entry_count=0
seen_digests="|"
while IFS= read -r line || [[ -n "$line" ]]; do
  read -r digest asset extra <<< "$line"
  if [[ -n "${extra:-}" || ! "${digest:-}" =~ ^[0-9a-f]{64}$ ]]; then
    echo "Invalid Codex checksum entry: ${line}" >&2
    exit 1
  fi
  if [[ "$seen_digests" == *"|${digest}|"* ]]; then
    echo "Duplicate Codex checksum digest: ${digest}" >&2
    exit 1
  fi
  seen_digests+="${digest}|"
  case "${asset:-}" in
    codex-aarch64-apple-darwin.tar.gz) darwin_arm64=$((darwin_arm64 + 1)) ;;
    codex-x86_64-apple-darwin.tar.gz) darwin_x86_64=$((darwin_x86_64 + 1)) ;;
    codex-aarch64-unknown-linux-musl.tar.gz) linux_arm64=$((linux_arm64 + 1)) ;;
    codex-x86_64-unknown-linux-musl.tar.gz) linux_x86_64=$((linux_x86_64 + 1)) ;;
    *)
      echo "Unexpected Codex checksum asset: ${asset:-<missing>}" >&2
      exit 1
      ;;
  esac
  entry_count=$((entry_count + 1))
done < "$checksums_file"

if [[ "$entry_count" -ne 4 || "$darwin_arm64" -ne 1 || "$darwin_x86_64" -ne 1 || "$linux_arm64" -ne 1 || "$linux_x86_64" -ne 1 ]]; then
  echo "Codex checksums must contain each supported archive exactly once." >&2
  exit 1
fi

echo "Pinned Codex CLI ${cli_version} metadata matches generated protocol ${generated_version}."
