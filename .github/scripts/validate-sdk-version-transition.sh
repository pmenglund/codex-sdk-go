#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "usage: $0 BASE_COMMIT HEAD_COMMIT [REPOSITORY]" >&2
}

if [[ $# -lt 2 || $# -gt 3 ]]; then
  usage
  exit 2
fi

base_commit="$1"
head_commit="$2"
repository="${3:-.}"
version_path=".github/sdk-version"
canonical_component='(0|[1-9][0-9]*)'

base_commit="$(git -C "$repository" rev-parse --verify "${base_commit}^{commit}")"
head_commit="$(git -C "$repository" rev-parse --verify "${head_commit}^{commit}")"
head_version="$(git -C "$repository" show "${head_commit}:${version_path}")"
if [[ ! "$head_version" =~ ^${canonical_component}\.${canonical_component}\.${canonical_component}$ ]]; then
  echo "Head SDK version must match canonical MAJOR.MINOR.PATCH." >&2
  exit 1
fi

if ! base_version="$(git -C "$repository" show "${base_commit}:${version_path}" 2>/dev/null)"; then
  exit 0
fi
if [[ "$base_version" == "$head_version" ]]; then
  exit 0
fi
if [[ ! "$base_version" =~ ^${canonical_component}\.${canonical_component}\.${canonical_component}$ ]]; then
  echo "Base SDK version must match canonical MAJOR.MINOR.PATCH." >&2
  exit 1
fi
if [[ "$(printf '%s\n%s\n' "$base_version" "$head_version" | sort -V | tail -n 1)" != "$head_version" ]]; then
  echo "SDK version ${head_version} must be greater than ${base_version}." >&2
  exit 1
fi

previous_tag="v${base_version}"
if ! git -C "$repository" show-ref --verify --quiet "refs/tags/${previous_tag}"; then
  echo "Previous SDK version ${previous_tag} is not released; finish it before merging another version bump." >&2
  exit 1
fi
if [[ "$(git -C "$repository" cat-file -t "refs/tags/${previous_tag}")" != "tag" ]]; then
  echo "Previous SDK version ${previous_tag} is not an annotated release tag." >&2
  exit 1
fi
previous_commit="$(git -C "$repository" rev-parse --verify "refs/tags/${previous_tag}^{commit}")"
if ! git -C "$repository" merge-base --is-ancestor "$previous_commit" "$base_commit"; then
  echo "Previous SDK release ${previous_tag} is not in the base branch history." >&2
  exit 1
fi
