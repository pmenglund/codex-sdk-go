#!/usr/bin/env bash
set -euo pipefail

script_dir="$(dirname "$(realpath "${BASH_SOURCE[0]}")")"
validator="${script_dir}/validate-codex-metadata.sh"
temp_dir="$(mktemp -d)"
trap 'rm -rf "$temp_dir"' EXIT

write_valid_fixture() {
  mkdir -p "${temp_dir}/.github/codex" "${temp_dir}/protocol"
  printf '0.146.0\n' > "${temp_dir}/.github/codex/version"
  printf 'package protocol\n\nconst GeneratedCodexVersion = "0.146.1"\n' > "${temp_dir}/protocol/metadata_gen.go"
  {
    printf '%064d  codex-aarch64-apple-darwin.tar.gz\n' 1
    printf '%064d  codex-x86_64-apple-darwin.tar.gz\n' 2
    printf '%064d  codex-aarch64-unknown-linux-musl.tar.gz\n' 3
    printf '%064d  codex-x86_64-unknown-linux-musl.tar.gz\n' 4
  } > "${temp_dir}/.github/codex/checksums.txt"
}

expect_failure() {
  local name="$1"
  if "$validator" "$temp_dir" >/dev/null 2>&1; then
    echo "expected ${name} metadata to fail" >&2
    exit 1
  fi
}

write_valid_fixture
"$validator" "$temp_dir" >/dev/null

write_valid_fixture
printf '0.146\n' > "${temp_dir}/.github/codex/version"
expect_failure "malformed CLI version"

write_valid_fixture
printf 'package protocol\n\nconst GeneratedCodexVersion = "0.147.0"\n' > "${temp_dir}/protocol/metadata_gen.go"
expect_failure "mismatched protocol version"

write_valid_fixture
printf 'invalid  codex-aarch64-apple-darwin.tar.gz\n' > "${temp_dir}/.github/codex/checksums.txt"
expect_failure "invalid checksum"

write_valid_fixture
sed -i.bak '/codex-x86_64-unknown-linux-musl/d' "${temp_dir}/.github/codex/checksums.txt"
expect_failure "missing archive"

write_valid_fixture
printf '%064d  codex-unexpected.tar.gz\n' 5 >> "${temp_dir}/.github/codex/checksums.txt"
expect_failure "unexpected archive"

write_valid_fixture
printf '%064d  codex-aarch64-apple-darwin.tar.gz\n' 5 >> "${temp_dir}/.github/codex/checksums.txt"
expect_failure "duplicate archive"

write_valid_fixture
printf '%064d  codex-x86_64-unknown-linux-musl.tar.gz\n' 1 >> "${temp_dir}/.github/codex/checksums.txt"
sed -i.bak '/0004  codex-x86_64-unknown-linux-musl/d' "${temp_dir}/.github/codex/checksums.txt"
expect_failure "duplicate digest"
