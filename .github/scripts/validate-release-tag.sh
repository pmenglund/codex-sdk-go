#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "usage: $0 REQUESTED_TAG [REPOSITORY]" >&2
}

if [[ $# -lt 1 || $# -gt 2 ]]; then
  usage
  exit 2
fi
requested_tag="$1"
repository="${2:-.}"
minimum_tag="v0.145.0"

canonical_component='(0|[1-9][0-9]*)'
if [[ ! "$requested_tag" =~ ^v${canonical_component}\.${canonical_component}\.${canonical_component}$ ]]; then
  echo "Requested tag must match vMAJOR.MINOR.PATCH." >&2
  exit 1
fi
if [[ "$(printf '%s\n%s\n' "$minimum_tag" "$requested_tag" | sort -V | head -n 1)" != "$minimum_tag" ]]; then
  echo "The typed-union migration requires ${minimum_tag} or newer, not ${requested_tag}." >&2
  exit 1
fi
if git -C "$repository" show-ref --verify --quiet "refs/tags/${requested_tag}"; then
  echo "Tag ${requested_tag} already exists and will not be moved." >&2
  exit 1
fi
latest_tag="$({
  git -C "$repository" tag --list | awk '/^v[0-9]+\.[0-9]+\.[0-9]+$/ { print }' | sort -V | tail -n 1
})"
if [[ -n "$latest_tag" && "$(printf '%s\n%s\n' "$latest_tag" "$requested_tag" | sort -V | tail -n 1)" != "$requested_tag" ]]; then
  echo "Requested tag ${requested_tag} must be greater than ${latest_tag}." >&2
  exit 1
fi
