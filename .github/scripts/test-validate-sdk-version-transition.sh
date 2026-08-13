#!/usr/bin/env bash
set -euo pipefail

script_dir="$(dirname "$(realpath "${BASH_SOURCE[0]}")")"
validator="${script_dir}/validate-sdk-version-transition.sh"
temp_dir="$(mktemp -d)"
trap 'rm -rf "$temp_dir"' EXIT

git -C "$temp_dir" init -q
git -C "$temp_dir" config user.name test
git -C "$temp_dir" config user.email test@example.com
mkdir -p "${temp_dir}/.github"
printf 'fixture\n' > "${temp_dir}/fixture"
git -C "$temp_dir" add fixture
git -C "$temp_dir" commit -qm initial
initial_commit="$(git -C "$temp_dir" rev-parse HEAD)"

printf '0.147.0\n' > "${temp_dir}/.github/sdk-version"
git -C "$temp_dir" add .github/sdk-version
git -C "$temp_dir" commit -qm first-version
first_version_commit="$(git -C "$temp_dir" rev-parse HEAD)"
"$validator" "$initial_commit" "$first_version_commit" "$temp_dir"

printf '0.148.0\n' > "${temp_dir}/.github/sdk-version"
git -C "$temp_dir" commit -qam next-version
next_version_commit="$(git -C "$temp_dir" rev-parse HEAD)"
if "$validator" "$first_version_commit" "$next_version_commit" "$temp_dir" >/dev/null 2>&1; then
  echo "expected an unreleased previous SDK version to block the next bump" >&2
  exit 1
fi

git -C "$temp_dir" tag -a v0.147.0 "$first_version_commit" -m v0.147.0
"$validator" "$first_version_commit" "$next_version_commit" "$temp_dir"

printf '0.146.0\n' > "${temp_dir}/.github/sdk-version"
git -C "$temp_dir" commit -qam decreasing-version
decreasing_commit="$(git -C "$temp_dir" rev-parse HEAD)"
if "$validator" "$next_version_commit" "$decreasing_commit" "$temp_dir" >/dev/null 2>&1; then
  echo "expected a decreasing SDK version to fail" >&2
  exit 1
fi

git -C "$temp_dir" checkout -q "$next_version_commit"
printf 'docs\n' >> "${temp_dir}/fixture"
git -C "$temp_dir" commit -qam unrelated-change
unrelated_commit="$(git -C "$temp_dir" rev-parse HEAD)"
"$validator" "$next_version_commit" "$unrelated_commit" "$temp_dir"
