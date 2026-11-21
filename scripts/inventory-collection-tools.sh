#!/usr/bin/env bash
# Inventory external collection tools on the machine where ARTEX workers execute.
# This is discovery only: it never installs, updates, configures, or runs a scan.
# Usage: bash scripts/inventory-collection-tools.sh [--strict]
set -uo pipefail

strict=0
case "${1:-}" in
  "") ;;
  --strict) strict=1 ;;
  -h|--help)
    echo "Usage: bash scripts/inventory-collection-tools.sh [--strict]"
    echo "Default: report present/missing tools and exit 0."
    echo "--strict: exit 1 if any candidate tool is missing."
    exit 0
    ;;
  *) echo "unknown option: $1" >&2; exit 2 ;;
esac

missing=0
printf 'tool\tstatus\tpath\tversion\n'

check_tool() {
  local name="$1"
  shift
  local path version
  if ! path=$(command -v "$name" 2>/dev/null); then
    printf '%s\tMISSING\t-\t-\n' "$name"
    missing=$((missing + 1))
    return
  fi
  # Version flags are tool-specific; output is captured locally and reduced to
  # one line. No target, URL, network scan, or mutation is supplied.
  version=$("$path" "$@" 2>&1)
  version=$(printf '%s\n' "$version" | sed -n '1p' | tr '\t\r\n' '   ')
  [ -n "$version" ] || version="version command returned no text"
  printf '%s\tPRESENT\t%s\t%s\n' "$name" "$path" "$version"
}

check_tool nmap --version
check_tool nuclei -version
check_tool httpx -version
check_tool katana -version
check_tool ffuf -V
check_tool dalfox version
check_tool bbot --version

if [ "$strict" -eq 1 ] && [ "$missing" -gt 0 ]; then
  printf 'Missing candidate tools: %s\n' "$missing" >&2
  exit 1
fi
