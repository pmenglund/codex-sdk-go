#!/usr/bin/env bash
set -euo pipefail

script_dir="$(dirname "$(realpath "${BASH_SOURCE[0]}")")"
manifest="${script_dir}/content-manifest.sh"
temp_dir="$(mktemp -d)"
trap 'rm -rf "$temp_dir"' EXIT

git -C "$temp_dir" init -q
git -C "$temp_dir" config user.name test
git -C "$temp_dir" config user.email test@example.com
mkdir -p "${temp_dir}/generated"
printf 'tracked\n' > "${temp_dir}/generated/tracked.go"
git -C "$temp_dir" add generated/tracked.go
git -C "$temp_dir" commit -qm fixture
printf 'first\n' > "${temp_dir}/generated/new.go"

first="$(cd "$temp_dir" && "$manifest" generated)"
printf 'second\n' > "${temp_dir}/generated/new.go"
second="$(cd "$temp_dir" && "$manifest" generated)"
if [[ "$first" == "$second" ]]; then
  echo "content manifest ignored an untracked generated-file change" >&2
  exit 1
fi
