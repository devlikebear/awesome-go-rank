#!/usr/bin/env bash

set -euo pipefail

workflow="${1:-.github/workflows/security.yml}"
go_mod="${2:-go.mod}"
failures=0

fail() {
  printf 'security workflow policy: %s\n' "$1" >&2
  failures=$((failures + 1))
}

shopt -s nullglob
workflow_files=("$(dirname "$workflow")"/*.yml "$(dirname "$workflow")"/*.yaml)
for workflow_file in "${workflow_files[@]}"; do
  while IFS= read -r uses_line; do
    action_ref="${uses_line#*uses: }"
    action_ref="${action_ref%% #*}"
    if [[ ! "$action_ref" =~ @[0-9a-f]{40}$ ]]; then
      fail "action must be pinned to a full commit SHA in $workflow_file: $action_ref"
    fi
  done < <(grep -E '^[[:space:]]*(-[[:space:]]+)?uses:' "$workflow_file")
done

if grep -Eq 'go install .+@(latest|master|main)([[:space:]]|$)' "$workflow"; then
  fail "Go security tools must use an exact module version"
fi

if grep -q -- '-no-fail' "$workflow"; then
  fail "Gosec findings must fail the scan step"
fi

if ! grep -Eq "exit-code:[[:space:]]*['\"]?1['\"]?" "$workflow"; then
  fail "Trivy must return a non-zero exit code for configured findings"
fi

if grep -Eq '^[[:space:]]*continue-on-error:[[:space:]]*true' "$workflow"; then
  fail "security scanner failures must not be ignored"
fi

always_uploads="$(grep -Ec '^[[:space:]]*if:[[:space:]]*always\(\)' "$workflow" || true)"
if (( always_uploads < 2 )); then
  fail "both SARIF uploads must run even when their scanner reports findings"
fi

go_version="$(awk '$1 == "go" { print $2; exit }' "$go_mod")"
IFS=. read -r go_major go_minor go_patch <<< "$go_version"
go_patch="${go_patch:-0}"
if (( go_major < 1 || (go_major == 1 && go_minor < 25) || (go_major == 1 && go_minor == 25 && go_patch < 12) )); then
  fail "Go must be at least 1.25.12 to include the required standard-library security fixes"
fi

if (( failures > 0 )); then
  exit 1
fi

printf 'security workflow policy: ok\n'
