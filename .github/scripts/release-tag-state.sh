#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "usage: $0 REQUESTED_TAG EXPECTED_COMMIT [REPOSITORY]" >&2
}

if [[ $# -lt 2 || $# -gt 3 ]]; then
  usage
  exit 2
fi

requested_tag="$1"
expected_commit="$2"
repository="${3:-.}"
validator="$(dirname "$(realpath "${BASH_SOURCE[0]}")")/validate-release-tag.sh"

expected_commit="$(git -C "$repository" rev-parse --verify "${expected_commit}^{commit}")"
if git -C "$repository" show-ref --verify --quiet "refs/tags/${requested_tag}"; then
  tag_object_type="$(git -C "$repository" cat-file -t "refs/tags/${requested_tag}")"
  if [[ "$tag_object_type" != "tag" ]]; then
    echo "Tag ${requested_tag} is not an annotated release tag." >&2
    exit 1
  fi
  tagged_commit="$(git -C "$repository" rev-parse --verify "refs/tags/${requested_tag}^{commit}")"
  if [[ "$tagged_commit" != "$expected_commit" ]]; then
    echo "Tag ${requested_tag} points to ${tagged_commit}, not expected commit ${expected_commit}." >&2
    exit 1
  fi
  echo present
  exit 0
fi

"$validator" "$requested_tag" "$repository"
echo absent
