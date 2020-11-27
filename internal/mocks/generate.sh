#!/usr/bin/env bash
set -e

(
	type mockgen
	type sed
	type go
) 1>/dev/null

export PACKAGE="mocks"
export IN_PATH="int_test.go"
export OUT_PATH="MockTB_gen.go"

function main() {
	generateMock
	updateMock
	fmtFile
}

function generateMock() {
	mockgen -source "${IN_PATH}" -destination "${OUT_PATH}" -package "${PACKAGE}"
}

function updateMock() {
	local -a args=()

	if [[ "$(uname -s)" =~ Linux* ]]; then
		args+=("--in-place")
	fi

	# add testing.TB as dependency to MockTB
#	sed "${args[@]}" '/^import /a \"testing\"' "${OUT_PATH}"
	sed "${args[@]}" '/^type MockTB struct/a testing.TB' "${OUT_PATH}"
}

function fmtFile() {
	go fmt "${OUT_PATH}" 1>/dev/null
}

main "${@}"
