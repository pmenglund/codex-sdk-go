#!/usr/bin/env bash
set -euo pipefail

script_dir="$(dirname "$(realpath "${BASH_SOURCE[0]}")")"
state_resolver="${script_dir}/release-tag-state.sh"
temp_dir="$(mktemp -d)"
trap 'rm -rf "$temp_dir"' EXIT

git -C "$temp_dir" init -q
git -C "$temp_dir" config user.name test
git -C "$temp_dir" config user.email test@example.com
touch "${temp_dir}/fixture"
git -C "$temp_dir" add fixture
git -C "$temp_dir" commit -qm first
first_commit="$(git -C "$temp_dir" rev-parse HEAD)"

if [[ "$($state_resolver v0.145.0 "$first_commit" "$temp_dir")" != "absent" ]]; then
  echo "expected absent state for a new valid tag" >&2
  exit 1
fi

git -C "$temp_dir" tag -a v0.145.0 -m v0.145.0
if [[ "$($state_resolver v0.145.0 "$first_commit" "$temp_dir")" != "present" ]]; then
  echo "expected present state for an annotated tag at the expected commit" >&2
  exit 1
fi

printf 'second\n' > "${temp_dir}/fixture"
git -C "$temp_dir" commit -qam second
second_commit="$(git -C "$temp_dir" rev-parse HEAD)"
if "$state_resolver" v0.145.0 "$second_commit" "$temp_dir" >/dev/null 2>&1; then
  echo "expected a conflicting existing tag to fail" >&2
  exit 1
fi

git -C "$temp_dir" tag v0.146.0
if "$state_resolver" v0.146.0 "$second_commit" "$temp_dir" >/dev/null 2>&1; then
  echo "expected a lightweight release tag to fail" >&2
  exit 1
fi
