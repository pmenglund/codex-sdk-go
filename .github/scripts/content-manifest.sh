#!/usr/bin/env bash
set -euo pipefail

if [[ $# -eq 0 ]]; then
  echo "usage: $0 PATH..." >&2
  exit 2
fi

git ls-files --cached --others --exclude-standard -- "$@" | LC_ALL=C sort | while IFS= read -r file; do
  [[ -f "$file" ]] || continue
  if command -v sha256sum >/dev/null 2>&1; then
    digest="$(sha256sum "$file" | awk '{ print $1 }')"
  else
    digest="$(shasum -a 256 "$file" | awk '{ print $1 }')"
  fi
  printf '%s  %s\n' "$digest" "$file"
done
