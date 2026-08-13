#!/usr/bin/env bash
set -euo pipefail

script_dir="$(dirname "$(realpath "${BASH_SOURCE[0]}")")"
resolver="${script_dir}/sdk-release-tag.sh"
temp_dir="$(mktemp -d)"
trap 'rm -rf "$temp_dir"' EXIT

mkdir -p "${temp_dir}/.github"

expect_failure() {
  if "$resolver" "$temp_dir" >/dev/null 2>&1; then
    echo "expected SDK version fixture to fail: $1" >&2
    exit 1
  fi
}

expect_failure missing

printf '0.147.0\n' > "${temp_dir}/.github/sdk-version"
resolved_tag="$($resolver "$temp_dir")"
if [[ "$resolved_tag" != "v0.147.0" ]]; then
  echo "expected v0.147.0, got ${resolved_tag}" >&2
  exit 1
fi

for invalid_version in 'v0.147.0' '0.147' '0.147.0-rc.1' '00.147.0' '0.0147.0' '0.147.00' ' 0.147.0' '0.147.0 ' $'0.147.0\n' $'0.147.0\n0.148.0'; do
  printf '%s\n' "$invalid_version" > "${temp_dir}/.github/sdk-version"
  expect_failure "$invalid_version"
done
