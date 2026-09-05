#!/usr/bin/env bash
set -euo pipefail

script_dir="$(dirname "$(realpath "${BASH_SOURCE[0]}")")"
validator="${script_dir}/validate-release-tag.sh"
temp_dir="$(mktemp -d)"
trap 'rm -rf "$temp_dir"' EXIT

git -C "$temp_dir" init -q
git -C "$temp_dir" config user.name test
git -C "$temp_dir" config user.email test@example.com
touch "${temp_dir}/fixture"
git -C "$temp_dir" add fixture
git -C "$temp_dir" commit -qm fixture
git -C "$temp_dir" tag v0.145.0
git -C "$temp_dir" tag v0.146.0-rc.1

expect_failure() {
  if "$validator" "$1" "$temp_dir" >/dev/null 2>&1; then
    echo "expected release tag $1 to fail validation" >&2
    exit 1
  fi
}

expect_failure 0.146.0
expect_failure v00.146.0
expect_failure v0.0146.0
expect_failure v0.146.00
expect_failure v0.144.9
expect_failure v0.145.0
expect_failure v0.144.0
"$validator" v0.146.0 "$temp_dir"
