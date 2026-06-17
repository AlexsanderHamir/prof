#!/usr/bin/env bash
# Run the full test suite and print statement coverage using the Go toolchain.
#
# Uses -coverpkg=./... so integration tests in tests/ count toward every package
# they exercise (see https://go.dev/blog/integration-test-coverage).
#
# Usage (from repo root):
#   ./scripts/test-cover.sh
#   ./scripts/test-cover.sh -html
#   ./scripts/test-cover.sh -html -o /tmp/coverage.out
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

PROFILE="${COVERAGE_PROFILE:-coverage.out}"
HTML="${COVERAGE_HTML:-coverage.html}"
WRITE_HTML=0

while [[ $# -gt 0 ]]; do
	case "$1" in
	-html)
		WRITE_HTML=1
		shift
		;;
	-o)
		PROFILE="$2"
		shift 2
		;;
	-h|--help)
		echo "Usage: $0 [-html] [-o coverage.out]"
		exit 0
		;;
	*)
		echo "unknown argument: $1" >&2
		exit 2
		;;
	esac
done

echo "Running tests with coverage (-coverpkg=./...)..."
mapfile -t PKGS < <(go list ./... | grep -v '/scripts/')
go test "${PKGS[@]}" \
	-count=1 \
	-covermode=atomic \
	-coverpkg=./... \
	-coverprofile="$PROFILE"

echo ""
echo "=== Coverage summary ==="
COVER_MAIN="$ROOT/scripts/coverreport/main.go"
if [[ "$WRITE_HTML" -eq 1 ]]; then
	go run "$COVER_MAIN" -html "$HTML" "$PROFILE"
else
	go run "$COVER_MAIN" "$PROFILE"
fi

echo ""
echo "Profile: $ROOT/$PROFILE"
if [[ "$WRITE_HTML" -eq 1 ]]; then
	echo "HTML:    $ROOT/$HTML"
fi
echo "Tip: go tool cover -func=$PROFILE"
