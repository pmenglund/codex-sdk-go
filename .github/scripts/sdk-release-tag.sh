#!/usr/bin/env bash
set -euo pipefail

repository="${1:-.}"
version_file="${repository}/.github/sdk-version"

if [[ ! -f "$version_file" ]]; then
  echo "Missing SDK version file: ${version_file}" >&2
  exit 1
fi

sdk_version="$(<"$version_file")"
canonical_component='(0|[1-9][0-9]*)'
if [[ ! "$sdk_version" =~ ^${canonical_component}\.${canonical_component}\.${canonical_component}$ ]] || [[ "$(wc -l < "$version_file" | tr -d '[:space:]')" != "1" ]]; then
  echo "SDK version must match MAJOR.MINOR.PATCH exactly." >&2
  exit 1
fi

printf 'v%s\n' "$sdk_version"
