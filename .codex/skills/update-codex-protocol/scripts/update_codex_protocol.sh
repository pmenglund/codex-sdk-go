#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage: update_codex_protocol.sh [--allow-dirty]

Fetch upstream openai/codex tags, choose the highest stable rust-vMAJOR.MINOR.PATCH
tag, generate twice to prove determinism, and run the complete local quality gate.
The script never commits, tags, stages, or pushes changes.
Staticcheck v0.7.0 and govulncheck v1.3.0 are run through Go's pinned module
tool support, so no separately installed analyzer binary is required.

Options:
  --allow-dirty  Run while preserving intentional pre-existing changes.
  -h, --help     Show this help.
USAGE
}

allow_dirty=0
while (($#)); do
  case "$1" in
    --allow-dirty) allow_dirty=1 ;;
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

for tool in git go cargo; do
  if ! command -v "$tool" >/dev/null 2>&1; then
    echo "error: required tool is not installed: $tool" >&2
    exit 1
  fi
done

if [[ ! -f "gen.go" || ! -d "internal/codegen" ]]; then
  echo "error: run this script from the codex-sdk-go repository" >&2
  exit 1
fi

if [[ "$allow_dirty" -eq 0 && -n "$(git status --porcelain)" ]]; then
  echo "error: SDK worktree has pre-existing changes; rerun with --allow-dirty only after reviewing them" >&2
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

latest_tag="$({
  git -C "$codex_root" tag --list 'rust-v*' | awk '
    /^rust-v[0-9]+\.[0-9]+\.[0-9]+$/ { print }
  ' | sort -V | tail -n 1
})"
if [[ -z "$latest_tag" ]]; then
  echo "error: no stable rust-vMAJOR.MINOR.PATCH tags found in $codex_root" >&2
  exit 1
fi
echo "Selected upstream Codex tag: $latest_tag"

generate() {
  CODEX_REPO_ROOT="$codex_root" CODEX_REPO_REF="$latest_tag" go generate ./...
}

echo "Running first deterministic generation"
generate
first_manifest="$(.github/scripts/content-manifest.sh protocol rpc internal/codegen)"
echo "Running second deterministic generation"
generate
second_manifest="$(.github/scripts/content-manifest.sh protocol rpc internal/codegen)"
if [[ "$first_manifest" != "$second_manifest" ]]; then
  echo "error: a second generation changed generated output" >&2
  exit 1
fi

unformatted="$(gofmt -l $(git ls-files '*.go'))"
if [[ -n "$unformatted" ]]; then
  echo "$unformatted"
  exit 1
fi
.github/scripts/validate-codex-metadata.sh
.github/scripts/test-validate-codex-metadata.sh
.github/scripts/test-install-codex-cli.sh
go vet ./...
go test ./...
go test -race ./...
go run honnef.co/go/tools/cmd/staticcheck@v0.7.0 ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.3.0 ./...
git diff --check

echo "Generation and all quality gates passed. Changes remain unstaged for review."
git status --short
