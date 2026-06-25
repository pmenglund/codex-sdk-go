#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage: update_codex_protocol.sh [--allow-dirty] [--no-commit]

Fetch upstream openai/codex tags, choose the highest stable rust-vMAJOR.MINOR.PATCH tag,
run go generate ./..., run go test ./..., and commit the resulting changes.

Options:
  --allow-dirty  Run even when the SDK worktree already has changes.
  --no-commit    Leave validated changes uncommitted.
  -h, --help     Show this help.
USAGE
}

allow_dirty=0
no_commit=0

while (($#)); do
  case "$1" in
    --allow-dirty)
      allow_dirty=1
      ;;
    --no-commit)
      no_commit=1
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
  shift
done

sdk_root="$(git rev-parse --show-toplevel)"
cd "$sdk_root"

if [[ ! -f "gen.go" || ! -d "internal/codegen" ]]; then
  echo "error: run this script from the codex-sdk-go repository" >&2
  exit 1
fi

if [[ "$allow_dirty" -eq 0 && -n "$(git status --porcelain)" ]]; then
  echo "error: SDK worktree has pre-existing changes; rerun with --allow-dirty only if they should be included" >&2
  git status --short
  exit 1
fi

strip_quotes() {
  local value="$1"
  value="${value#"${value%%[![:space:]]*}"}"
  value="${value%"${value##*[![:space:]]}"}"
  if [[ "$value" == \"*\" && "$value" == *\" ]]; then
    value="${value:1:${#value}-2}"
  elif [[ "$value" == \'*\' && "$value" == *\' ]]; then
    value="${value:1:${#value}-2}"
  fi
  printf '%s\n' "$value"
}

expand_home() {
  local value="$1"
  value="${value//\$\{HOME\}/$HOME}"
  value="${value//\$HOME/$HOME}"
  if [[ "$value" == "~/"* ]]; then
    value="$HOME/${value#~/}"
  fi
  printf '%s\n' "$value"
}

codex_root="${CODEX_REPO_ROOT:-}"
if [[ -z "$codex_root" && -f ".envrc" ]]; then
  envrc_line="$(grep -E '^[[:space:]]*(export[[:space:]]+)?CODEX_REPO_ROOT=' .envrc | tail -n 1 || true)"
  if [[ -n "$envrc_line" ]]; then
    envrc_value="${envrc_line#*=}"
    codex_root="$(expand_home "$(strip_quotes "$envrc_value")")"
  fi
fi

if [[ -z "$codex_root" && -d "../codex/.git" ]]; then
  codex_root="../codex"
fi

if [[ -z "$codex_root" ]]; then
  echo "error: CODEX_REPO_ROOT is not set, not present in .envrc, and ../codex was not found" >&2
  exit 1
fi

if ! git -C "$codex_root" rev-parse --show-toplevel >/dev/null 2>&1; then
  echo "error: CODEX_REPO_ROOT does not point to a git repository: $codex_root" >&2
  exit 1
fi

codex_root="$(cd "$codex_root" && pwd)"

echo "Fetching upstream Codex tags in $codex_root"
git -C "$codex_root" fetch --tags --force

latest_tag="$(
  git -C "$codex_root" tag --list 'rust-v*' | python3 -c '
import re
import sys

tags = []
for raw in sys.stdin:
    tag = raw.strip()
    match = re.fullmatch(r"rust-v([0-9]+)\.([0-9]+)\.([0-9]+)", tag)
    if match:
        tags.append((tuple(int(part) for part in match.groups()), tag))

if tags:
    tags.sort()
    print(tags[-1][1])
'
)"

if [[ -z "$latest_tag" ]]; then
  echo "error: no stable rust-vMAJOR.MINOR.PATCH tags found in $codex_root" >&2
  exit 1
fi

echo "Selected upstream Codex tag: $latest_tag"

echo "Running go generate ./..."
CODEX_REPO_ROOT="$codex_root" CODEX_REPO_REF="$latest_tag" go generate ./...

echo "Running go test ./..."
go test ./...

if git diff --quiet && git diff --cached --quiet; then
  echo "No changes to commit."
  exit 0
fi

if [[ "$no_commit" -eq 1 ]]; then
  echo "Validation passed; leaving changes uncommitted because --no-commit was set."
  git status --short
  exit 0
fi

git add .
git commit -m "Update Codex protocol from $latest_tag"
