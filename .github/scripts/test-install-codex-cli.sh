#!/usr/bin/env bash
set -euo pipefail

script_dir="$(dirname "$(realpath "${BASH_SOURCE[0]}")")"
installer="${script_dir}/install-codex-cli.sh"
temp_dir="$(mktemp -d)"
trap 'rm -rf "$temp_dir"' EXIT

mkdir -p "${temp_dir}/fixture" "${temp_dir}/metadata" "${temp_dir}/bin"
printf '0.144.4\n' > "${temp_dir}/metadata/version"

targets=(linux-x86_64 linux-aarch64 darwin-aarch64 darwin-x86_64)
assets=(
  codex-x86_64-unknown-linux-musl.tar.gz
  codex-aarch64-unknown-linux-musl.tar.gz
  codex-aarch64-apple-darwin.tar.gz
  codex-x86_64-apple-darwin.tar.gz
)
: > "${temp_dir}/metadata/checksums.txt"
for index in "${!targets[@]}"; do
  target="${targets[$index]}"
  asset="${assets[$index]}"
  binary="${asset%.tar.gz}"
  printf '#!/usr/bin/env sh\necho codex-cli 0.144.4\n' > "${temp_dir}/fixture/${binary}"
  chmod +x "${temp_dir}/fixture/${binary}"
  tar -czf "${temp_dir}/${asset}" -C "${temp_dir}/fixture" "$binary"
  if command -v sha256sum >/dev/null 2>&1; then
    digest="$(sha256sum "${temp_dir}/${asset}" | awk '{ print $1 }')"
  else
    digest="$(shasum -a 256 "${temp_dir}/${asset}" | awk '{ print $1 }')"
  fi
  printf '%s  %s\n' "$digest" "$asset" >> "${temp_dir}/metadata/checksums.txt"
  "$installer" --destination "${temp_dir}/bin-${target}" --target "$target" --archive "${temp_dir}/${asset}" --metadata-dir "${temp_dir}/metadata"
  test -x "${temp_dir}/bin-${target}/codex"
done

asset="${assets[0]}"

if "$installer" --destination "${temp_dir}/bad-target" --target plan9-mips --archive "${temp_dir}/${asset}" --metadata-dir "${temp_dir}/metadata" >/dev/null 2>&1; then
  echo "expected unsupported target to fail" >&2
  exit 1
fi

mkdir -p "${temp_dir}/missing"
if "$installer" --destination "${temp_dir}/missing-bin" --target linux-x86_64 --archive "${temp_dir}/${asset}" --metadata-dir "${temp_dir}/missing" >/dev/null 2>&1; then
  echo "expected missing metadata to fail" >&2
  exit 1
fi

printf '%064d  %s\n' 0 "$asset" > "${temp_dir}/metadata/checksums.txt"
if "$installer" --destination "${temp_dir}/bad-checksum" --target linux-x86_64 --archive "${temp_dir}/${asset}" --metadata-dir "${temp_dir}/metadata" >/dev/null 2>&1; then
  echo "expected checksum mismatch to fail" >&2
  exit 1
fi

printf 'not-codex\n' > "${temp_dir}/fixture/wrong-binary"
tar -czf "${temp_dir}/missing-binary.tar.gz" -C "${temp_dir}/fixture" wrong-binary
if command -v sha256sum >/dev/null 2>&1; then
  digest="$(sha256sum "${temp_dir}/missing-binary.tar.gz" | awk '{ print $1 }')"
else
  digest="$(shasum -a 256 "${temp_dir}/missing-binary.tar.gz" | awk '{ print $1 }')"
fi
printf '%s  %s\n' "$digest" "$asset" > "${temp_dir}/metadata/checksums.txt"
if "$installer" --destination "${temp_dir}/missing-archive-binary" --target linux-x86_64 --archive "${temp_dir}/missing-binary.tar.gz" --metadata-dir "${temp_dir}/metadata" >/dev/null 2>&1; then
  echo "expected archive without the target binary to fail" >&2
  exit 1
fi

if output="$($installer --destination 2>&1)"; then
  echo "expected missing option value to fail" >&2
  exit 1
fi
if [[ "$output" != *"--destination requires a value"* ]]; then
  echo "expected a targeted missing-value error, got: $output" >&2
  exit 1
fi
