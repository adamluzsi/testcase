#!/usr/bin/env bash
set -e

(
	type mockgen
	type sed
	type go
) 1>/dev/null

export OUT_PATH="mocks.go"

function main() {
	generateMock
	updateMock
	fmtFile
}

function generateMock() {
	mockgen -source generate.go -destination "${OUT_PATH}" -package internal
}

function updateMock() {
	local -a args=()

	if [[ "$(uname -s)" =~ Linux* ]]; then
		args+=("--in-place")
	fi

  # add testing.TB as dependency to MockTB
	sed "${args[@]}" '/^import /a \"testing\"' "${OUT_PATH}"
	sed "${args[@]}" '/^type MockTB struct/a testing.TB' "${OUT_PATH}"
}

function fmtFile() {
	go fmt "${OUT_PATH}" 1>/dev/null
}

main "${@}"
